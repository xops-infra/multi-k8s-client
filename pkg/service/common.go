package service

import (
	"fmt"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/xops-infra/multi-k8s-client/pkg/io"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
)

type K8SService struct {
	IOs map[string]model.K8SIO
}

func NewK8SService(configs model.K8SConfig) model.K8SContract {
	var ios = make(map[string]model.K8SIO)
	for k, v := range configs.Clusters {
		newClient, err := io.NewK8SClient(v)
		if err != nil {
			panic(err)
		}
		ios[k] = newClient
	}
	return &K8SService{
		IOs: ios,
	}
}

func (s *K8SService) CrdFlinkDeploymentApply(req model.CreateFlinkClusterRequest) (model.CreateFlinkClusterResponse, error) {
	if io, ok := s.IOs[tea.StringValue(req.K8SClusterName)]; ok {
		resp, err := io.CrdFlinkDeploymentApply(tea.StringValue(req.NameSpace), req.ToYaml())
		if err != nil {
			return model.CreateFlinkClusterResponse{}, err
		}
		return model.CreateFlinkClusterResponse{
			Result: resp,
			Info:   fmt.Sprintf("\nsuccess\tkubectl port-forward svc/%s-rest 8081", *req.MetaDataName),
		}, nil
	}
	return model.CreateFlinkClusterResponse{}, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkDeploymentDelete(req model.DeleteFlinkClusterRequest) error {
	if io, ok := s.IOs[tea.StringValue(req.K8SClusterName)]; ok {
		return io.CrdFlinkDeploymentDelete(tea.StringValue(req.NameSpace), *req.Name)
	}
	return fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkSessionJobSubmit(req model.CreateFlinkSessionJobRequest) (any, error) {
	if io, ok := s.IOs[tea.StringValue(req.K8SClusterName)]; ok {
		return io.CrdFlinkSessionJobSubmit(tea.StringValue(req.NameSpace), req.ToYaml())
	}
	return nil, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkSessionJobDelete(req model.DeleteFlinkSessionJobRequest) error {
	if io, ok := s.IOs[tea.StringValue(req.K8SClusterName)]; ok {
		return io.CrdFlinkSessionJobDelete(tea.StringValue(req.NameSpace), *req.JobName)
	}
	return fmt.Errorf("cluster not found")
}
