//go:generate mockery -name=Sessioner
//go:generate mockery -name=Clusterer
//go:generate mockery -name=Transactioner

package cluster

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/tlsutil"
	"k8s.io/helm/pkg/version"

	"github.com/cirrocloud/structured/errors"
	"github.com/cirrocloud/structured/log"
)

type Sessioner interface {
	Clusterer
	Releaser
	Versioner
	NewTransactioner
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
	Tunnel      *kube.Tunnel
	Clientset   kubernetes.Interface
	kubeConfig  string
	kubeContext string
	settings    helm_env.EnvSettings
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

//Init establishes connections to the cluster
func (s *Session) Init() error {

	tillerNamespace := os.Getenv("TILLER_NAMESPACE")
	if tillerNamespace == "" {
		tillerNamespace = "kube-system"
	}

	if err := s.connect(tillerNamespace); err != nil {
		return errors.Wrap(err, "connection to kubernetes failed")
	}

	if err := s.connectionHealthCheck(); err != nil {
		return errors.Wrap(err, "kubernetes connection failed health check")
	}

	tillerVersion, err := s.Helm.GetVersion()
	if err != nil {
		return errors.Wrap(err, "failed to get Tiller version")
	}

	compatible := version.IsCompatible(version.Version, tillerVersion.Version.SemVer)
	log.WithFields(log.Fields{
		"tillerVersion":          tillerVersion.Version.SemVer,
		"clientServerCompatible": compatible,
		"Host":                   fmt.Sprintf(":%v", s.Tunnel.Local),
	}).Debug("Connected to Tiller")
	if !compatible {
		return errors.WithFields(errors.Fields{
			"tillerVersion": tillerVersion.Version.SemVer,
			"helmVersion":   version.Version,
			"Host":          fmt.Sprintf(":%v", s.Tunnel.Local),
		}).New("incompatible version numbers")
	}
	return nil
}

func (s *Session) ShortReport() map[string]interface{} {
	return map[string]interface{}{
		"KubeContext": s.kubeContext,
		"KubeConfig":  s.kubeConfig,
	}
}

func (s *Session) DetailedReport() map[string]interface{} {
	return map[string]interface{}{
		"KubeContext": s.kubeContext,
		"KubeConfig":  s.kubeConfig,
	}
}

//connect builds connections for all supported APIs
func (s *Session) connect(namespace string) error {

	//static analysis suggests use of these
	kubeContext := s.GetKubeContext()
	kubeConfig := s.GetKubeConfig()

	config, err := kube.GetConfig(kubeContext, kubeConfig).ClientConfig()
	if err != nil {
		return errors.WithFields(errors.Fields{
			"KubeConfig":  kubeConfig,
			"kubeContext": kubeContext,
		}).Wrap(err, "could not get kubernetes config for context")
	}

	// pupulate session TLS configuration from environment
	s.setTLSConfig()

	s.Clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "could not get kubernetes client")
	}

	podName, err := getTillerPodName(s.Clientset, namespace)
	if err != nil {
		return errors.Wrap(err, "could not get Tiller pod name")
	}
	const tillerPort = 44134
	s.Tunnel = kube.NewTunnel(s.Clientset.CoreV1().RESTClient(), config, namespace, podName, tillerPort)
	err = s.Tunnel.ForwardPort()
	if err != nil {
		return errors.Wrap(err, "could not get Tiller tunnel")
	}

	options := []helm.Option{
		helm.Host(fmt.Sprintf("127.0.0.1:%v", s.Tunnel.Local)),
		helm.ConnectTimeout(5),
	}

	if s.settings.TLSVerify || s.settings.TLSEnable {
		log.WithFields(log.Fields{
			"Key":  s.settings.TLSKeyFile,
			"Cert": s.settings.TLSCertFile,
			"CA":   s.settings.TLSCaCertFile,
		}).Debug("Using TLS for Tiller")
		tlsopts := tlsutil.Options{
			ServerName:         s.settings.TLSServerName,
			KeyFile:            s.settings.TLSKeyFile,
			CertFile:           s.settings.TLSCertFile,
			InsecureSkipVerify: true,
		}
		if s.settings.TLSVerify {
			tlsopts.CaCertFile = s.settings.TLSCaCertFile
			tlsopts.InsecureSkipVerify = false
		}
		tlscfg, err := tlsutil.ClientConfig(tlsopts)
		if err != nil {
			return errors.Wrap(err, "could not get TLS Client Configuration")
		}
		options = append(options, helm.WithTLS(tlscfg))
	}

	s.Helm = helm.NewClient(options...)

	return nil
}

