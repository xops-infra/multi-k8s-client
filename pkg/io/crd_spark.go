package io

import (
	"context"
	"fmt"

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

func (c *k8sClient) CrdSparkApplicationApply(yaml map[string]any) (any, error) {
	if yaml["metadata"].(map[string]any)["name"] == nil {
		return nil, fmt.Errorf("name is required")
	}
	sparkApplicationRes := GetGVR("sparkoperator.k8s.io", "v1beta2", "sparkapplications")

	sparkApplication := &unstructured.Unstructured{
		Object: yaml,
	}
	namsepace := "default"
	if yaml["metadata"].(map[string]any)["namespace"] != nil {
		namsepace = yaml["metadata"].(map[string]any)["namespace"].(string)
	}
	result, err := c.dynamic.Resource(sparkApplicationRes).Namespace(namsepace).Create(context.TODO(), sparkApplication, metav1.CreateOptions{})
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
