package io

import (
	"context"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *k8sClient) PvcList(filter model.Filter) (*v1.PersistentVolumeClaimList, error) {
	var namespace string
	if filter.NameSpace != nil {
		namespace = *filter.NameSpace
	} else {
		namespace = apiv1.NamespaceDefault
	}
	result, err := c.clientSet.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), filter.ToOptions())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *k8sClient) PvcApply(req model.ApplyPvcRequest) (any, error) {
	if req.Namespace == nil {
		req.Namespace = tea.String(apiv1.NamespaceDefault)
	}
	claim, err := req.NewPVC()
	if err != nil {
		return nil, err
	}
	result, err := c.clientSet.CoreV1().PersistentVolumeClaims(*req.Namespace).Apply(context.TODO(), claim, req.ToOptions())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *k8sClient) PvcDelete(namespace, name string) error {
	err := c.clientSet.CoreV1().PersistentVolumeClaims(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
