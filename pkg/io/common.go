package io

import (
	"strings"

	"github.com/xops-infra/multi-k8s-client/pkg/model"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type k8sClient struct {
	clientSet *kubernetes.Clientset
	dynamic   dynamic.Interface
}

func NewK8SClient(kubeConfig string) (model.K8SIO, error) {
	if strings.HasPrefix(kubeConfig, "~/") {
		kubeConfig = strings.Replace(kubeConfig, "~/", homedir.HomeDir()+"/", 1)
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return &k8sClient{
		clientSet: clientset,
		dynamic:   dynamicClient,
	}, nil
}

// for dynamic
// getGVR :- gets GroupVersionResource for dynamic client
func GetGVR(group, version, resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
}
