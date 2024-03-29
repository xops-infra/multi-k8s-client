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

func (s *K8SService) GetK8SCluster() ([]string, error) {
	var clusterNames []string
	for k := range s.IOs {
		clusterNames = append(clusterNames, k)
	}
	return clusterNames, nil
}

func (s *K8SService) CrdFlinkDeploymentList(k8sClusterName string, filter model.Filter) (model.CrdFlinkDeploymentGetResponse, error) {
	if io, ok := s.IOs[k8sClusterName]; ok {
		resp, err := io.CrdFlinkDeploymentList(filter)
		if err != nil {
			return model.CrdFlinkDeploymentGetResponse{}, err
		}
		var items []model.CrdFlinkDeployment
		for _, item := range resp.Items {
			// fmt.Println(tea.Prettify(item.Object))
			items = append(items, model.CrdFlinkDeployment{
				ClusterName:                item.GetName(),
				NameSpace:                  item.GetNamespace(),
				JobStatus:                  item.Object["status"].(map[string]any)["jobStatus"],
				Annotation:                 item.GetAnnotations(),
				ClusterInfo:                item.Object["status"].(map[string]any)["clusterInfo"],
				JobManagerDeploymentStatus: item.Object["status"].(map[string]any)["jobManagerDeploymentStatus"],
				Error:                      item.Object["status"].(map[string]any)["error"],
			})
		}
		return model.CrdFlinkDeploymentGetResponse{
			Total: len(resp.Items),
			Items: items,
		}, nil
	}
	return model.CrdFlinkDeploymentGetResponse{}, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkDeploymentApply(req model.CreateFlinkClusterRequest) (model.CreateFlinkClusterResponse, error) {
	if io, ok := s.IOs[tea.StringValue(req.K8SClusterName)]; ok {
		resp, err := io.CrdFlinkDeploymentApply(tea.StringValue(req.NameSpace), req.ToYaml())
		if err != nil {
			return model.CreateFlinkClusterResponse{}, err
		}
		return model.CreateFlinkClusterResponse{
			Result: resp,
			Info:   fmt.Sprintf("\nsuccess\tkubectl port-forward svc/%s-rest 8081", *req.ClusterName),
		}, nil
	}
	return model.CreateFlinkClusterResponse{}, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkDeploymentDelete(req model.DeleteFlinkClusterRequest) error {
	if io, ok := s.IOs[tea.StringValue(req.K8SClusterName)]; ok {
		return io.CrdFlinkDeploymentDelete(tea.StringValue(req.NameSpace), *req.ClusterName)
	}
	return fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkSessionJobList(k8sClusterName string, filter model.Filter) (model.CrdFlinkSessionJobGetResponse, error) {
	if io, ok := s.IOs[k8sClusterName]; ok {
		resp, err := io.CrdFlinkSessionJobList(filter)
		if err != nil {
			return model.CrdFlinkSessionJobGetResponse{}, err
		}
		var items []model.CrdFlinkSessionJobItem
		for _, item := range resp.Items {
			var lifecycleState string
			if _, ok := item.Object["status"].(map[string]any)["lifecycleState"]; !ok {
				lifecycleState = "-"
			} else {
				lifecycleState = item.Object["status"].(map[string]any)["lifecycleState"].(string)
			}
			job := model.CrdFlinkSessionJobItem{
				ClusterName:    item.Object["spec"].(map[string]any)["deploymentName"].(string),
				SubmitJobName:  item.GetName(),
				LifecycleState: lifecycleState,
				Job:            item.Object["spec"].(map[string]any)["job"],
				NameSpace:      item.GetNamespace(),
				Error:          item.Object["status"].(map[string]any)["error"],
				Annotation:     item.GetAnnotations(),
			}

			// fmt.Println(tea.Prettify(item))
			jobStatus := item.Object["status"].(map[string]any)["jobStatus"].(map[string]any)
			if jobStatus["state"] != nil {
				job.Status = jobStatus["state"].(string)
			}
			if jobStatus["jobName"] != nil {
				job.JobName = jobStatus["jobName"].(string)
			}
			if jobStatus["jobId"] != nil {
				job.JobId = jobStatus["jobId"].(string)
			}
			items = append(items, job)
		}
		return model.CrdFlinkSessionJobGetResponse{
			Total: len(resp.Items),
			Items: items,
		}, nil
	}
	return model.CrdFlinkSessionJobGetResponse{}, fmt.Errorf("cluster not found")
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
