package model

import (
	podV1 "k8s.io/api/core/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Filter struct {
	LabelSelector *string `json:"label_selector"` // key1=value1,key2=value2
	FieldSelector *string `json:"field_selector"` // key1=value1,key2=value2 "metadata.name=flink-session,metadata.namespace=default
}

func (s *Filter) ToOptions() metav1.ListOptions {
	var opts metav1.ListOptions
	if s.LabelSelector != nil {
		opts.LabelSelector = *s.LabelSelector
	}
	if s.FieldSelector != nil {
		opts.FieldSelector = *s.FieldSelector
	}
	return opts
}

type K8SIO interface {
	// POD
	PodList(namespace string) (*podV1.PodList, error)
	PodGet(namespace, name string) (*podV1.Pod, error)

	// RBAC
	RbacList(namespace string) (*rbacV1.RoleList, error)

	// CRD Flink
	CrdFlinkDeploymentList(Filter) (*unstructured.UnstructuredList, error)
	CrdFlinkDeploymentApply(namespace string, yaml map[string]any) (any, error)
	CrdFlinkDeploymentDelete(namespace, name string) error
	CrdFlinkSessionJobList(Filter) (*unstructured.UnstructuredList, error)
	CrdFlinkSessionJobSubmit(namespace string, yaml map[string]any) (any, error) // for flink session cluster, can't be used for application cluster
	CrdFlinkSessionJobDelete(namespace, name string) error
}

type K8SContract interface {
	GetK8SCluster() ([]string, error) // 获取当前程序注册支持的所有k8s集群

	CrdFlinkDeploymentList(k8sClusterName string, filter Filter) (CrdFlinkDeploymentGetResponse, error)
	CrdFlinkDeploymentApply(CreateFlinkClusterRequest) (CreateFlinkClusterResponse, error)
	CrdFlinkDeploymentDelete(DeleteFlinkClusterRequest) error
	CrdFlinkSessionJobList(k8sClusterName string, filter Filter) (CrdFlinkSessionJobGetResponse, error)
	CrdFlinkSessionJobSubmit(CreateFlinkSessionJobRequest) (any, error)
	CrdFlinkSessionJobDelete(DeleteFlinkSessionJobRequest) error
}

type K8SConfig struct {
	Clusters map[string]Cluster `json:"clusters"`
}

type Cluster struct {
	KubeConfig *string `json:"kube_config"` // base64
	KubePath   *string `json:"kube_path"`   // path
}
