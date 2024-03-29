package io

import (
	"context"

	"github.com/xops-infra/multi-k8s-client/pkg/model"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (c *k8sClient) CrdSparkApplicationList(filter model.Filter) (*unstructured.UnstructuredList, error) {
	sparkApplicationRes := GetGVR("sparkoperator.k8s.io", "v1beta2", "sparkapplications")
	var namespace string
	if filter.NameSpace != nil {
		namespace = *filter.NameSpace
	} else {
		namespace = apiv1.NamespaceDefault
	}
	result, err := c.dynamic.Resource(sparkApplicationRes).Namespace(namespace).List(context.TODO(), filter.ToOptions())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *k8sClient) CrdSparkApplicationSubmit(namespace string, yaml map[string]any) (any, error) {
	sparkApplicationRes := GetGVR("sparkoperator.k8s.io", "v1beta2", "sparkapplications")

	sparkApplication := &unstructured.Unstructured{
		Object: yaml,
	}
	if namespace == "" {
		namespace = apiv1.NamespaceDefault
	}
	sparkApplication.SetNamespace(namespace)
	result, err := c.dynamic.Resource(sparkApplicationRes).Namespace(namespace).Create(context.TODO(), sparkApplication, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *k8sClient) CrdSparkApplicationDelete(namespace, name string) error {
	sparkApplicationRes := GetGVR("sparkoperator.k8s.io", "v1beta2", "sparkapplications")
	err := c.dynamic.Resource(sparkApplicationRes).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
