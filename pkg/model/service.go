package model

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type Port struct {
	Name       *string `json:"name" binding:"required"`
	Protocol   *string `json:"protocol" binding:"required"`
	Port       *int32  `json:"port" binding:"required"`
	TargetPort *int32  `json:"targetPort"`
	NodePort   *int32  `json:"nodePort"`
}

type ServiceSpec struct {
	Type     *string           `json:"type"` // ClusterIP, NodePort, LoadBalancer
	Ports    []Port            `json:"ports"`
	Selector map[string]string `json:"selector"`
}

type ApplyServiceRequest struct {
	Namespace   *string           `json:"namespace"`
	Name        *string           `json:"name" binding:"required"`
	Spec        *ServiceSpec      `json:"spec" binding:"required"`
	Labels      map[string]string `json:"label"`
	Annotations map[string]string `json:"annotations"`
}

func (req *ApplyServiceRequest) NewService() (*corev1.ServiceApplyConfiguration, error) {
	var namespace string
	if req.Namespace != nil {
		namespace = *req.Namespace
	} else {
		namespace = v1.NamespaceDefault
	}
	if req.Name == nil || req.Spec == nil || req.Spec.Type == nil {
		return nil, fmt.Errorf("need name and spec and type required")
	}
	yaml := corev1.Service(*req.Name, namespace)

	// spec
	yaml.Spec = corev1.ServiceSpec()
	yaml.Spec.WithType(v1.ServiceType(*req.Spec.Type))
	if req.Spec.Ports != nil {
		for _, port := range req.Spec.Ports {
			if port.Name == nil || port.Protocol == nil || port.Port == nil {
				return nil, fmt.Errorf("need name, protocol and port required")
			}
			portal := v1.Protocol(*port.Protocol)
			pport := corev1.ServicePortApplyConfiguration{
				Name:     port.Name,
				Protocol: &portal,
				Port:     port.Port,
			}
			if port.TargetPort != nil {
				pport.TargetPort = &intstr.IntOrString{
					IntVal: *port.TargetPort,
				}
			}
			if port.NodePort != nil {
				pport.NodePort = port.NodePort
			}
			yaml.Spec.Ports = append(yaml.Spec.Ports, pport)
		}
	}
	if req.Spec.Selector != nil {
		yaml.Spec.Selector = req.Spec.Selector
	}
	// metadata
	if req.Labels != nil {
		yaml.Labels = req.Labels
	}
	if req.Annotations != nil {
		yaml.Annotations = req.Annotations
	}
	return yaml, nil
}

func (req *ApplyServiceRequest) ToOptions() metav1.ApplyOptions {
	return metav1.ApplyOptions{
		FieldManager: "multi-k8s-client",
	}
}