func (s *Session) setTLSConfig() {
	// Setup TLS as done in helm/cmd/helm

	if os.Getenv("HELM_TLS_ENABLE") == "true" || os.Getenv("HELM_TLS_ENABLE") == "1" {
		s.settings.TLSEnable = true
	}

	if os.Getenv("HELM_HOME") == "" {
		os.Setenv("HELM_HOME", helm_env.DefaultHelmHome)
	}

	log.Debugf("TLSCaCert: %s", helm_env.DefaultTLSCaCert)

	log.Debugf("using HELM_HOME home %s", os.Getenv("HELM_HOME"))

	if os.Getenv("HELM_TLS_CA_CERT") != "" {
		s.settings.TLSCaCertFile = os.Getenv("HELM_TLS_CA_CERT")
	} else {
		s.settings.TLSCaCertFile = os.ExpandEnv(helm_env.DefaultTLSCaCert)
	}

	if os.Getenv("HELM_TLS_CERT") != "" {
		s.settings.TLSCertFile = os.Getenv("HELM_TLS_CERT")
	} else {
		s.settings.TLSCertFile = os.ExpandEnv(helm_env.DefaultTLSCert)
	}

	if os.Getenv("HELM_TLS_KEY") != "" {
		s.settings.TLSKeyFile = os.Getenv("HELM_TLS_KEY")
	} else {
		s.settings.TLSKeyFile = os.ExpandEnv(helm_env.DefaultTLSKeyFile)
	}

	if os.Getenv("HELM_TLS_VERIFY") == "true" || os.Getenv("HELM_TLS_VERIFY") == "1" {
		s.settings.TLSVerify = true
	}
}

// connectionHealthCheck verifies connectivity and version compatability
func (s *Session) connectionHealthCheck() error {

	// NOTE: Its possible Helm client may not be connected until this point
	err := s.Helm.PingTiller()
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
		"Host":                   fmt.Sprintf(":%v", s.Tunnel.Local),
	}).Debug("Connected to Tiller")

	if !compatible {
		return errors.WithFields(errors.Fields{
			"tillerVersion": tillerVersion.Version.SemVer,
			"helmVersion":   version.Version,
			"Host":          fmt.Sprintf(":%v", s.Tunnel.Local),
		}).New("incompatible version numbers")
	}

	return nil
}

func getTillerPodName(client kubernetes.Interface, namespace string) (string, error) {
	// TODO use a const for labels
	selector := labels.Set{"app": "helm", "name": "tiller"}.AsSelector()
	pod, err := getFirstRunningPod(client, namespace, selector)
	if err != nil {
		return "", errors.Wrap(err, "failed to get first running pod")
	}
	return pod.ObjectMeta.GetName(), nil
}

func getFirstRunningPod(client kubernetes.Interface, namespace string, selector labels.Selector) (*core.Pod, error) {
	options := metav1.ListOptions{LabelSelector: selector.String()}
	pods, err := client.CoreV1().Pods(namespace).List(options)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) < 1 {
		return nil, errors.New("could not find tiller")
	}
	for _, p := range pods.Items {
		for _, v := range p.Status.Conditions {
			if v.Type == core.PodReady {
				return &p, nil
			}
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
