//go:generate mockery -name=Sessioner
//go:generate mockery -name=Clusterer
package cluster

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/charter-se/structured/errors"
	"github.com/charter-se/structured/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/version"
	podutil "k8s.io/kubernetes/pkg/api/pod"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
)

type Sessioner interface {
	Clusterer
	Releaser
}
type Clusterer interface {
	Init() error
	GetKubeConfig() string
	SetKubeConfig(c string)
	GetKubeContext() string
	SetKubeContext(c string)
}

type Session struct {
	Helm        helm.Interface
	Tiller      *kube.Tunnel
	Clientset   *internalclientset.Clientset
	kubeConfig  string
	kubeContext string
}

//NewSession returns a *Session with kubernetes connections established
func NewSession(kubeContext string, kubeConfig string) *Session {
	s := &Session{}
	s.kubeConfig = fullPath(kubeConfig)
	s.kubeContext = kubeContext
	return s
}
func (s *Session) GetKubeConfig() string {
	return s.kubeConfig
}
func (s *Session) SetKubeConfig(c string) {
	s.kubeConfig = c
}
func (s *Session) GetKubeContext() string {
	return s.kubeContext
}
func (s *Session) SetKubeContext(c string) {
	s.kubeContext = c
}

//Init establishes connextions to the cluster
func (s *Session) Init() error {

	tillerNamespace := os.Getenv("TILLER_NAMESPACE")
	if tillerNamespace == "" {
		tillerNamespace = "kube-system"
	}
	err := s.connect(tillerNamespace)
	if err != nil {
		return errors.Wrap(err, "connection to kubernetes failed")
	}

	err = s.Helm.PingTiller()
	if err != nil {
		return errors.Wrap(err, "helm.PingTiller() failed")
	}

	tillerVersion, err := s.Helm.GetVersion()
	if err != nil {
		return errors.Wrap(err, "failed to get Tiller version")
	}

	compatible := version.IsCompatible(version.Version, tillerVersion.Version.SemVer)
	log.WithFields(log.Fields{
		"tillerVersion":          tillerVersion.Version.SemVer,
		"clientServerCompatible": compatible,
		"Host":                   fmt.Sprintf(":%v", s.Tiller.Local),
	}).Debug("Connected to Tiller")
	if !compatible {
		return errors.WithFields(errors.Fields{
			"tillerVersion": tillerVersion.Version.SemVer,
			"helmVersion":   version.Version,
			"Host":          fmt.Sprintf(":%v", s.Tiller.Local),
		}).New("incompatible version numbers")
	}
	return nil
}

//connect builds connections for all supported APIs
func (s *Session) connect(namespace string) error {
	config, err := kube.GetConfig(s.GetKubeContext(), s.GetKubeConfig()).ClientConfig()
	if err != nil {
		return errors.WithFields(errors.Fields{
			"KubeConfig":  s.GetKubeConfig(),
			"kubeContext": s.GetKubeContext(),
		}).Wrap(err, "could not get kubernetes config for context")
	}
	s.Clientset, err = internalclientset.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "could not get kubernetes client")
	}
	podName, err := getTillerPodName(s.Clientset.Core(), namespace)
	if err != nil {
		return errors.Wrap(err, "could not get Tiller pod name")
	}
	const tillerPort = 44134
	s.Tiller = kube.NewTunnel(s.Clientset.Core().RESTClient(), config, namespace, podName, tillerPort)
	err = s.Tiller.ForwardPort()
	if err != nil {
		return errors.Wrap(err, "could not get Tiller tunnel")
	}

	s.Helm = helm.NewClient(helm.Host(fmt.Sprintf(":%v", s.Tiller.Local)), helm.ConnectTimeout(5))
	clientVersion, err := s.Helm.GetVersion()
	if err != nil {
		return errors.Wrap(err, "failed to get helm client version")
	}
	log.WithFields(log.Fields{
		"Version": clientVersion.Version,
	}).Debug("Helm client")

	return nil
}

func getTillerPodName(client internalversion.PodsGetter, namespace string) (string, error) {
	// TODO use a const for labels
	selector := labels.Set{"app": "helm", "name": "tiller"}.AsSelector()
	pod, err := getFirstRunningPod(client, namespace, selector)
	if err != nil {
		return "", errors.Wrap(err, "failed to get first running pod")
	}
	return pod.ObjectMeta.GetName(), nil
}

func getFirstRunningPod(client internalversion.PodsGetter, namespace string, selector labels.Selector) (*core.Pod, error) {
	options := metav1.ListOptions{LabelSelector: selector.String()}
	pods, err := client.Pods(namespace).List(options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) < 1 {
		return nil, errors.New("could not find tiller")
	}
	for _, p := range pods.Items {
		if podutil.IsPodReady(&p) {
			return &p, nil
		}
	}
	return nil, errors.New("could not find a ready tiller pod")
}

func fullPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		usr, _ := user.Current()
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}