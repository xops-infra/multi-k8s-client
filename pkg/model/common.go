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

	// CRD
	CrdApplyFlinkDeployment(namespace string, yaml map[string]any) (any, error)
	CrdSubmitFlinkSessionJob(namespace string, yaml map[string]any) (any, error) // for flink session cluster, can't be used for application cluster
	CrdDeleteFlinkDeployment(namespace, name string) error
}
