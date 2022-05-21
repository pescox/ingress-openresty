package kube

import (
	"flag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClient() *kubernetes.Clientset {

	config := GetConfig()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset
}

func GetConfig() *rest.Config {
	kubeconfig := flag.String("kubeconfig", clientcmd.NewDefaultPathOptions().GetDefaultFilename(), "kubeconfig file path")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
	}
	return config
}
