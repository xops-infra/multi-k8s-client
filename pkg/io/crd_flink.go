package io

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (c *k8sClient) CrdApplyFlinkDeployment(namespace string, yaml map[string]any) (any, error) {
	flinkDeploymentRes := GetGVR("flink.apache.org", "v1beta1", "flinkdeployments")

	flinkDeployment := &unstructured.Unstructured{
		Object: yaml,
	}
	if namespace == "" {
		namespace = apiv1.NamespaceDefault
	}
	flinkDeployment.SetNamespace(namespace)
	result, err := c.dynamic.Resource(flinkDeploymentRes).Namespace(namespace).Create(context.TODO(), flinkDeployment, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *k8sClient) CrdDeleteFlinkDeployment(namespace, name string) error {
	flinkDeploymentRes := GetGVR("flink.apache.org", "v1beta1", "flinkdeployments")
	if namespace == "" {
		namespace = apiv1.NamespaceDefault
	}
	err := c.dynamic.Resource(flinkDeploymentRes).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *k8sClient) CrdSubmitFlinkSessionJob(namespace string, yaml map[string]any) (any, error) {
	flinkJobRes := GetGVR("flink.apache.org", "v1beta1", "flinksessionjobs")
	flinkJob := &unstructured.Unstructured{
		Object: yaml,
	}
	if namespace == "" {
		namespace = apiv1.NamespaceDefault
	}
	flinkJob.SetNamespace(namespace)
	result, err := c.dynamic.Resource(flinkJobRes).Namespace(namespace).Create(context.TODO(), flinkJob, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return result, nil
}
