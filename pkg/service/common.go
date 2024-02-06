package service

import (
	"fmt"

	"github.com/xops-infra/multi-k8s-client/pkg/model"
)

type K8SService struct {
	IOs map[string]model.K8SIO
}

func NewK8SService(ios map[string]model.K8SIO) model.K8SContract {
	return &K8SService{
		IOs: ios,
	}
}

func (s *K8SService) CrdFlinkDeploymentApply(cluster, namespace string, yaml map[string]any) (any, error) {
	if io, ok := s.IOs[cluster]; ok {
		return io.CrdFlinkDeploymentApply(namespace, yaml)
	}
	return nil, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkDeploymentDelete(cluster, namespace, name string) error {
	if io, ok := s.IOs[cluster]; ok {
		return io.CrdFlinkDeploymentDelete(namespace, name)
	}
	return fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkSessionJobSubmit(cluster, namespace string, yaml map[string]any) (any, error) {
	if io, ok := s.IOs[cluster]; ok {
		return io.CrdFlinkSessionJobSubmit(namespace, yaml)
	}
	return nil, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkSessionJobDelete(cluster, namespace, name string) error {
	if io, ok := s.IOs[cluster]; ok {
		return io.CrdFlinkSessionJobDelete(namespace, name)
	}
	return fmt.Errorf("cluster not found")
}
