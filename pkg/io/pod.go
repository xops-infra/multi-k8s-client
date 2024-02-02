package io

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *k8sClient) PodList(namespace string) (*v1.PodList, error) {
	return c.clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
}

func (c *k8sClient) PodGet(namespace, podName string) (*v1.Pod, error) {
	return c.clientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
}
