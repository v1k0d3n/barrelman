package cluster

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
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

func NewSession() *Session {

	s := &Session{}

	kubeContext := ""

	tillerNamespace := os.Getenv("TILLER_NAMESPACE")
	if tillerNamespace == "" {
		tillerNamespace = "kube-system"
	}
	err := s.connect(kubeContext, tillerNamespace)
	if err != nil {
		log.WithField("error", err).Fatalf("Could not set up connection to helm")
		return &Session{}
	}

	log.WithField("host", tillerHost).Info("Tiller host")

	s.Helm = helm.NewClient(helm.Host(tillerHost), helm.ConnectTimeout(5))
	log.WithField("client", s.Helm).Debug("Helm client")

	err = s.Helm.PingTiller()
	if err != nil {
		log.WithField("error", err).Fatalf("helm.PingTiller() failed")
	}

	tillerVersion, err := s.Helm.GetVersion()
	if err != nil {
		log.WithField("error", err).Fatalf("failed to get Tiller version")
		return nil
	}

	compatible := version.IsCompatible(version.Version, tillerVersion.Version.SemVer)
	log.WithFields(log.Fields{"tillerVersion": tillerVersion.Version.SemVer, "clientServerCompatible": compatible}).Info("Connected to Tiller")
	if !compatible {
		log.WithFields(log.Fields{
			"Helm Version":   version.Version,
			"Tiller Version": tillerVersion.Version.SemVer,
		}).Warnf("incompatible version numbers")
	}
	return s
}

//connect builds connections for all supported APIs
func (s *Session) connect(context string, namespace string) error {
	config, err := kube.GetConfig(context, "").ClientConfig()
	if err != nil {
		return fmt.Errorf("could not get kubernetes config for context '%s': %s", context, err)
	}
	s.Clientset, err = internalclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("could not get kubernetes client: %s", err)
	}
	podName, err := getTillerPodName(s.Clientset.Core(), namespace)
	if err != nil {
		return fmt.Errorf("could not get Tiller pod name: %s", err)
	}
	const tillerPort = 44134
	s.Tiller = kube.NewTunnel(s.Clientset.Core().RESTClient(), config, namespace, podName, tillerPort)
	err = s.Tiller.ForwardPort()
	if err != nil {
		return fmt.Errorf("could not get Tiller tunnel: %s", err)
	}

	return nil
}

func getTillerPodName(client internalversion.PodsGetter, namespace string) (string, error) {
	// TODO use a const for labels
	selector := labels.Set{"app": "helm", "name": "tiller"}.AsSelector()
	pod, err := getFirstRunningPod(client, namespace, selector)
	if err != nil {
		return "", err
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
		return nil, fmt.Errorf("could not find tiller")
	}
	for _, p := range pods.Items {
		if podutil.IsPodReady(&p) {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("could not find a ready tiller pod")
}
