package io

import (
	"context"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *k8sClient) ConfigMapList(filter model.Filter) (*v1.ConfigMapList, error) {
	var namespace string
	if filter.NameSpace != nil {
		namespace = *filter.NameSpace
	} else {
		namespace = v1.NamespaceDefault
	}
	result, err := c.clientSet.CoreV1().ConfigMaps(namespace).List(context.TODO(), filter.ToOptions())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *k8sClient) ConfigMapApply(req model.ApplyConfigMapRequest) (any, error) {
	if req.Namespace == nil {
		req.Namespace = tea.String(v1.NamespaceDefault)
	}
	configMap, err := req.NewConfigMap()
	if err != nil {
		return nil, err
	}
	result, err := c.clientSet.CoreV1().ConfigMaps(*req.Namespace).Apply(context.TODO(), configMap, req.ToOptions())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *k8sClient) ConfigMapDelete(namespace, name string) error {
	err := c.clientSet.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
