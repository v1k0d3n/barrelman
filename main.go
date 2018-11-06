package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/charter-se/barrelman/cluster"
	"github.com/charter-se/barrelman/yamlpack"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/helm/pkg/kube"
)

func main() {
	fmt.Printf("Barrelman Engage!\n")
	kubeconfig := fmt.Sprintf("%v/.kube/config", userHomeDir())
	// send logs to stderr so we can use 'kubectl logs'
	config, err := getConfig(kubeconfig)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to load client config")
		return
	}

	// build the Kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create kubernetes client")
		return
	}

	// list pods
	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to retrieve pods")
		return
	}

	for _, p := range pods.Items {
		log.WithFields(log.Fields{
			"namespace": p.Namespace,
			"name":      p.Name,
		}).Info("Found pods")
	}

	_ = cluster.NewSession()
	yp := yamlpack.New()
	if err := yp.Import("testdata/armada-osh.yaml"); err != nil {
		fmt.Printf("Error importing \"this\": %v\n", err)
	}

	// k.PrintMeta()

	// for name, f := range yp.Files {
	// 	fmt.Println("_________________________")
	// 	fmt.Printf("This: %v\n", name)
	// 	for _, k := range f {
	// 		fmt.Printf("Schema: %v\n", k.Viper.Get("schema"))
	// 		fmt.Printf("Metdata Name: %v\n", k.Viper.Get("metadata.name"))
	// 		fmt.Printf("Metdata Schema: %v\n", k.Viper.Get("metadata.schema"))
	// 		y, err := k.Yaml()
	// 		if err != nil {
	// 			fmt.Printf("Failed to marshal data: %v\n", err)
	// 			return
	// 		}
	// 		fmt.Printf("Data:\n%v\n", string(y))
	// 	}
	// 	fmt.Printf("\n\n")
	// }
	// fmt.Printf("Yaml Sections:\n")
	// for _, s := range yp.ListYamls() {
	// 	fmt.Printf("\t%v\n", s)
	// }

}

func getConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	return rest.InClusterConfig()
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
