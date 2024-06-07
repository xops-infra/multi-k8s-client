package io

import (
	"context"

	"github.com/xops-infra/multi-k8s-client/pkg/model"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
)

func (c *k8sClient) DeploymentList(filter model.Filter) (*appv1.DeploymentList, error) {
	var namespace string
	if filter.NameSpace != nil {
		namespace = *filter.NameSpace
	} else {
		namespace = corev1.NamespaceDefault
	}
	result, err := c.clientSet.AppsV1().Deployments(namespace).List(context.TODO(), filter.ToOptions())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *k8sClient) DeploymentApply(namespace string, yaml *appsv1.DeploymentApplyConfiguration) (any, error) {
	result, err := c.clientSet.AppsV1().Deployments(namespace).Apply(context.TODO(), yaml, metav1.ApplyOptions{Force: true, FieldManager: "multi-k8s-client"})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *k8sClient) DeploymentDelete(namespace, name string) error {
	err := c.clientSet.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
