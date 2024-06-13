package io

import (
	"context"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (io *k8sClient) ServiceList(filter model.Filter) (*v1.ServiceList, error) {
	var namespace string
	if filter.NameSpace != nil {
		namespace = *filter.NameSpace
	} else {
		namespace = v1.NamespaceDefault
	}

	resp, err := io.clientSet.CoreV1().Services(namespace).List(context.TODO(), filter.ToOptions())
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (io *k8sClient) ServiceApply(req model.ApplyServiceRequest) (*v1.Service, error) {
	if req.Namespace == nil {
		req.Namespace = tea.String(v1.NamespaceDefault)
	}
	service, err := req.NewService()
	if err != nil {
		return nil, err
	}
	result, err := io.clientSet.CoreV1().Services(*req.Namespace).Apply(context.TODO(), service, req.ToOptions())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (io *k8sClient) ServiceDelete(namespace, name string) error {
	err := io.clientSet.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
