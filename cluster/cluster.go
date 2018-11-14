package cluster

import (
	"fmt"
	"os"

	"github.com/charter-se/barrelman/errors"
	"github.com/charter-se/barrelman/log"
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

type Session struct {
	Helm      helm.Interface
	Tiller    *kube.Tunnel
	Clientset *internalclientset.Clientset
}

func NewSession(kubeConfig string) (*Session, error) {

	s := &Session{}

	tillerNamespace := os.Getenv("TILLER_NAMESPACE")
	if tillerNamespace == "" {
		tillerNamespace = "kube-system"
	}
	err := s.connect(kubeConfig, tillerNamespace)
	if err != nil {
		return &Session{}, errors.Wrap(err, "connection to kubernetes failed")
	}

	err = s.Helm.PingTiller()
	if err != nil {
		return &Session{}, errors.Wrap(err, "helm.PingTiller() failed")
	}

	tillerVersion, err := s.Helm.GetVersion()
	if err != nil {
		return &Session{}, errors.Wrap(err, "failed to get Tiller version")
	}

	compatible := version.IsCompatible(version.Version, tillerVersion.Version.SemVer)
	log.WithFields(log.Fields{
		"tillerVersion":          tillerVersion.Version.SemVer,
		"clientServerCompatible": compatible,
		"Host":                   fmt.Sprintf(":%v", s.Tiller.Local),
	}).Info("Connected to Tiller")
	if !compatible {
		return &Session{}, errors.WithFields(errors.Fields{
			"tillerVersion":          tillerVersion.Version.SemVer,
			"clientServerCompatible": compatible,
			"Host":                   fmt.Sprintf(":%v", s.Tiller.Local),
		}).New("incompatible version numbers")
	}
	return s, nil
}

//connect builds connections for all supported APIs
func (s *Session) connect(kubeConfig string, namespace string) error {
	config, err := kube.GetConfig("", kubeConfig).ClientConfig()
	if err != nil {
		return errors.WithFields(errors.Fields{"KubeConfig": kubeConfig}).Wrap(err, "could not get kubernetes config for context")
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
	log.WithField("client", s.Helm).Debug("Helm client")

	return nil
}

func getTillerPodName(client internalversion.PodsGetter, namespace string) (string, error) {
	// TODO use a const for labels
	selector := labels.Set{"app": "helm", "name": "tiller"}.AsSelector()
	pod, err := getFirstRunningPod(client, namespace, selector)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get first running pod")
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
