package model

import (
	podV1 "k8s.io/api/core/v1"
	rbacV1 "k8s.io/api/rbac/v1"
)

type K8SIO interface {
	// POD
	PodList(namespace string) (*podV1.PodList, error)
	PodGet(namespace, name string) (*podV1.Pod, error)

	// RBAC
	RbacList(namespace string) (*rbacV1.RoleList, error)

	// CRD Flink
	CrdFlinkDeploymentApply(namespace string, yaml map[string]any) (any, error)
	CrdFlinkDeploymentDelete(namespace, name string) error
	CrdFlinkSessionJobSubmit(namespace string, yaml map[string]any) (any, error) // for flink session cluster, can't be used for application cluster
	CrdFlinkSessionJobDelete(namespace, name string) error
}

type K8SContract interface {
	CrdFlinkDeploymentApply(CreateFlinkClusterRequest) (CreateFlinkClusterResponse, error)
	CrdFlinkDeploymentDelete(DeleteFlinkClusterRequest) error
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
