package io

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/xops-infra/multi-k8s-client/pkg/model"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type k8sClient struct {
	clusterInfo model.ClusterInfo
	clientSet   *kubernetes.Clientset
	dynamic     dynamic.Interface
}

// kubePath or kubeConfig(base64 kubeconfig), kubePath > kubeConfig if both exist
func NewK8SClient(cfg model.Cluster) (model.K8SIO, error) {
	var config *rest.Config
	if cfg.KubePath != nil {
		var _kubePath string
		if strings.HasPrefix(*cfg.KubePath, "~/") {
			_kubePath = strings.Replace(*cfg.KubePath, "~/", homedir.HomeDir()+"/", 1)
		}
		_config, err := clientcmd.BuildConfigFromFlags("", _kubePath)
		if err != nil {
			return nil, err
		}
		config = _config
	} else if cfg.KubeConfig != nil {
		decodeBase64Config, err := base64.StdEncoding.DecodeString(*cfg.KubeConfig)
		if err != nil {
			return nil, err
		}
		_config, err := clientcmd.NewClientConfigFromBytes(decodeBase64Config)
		if err != nil {
			return nil, err
		}
		config, err = _config.ClientConfig()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("need kubePath or kubeConfig")
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &k8sClient{
		clientSet: clientset,
		dynamic:   dynamicClient,
		clusterInfo: model.ClusterInfo{
			Name:  cfg.Name,
			Alias: cfg.Alias,
		},
	}, nil
}

func (c *k8sClient) GetClusterInfo() model.ClusterInfo {
	return c.clusterInfo
}

// for dynamic
// getGVR :- gets GroupVersionResource for dynamic client
func GetGVR(group, version, resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
}
