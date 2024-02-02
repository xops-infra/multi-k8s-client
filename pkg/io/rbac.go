package io

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *k8sClient) RbacList(namespace string) (*v1.RoleList, error) {
	return c.clientSet.RbacV1().Roles(namespace).List(context.TODO(), metav1.ListOptions{})
}
