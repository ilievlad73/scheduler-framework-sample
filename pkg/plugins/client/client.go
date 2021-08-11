package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Connect() (*kubernetes.Clientset, *rest.Config, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return clientset, config, nil
}
