package model

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type ApplyPvcRequest struct {
	Name             *string           `json:"name" binding:"required"`
	Namespace        *string           `json:"namespace"`
	Label            map[string]string `json:"label"`
	StorageClassName *string           `json:"storageClassName"`
	StorageSize      *int              `json:"storageSize" default:"10"` // 10G
}

func (a ApplyPvcRequest) ToOptions() metav1.ApplyOptions {
	return metav1.ApplyOptions{Force: true, FieldManager: "multi-k8s-client"}
}

func (a ApplyPvcRequest) NewPVC() (*corev1.PersistentVolumeClaimApplyConfiguration, error) {
	if a.Name == nil {
		return nil, fmt.Errorf("name is required")
	}
	namespace := v1.NamespaceDefault
	if a.Namespace != nil {
		namespace = *a.Namespace
	}
	config := corev1.PersistentVolumeClaim(*a.Name, namespace)
	if a.Label != nil {
		// 自动加上 app标签
		a.Label["app"] = *a.Name
		config.WithLabels(a.Label)
	}
	spec := &corev1.PersistentVolumeClaimSpecApplyConfiguration{}
	spec.StorageClassName = a.StorageClassName
	// 默认 10G
	storageSize := resource.NewQuantity(int64(10*1024*1024*1024), resource.BinarySI)
	if a.StorageSize != nil {
		storageSize = resource.NewQuantity(int64(*a.StorageSize*1024*1024*1024), resource.BinarySI)
	}
	spec.Resources = &corev1.VolumeResourceRequirementsApplyConfiguration{
		Requests: &v1.ResourceList{
			v1.ResourceStorage: *storageSize,
		},
	}
	spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}

	config.Spec = spec
	return config, nil
}
