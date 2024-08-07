package model

import (
	appv1 "k8s.io/api/apps/v1"
	podV1 "k8s.io/api/core/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Filter struct {
	NameSpace     *string `json:"namespace"`      // default: default
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
	GetClusterInfo() ClusterInfo
	// POD
	PodList(namespace string) (*podV1.PodList, error)
	PodGet(namespace, name string) (*podV1.Pod, error)

	// DEPLOYMENT
	DeploymentList(filter Filter) (*appv1.DeploymentList, error)
	DeploymentApply(req ApplyDeploymentRequest) (any, error)
	DeploymentCreate(dep *appv1.Deployment) (any, error)
	DeploymentDelete(namespace, name string) error
	DeploymentScale(namespace, name string, replicas int32) (any, error)
	DeploymentRestart(namespace, name string) (any, error)

	// SERVICE
	ServiceList(filter Filter) (*podV1.ServiceList, error)
	ServiceApply(req ApplyServiceRequest) (*podV1.Service, error)
	ServiceDelete(namespace, name string) error

	// CONFIGMAP
	ConfigMapList(filter Filter) (*podV1.ConfigMapList, error)
	ConfigMapApply(req ApplyConfigMapRequest) (any, error)
	ConfigMapDelete(namespace, name string) error

	// PVC
	PvcList(filter Filter) (*podV1.PersistentVolumeClaimList, error)
	PvcApply(req ApplyPvcRequest) (any, error)
	PvcDelete(namespace, name string) error

	// RBAC
	RbacList(namespace string) (*rbacV1.RoleList, error)

	// CRD Flink
	CrdFlinkDeploymentList(Filter) (*unstructured.UnstructuredList, error)
	CrdFlinkDeploymentApply(yaml map[string]any) (any, error)
	CrdFlinkDeploymentDelete(namespace, name string) error

	CrdFlinkSessionJobList(Filter) (*unstructured.UnstructuredList, error)
	CrdFlinkSessionJobSubmit(namespace string, yaml map[string]any) (any, error) // for flink session cluster, can't be used for application cluster
	CrdFlinkSessionJobDelete(namespace, name string) error

	// CRD Spark
	CrdSparkApplicationList(Filter) (*unstructured.UnstructuredList, error)
	CrdSparkApplicationApply(yaml map[string]any) (any, error)
	CrdSparkApplicationDelete(namespace, name string) error
}

type K8SContract interface {
	GetK8SCluster() []ClusterInfo // 获取当前程序注册支持的所有k8s集群

	// Flink
	CrdFlinkDeploymentList(k8sClusterName string, filter Filter) (CrdFlinkDeploymentGetResponse, error)
	CrdFlinkDeploymentApply(k8sClusterName string, req CreateFlinkClusterRequest) (CreateResponse, error)
	CrdFlinkDeploymentDelete(k8sClusterName string, req DeleteFlinkClusterRequest) error
	CrdFlinkSessionJobList(k8sClusterName string, filter Filter) (CrdFlinkSessionJobGetResponse, error)
	CrdFlinkSessionJobSubmit(k8sClusterName string, req CreateFlinkSessionJobRequest) (any, error)
	CrdFlinkSessionJobDelete(k8sClusterName string, req DeleteFlinkSessionJobRequest) error
	CrdFlinkDeploymentRestart(k8sClusterName string, req RestartFlinkClusterRequest) error
	CrdFlinkTMScale(k8sClusterName string, req CrdFlinkTMScaleRequest) error
	// FlinkV1.12.7
	FlinkV12ClusterList(k8sClusterName string, filter FilterFlinkV12) (CrdFlinkDeploymentGetResponse, error)
	FlinkV12ClustertApply(k8sClusterName string, req CreateFlinkV12ClusterRequest) (CreateResponse, error)
	FlinkV12ClusterDelete(k8sClusterName string, req DeleteFlinkClusterRequest) error

	// Spark
	CrdSparkApplicationList(k8sClusterName string, filter Filter) (CrdSparkApplicationGetResponse, error)
	CrdSparkApplicationGet(k8sClusterName, namespace, name string) (CrdResourceDetail, error)
	CrdSparkApplicationApply(k8sClusterName string, req CreateSparkApplicationRequest) (CreateResponse, error)
	CrdSparkApplicationDelete(k8sClusterName string, req DeleteSparkApplicationRequest) error
}

type ClusterInfo struct {
	Name  *string `json:"name"`
	Alias *string `json:"alias"`
}

type Cluster struct {
	Name       *string `json:"name" binding:"required"`
	Alias      *string `json:"alias" binding:"required"`
	KubeConfig *string `json:"kube_config"` // base64
	KubePath   *string `json:"kube_path"`   // path
}

type CreateResponse struct {
	Result any    `json:"result"`
	Info   string `json:"info"`
}

type CrdResourceDetail struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Metadata   any    `json:"metadata"`
	Spec       any    `json:"spec"`
	Status     any    `json:"status"`
}
