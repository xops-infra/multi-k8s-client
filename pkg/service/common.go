package service

import (
	"fmt"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/xops-infra/multi-k8s-client/pkg/io"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
)

type K8SService struct {
	IOs map[string]model.K8SIO
}

func NewK8SService(configs []model.Cluster) (model.K8SContract, error) {
	var ios = make(map[string]model.K8SIO)
	for _, cluster := range configs {
		newClient, err := io.NewK8SClient(cluster)
		if err != nil {
			return nil, err
		}
		if cluster.Alias == nil || cluster.Name == nil {
			return nil, fmt.Errorf("cluster name or alias is nil")
		}
		ios[*cluster.Alias] = newClient
	}
	return &K8SService{
		IOs: ios,
	}, nil
}

func (s *K8SService) GetK8SCluster() []model.ClusterInfo {
	var clusterNames []model.ClusterInfo
	for _, v := range s.IOs {
		clusterNames = append(clusterNames, v.GetClusterInfo())
	}
	return clusterNames
}

// 因为 JM 是单副本，所以支持的只有 TM副本调整
func (s *K8SService) CrdFlinkTMScale(k8sClusterName string, req model.CrdFlinkTMScaleRequest) error {
	if io, ok := s.IOs[k8sClusterName]; ok {
		_, err := io.DeploymentScale(tea.StringValue(req.NameSpace), fmt.Sprintf(model.TaskManagerDeploymentName, *req.ClusterName), *req.Replicas)
		return err
	}
	return fmt.Errorf("cluster %s not found, available cluster: %v", k8sClusterName, tea.Prettify(s.GetK8SCluster()))
}

