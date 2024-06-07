package model

import (
	"fmt"

	"github.com/alibabacloud-go/tea/tea"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type ApplyConfigMapRequest struct {
	Namespace *string
	Name      *string
	Labels    map[string]string
	Data      map[string]string
}

func (req *ApplyConfigMapRequest) NewConfigMap() (*corev1.ConfigMapApplyConfiguration, error) {
	if req.Name == nil {
		return nil, fmt.Errorf("name is required")
	}
	if req.Namespace == nil {
		req.Namespace = tea.String(v1.NamespaceDefault)
	}
	configMap := corev1.ConfigMap(*req.Name, *req.Namespace).WithLabels(req.Labels).WithData(req.Data)
	return configMap, nil
}

func (req *ApplyConfigMapRequest) ToOptions() metav1.ApplyOptions {
	return metav1.ApplyOptions{
		FieldManager: "multi-k8s-client",
	}
}
