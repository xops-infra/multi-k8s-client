package io

import (
	"context"

	"github.com/xops-infra/multi-k8s-client/pkg/model"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *k8sClient) PodList(filter model.Filter) (*v1.PodList, error) {
	var namespace string
	if filter.NameSpace != nil {
		namespace = *filter.NameSpace
	} else {
		namespace = apiv1.NamespaceDefault
	}
	return c.clientSet.CoreV1().Pods(namespace).List(context.TODO(), filter.ToOptions())
}

func (c *k8sClient) PodGet(namespace, podName string) (*v1.Pod, error) {
	return c.clientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
}