func (s *K8SService) CrdFlinkDeploymentRestart(k8sClusterName string, req model.RestartFlinkClusterRequest) error {
	if io, ok := s.IOs[k8sClusterName]; ok {
		var deploymentName []string
		switch req.Type {
		case model.FlinkTypeALL:
			deploymentName = []string{fmt.Sprintf(model.TaskManagerDeploymentName, *req.ClusterName),
				fmt.Sprintf(model.JobManagerDeploymentName, *req.ClusterName)}
		case model.FlinkTypeJM:
			deploymentName = []string{fmt.Sprintf(model.JobManagerDeploymentName, *req.ClusterName)}
		case model.FlinkTypeTM:
			deploymentName = []string{fmt.Sprintf(model.TaskManagerDeploymentName, *req.ClusterName)}
		default:
			return fmt.Errorf("type not found, only support ALL, TM, JM")
		}
		for _, name := range deploymentName {
			_, err := io.DeploymentRestart(tea.StringValue(req.NameSpace), name)
			if err != nil {
				return err
			}
		}

		return nil
	}
	return fmt.Errorf("cluster %s not found, available cluster: %v", k8sClusterName, tea.Prettify(s.GetK8SCluster()))
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
			// 增加 LoadBlance 连接信息
			lbResp, err := io.ServiceList(model.Filter{
				NameSpace:     tea.String(item.GetNamespace()),
				FieldSelector: tea.String(fmt.Sprintf("metadata.name=%s", fmt.Sprintf(model.JobManagerLBServiceName, item.GetName()))), // app-session-jobmanager-lb-service
			})
			if err != nil {
				return model.CrdFlinkDeploymentGetResponse{}, err
			}
			v := model.CrdFlinkDeployment{
				ClusterName:  item.GetName(),
				NameSpace:    item.GetNamespace(),
				Labels:       item.GetLabels(),
				Status:       item.Object["status"].(map[string]any),
				Annotation:   item.GetAnnotations(),
				LoadBalancer: map[string]any{},
			}
			for k, item := range lbResp.Items {
				v.Status.(map[string]any)[fmt.Sprintf("loadbalance-%d", k)] = fmt.Sprintf("%s:%d", item.Status.LoadBalancer.Ingress[0].IP, item.Spec.Ports[0].Port)
				v.LoadBalancer.(map[string]any)[fmt.Sprintf("loadbalance-%d", k)] = fmt.Sprintf("%s:%d", item.Status.LoadBalancer.Ingress[0].IP, item.Spec.Ports[0].Port)
			}

			items = append(items, v)
		}
		return model.CrdFlinkDeploymentGetResponse{
			Total: len(resp.Items),
			Items: items,
		}, nil
	}
	return model.CrdFlinkDeploymentGetResponse{}, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkDeploymentApply(k8sCluster string, req model.CreateFlinkClusterRequest) (model.CreateResponse, error) {
	if io, ok := s.IOs[k8sCluster]; ok {
		var response model.CreateResponse
		_, err := io.CrdFlinkDeploymentApply(req.ToYaml())
		if err != nil {
			return model.CreateResponse{}, err
		}
		response.Info = "create FlinkDeployment success!"
		if req.LoadBalancer != nil {
			// 创建 loadbalancer
			LBServiceYaml := req.NewLBService()
			_, err := io.ServiceApply(LBServiceYaml)
			if err != nil {
				return model.CreateResponse{}, err
			}
			response.Info += " create LoadBalancer success!"
		}
		return response, nil
	}
	return model.CreateResponse{}, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkDeploymentDelete(k8sClusterName string, req model.DeleteFlinkClusterRequest) error {
	if io, ok := s.IOs[k8sClusterName]; ok {
		// 删除 deployment,如果存在 LB 也一起删掉
		err := io.CrdFlinkDeploymentDelete(tea.StringValue(req.NameSpace), *req.ClusterName)
		if err != nil {
			return err
		}
		// 删除 service
		io.ServiceDelete(tea.StringValue(req.NameSpace), fmt.Sprintf(model.JobManagerLBServiceName, *req.ClusterName))

		return nil
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

func (s *K8SService) CrdFlinkSessionJobSubmit(k8sClusterName string, req model.CreateFlinkSessionJobRequest) (any, error) {
	if io, ok := s.IOs[k8sClusterName]; ok {
		return io.CrdFlinkSessionJobSubmit(tea.StringValue(req.NameSpace), req.ToYaml())
	}
	return nil, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdFlinkSessionJobDelete(k8sClusterName string, req model.DeleteFlinkSessionJobRequest) error {
	if io, ok := s.IOs[k8sClusterName]; ok {
		return io.CrdFlinkSessionJobDelete(tea.StringValue(req.NameSpace), *req.JobName)
	}
	return fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdSparkApplicationList(k8sClusterName string, filter model.Filter) (model.CrdSparkApplicationGetResponse, error) {
	if io, ok := s.IOs[k8sClusterName]; ok {
		resp, err := io.CrdSparkApplicationList(filter)
		if err != nil {
			return model.CrdSparkApplicationGetResponse{}, err
		}
		var items []model.CrdSparkApplication
		for _, item := range resp.Items {
			// fmt.Println(tea.Prettify(item))
			items = append(items, model.CrdSparkApplication{
				Name:       item.GetName(),
				Namespace:  item.GetNamespace(),
				Status:     item.Object["status"].(map[string]any)["applicationState"].(map[string]any)["state"].(string),
				Attempts:   item.Object["status"].(map[string]any)["executionAttempts"].(int64),
				StartTime:  item.Object["status"].(map[string]any)["lastSubmissionAttemptTime"].(string),
				FinishTime: item.Object["status"].(map[string]any)["terminationTime"].(string),
				Age:        time.Since(item.GetCreationTimestamp().Time).String(),
			})
		}
		return model.CrdSparkApplicationGetResponse{
			Total: len(resp.Items),
			Items: items,
		}, nil
	}
	return model.CrdSparkApplicationGetResponse{}, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdSparkApplicationGet(k8sClusterName, namespace, name string) (model.CrdResourceDetail, error) {
	if io, ok := s.IOs[k8sClusterName]; ok {
		resp, err := io.CrdSparkApplicationList(model.Filter{
			NameSpace:     &namespace,
			FieldSelector: tea.String(fmt.Sprintf("metadata.name=%s", name)),
		})
		if err != nil {
			return model.CrdResourceDetail{}, err
		}
		if len(resp.Items) != 1 {
			return model.CrdResourceDetail{}, fmt.Errorf("spark application not found")
		}
		item := resp.Items[0]
		return model.CrdResourceDetail{
			Kind:       item.GetObjectKind().GroupVersionKind().Kind,
			ApiVersion: item.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Name:       item.GetName(),
			Metadata:   item.Object["metadata"].(map[string]any),
			Spec:       item.Object["spec"].(map[string]any),
			Status:     item.Object["status"].(map[string]any),
		}, nil
	}
	return model.CrdResourceDetail{}, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdSparkApplicationApply(k8sClusterName string, req model.CreateSparkApplicationRequest) (model.CreateResponse, error) {
	if io, ok := s.IOs[k8sClusterName]; ok {
		resp, err := io.CrdSparkApplicationApply(req.ToYaml())
		if err != nil {
			return model.CreateResponse{}, err
		}
		return model.CreateResponse{
			Result: resp,
			Info:   fmt.Sprintf("\nsuccess\tkubectl port-forward svc/%s 8081", *req.Name),
		}, nil
	}
	return model.CreateResponse{}, fmt.Errorf("cluster not found")
}

func (s *K8SService) CrdSparkApplicationDelete(k8sClusterName string, req model.DeleteSparkApplicationRequest) error {
	if io, ok := s.IOs[k8sClusterName]; ok {
		return io.CrdSparkApplicationDelete(tea.StringValue(req.Namespace), *req.Name)
	}
	return fmt.Errorf("cluster not found")
}
