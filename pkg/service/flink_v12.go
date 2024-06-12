package service

import (
	"fmt"
	"strings"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
)

// 查询 flinkNamespace 下的所有 deployment
func (s *K8SService) FlinkV12ClusterList(k8sClusterName string, filter model.FilterFlinkV12) (model.CrdFlinkDeploymentGetResponse, error) {
	if io, ok := s.IOs[k8sClusterName]; ok {
		f := model.Filter{
			NameSpace: tea.String("default"),
		}
		if filter.NameSpace != nil {
			f.NameSpace = filter.NameSpace
		}
		if filter.Owner != nil {
			f.LabelSelector = tea.String(fmt.Sprintf("owner=%s", *filter.Owner))
		}
		if filter.Name != nil {
			f.LabelSelector = tea.String(fmt.Sprintf("app=%s", *filter.Name))
		}
		resp, err := io.DeploymentList(f)
		if err != nil {
			return model.CrdFlinkDeploymentGetResponse{}, err
		}

		clusterMap := make(map[string]model.CrdFlinkDeployment)
		for _, item := range resp.Items {
			// 创建的集群规则特这是带有 -jobmanager 或者 -taskmanager 的 deployment
			var clustername string
			if strings.HasSuffix(item.GetName(), "-jobmanager") {
				clustername = strings.TrimSuffix(item.GetName(), "-jobmanager")
			} else if strings.HasSuffix(item.GetName(), "-taskmanager") {
				clustername = strings.TrimSuffix(item.GetName(), "-taskmanager")
			}
			if clustername == "" {
				continue
			}

			if _, ok := clusterMap[clustername]; ok {
				// 已经存在的集群 丰富数据
				clusterMap[clustername].Status.(map[string]any)[item.GetName()] = item.Status
				clusterMap[clustername].Annotation.(map[string]any)[item.GetName()] = item.GetAnnotations()
				continue
			}
			// TODO: 增加 configmap配置信息
			var clusterItem = model.CrdFlinkDeployment{
				NameSpace:   item.GetNamespace(),
				ClusterName: clustername,
				Annotation: map[string]any{
					item.GetName(): item.GetAnnotations(),
				},
				Status: map[string]any{
					item.GetName(): item.Status,
				},
				Labels: item.GetLabels(), // 集群创建时的标签都一样这里就取第一个
			}
			clusterMap[clustername] = clusterItem
		}

		// 转换
		items := make([]model.CrdFlinkDeployment, 0)
		for _, v := range clusterMap {
			items = append(items, v)
		}
		return model.CrdFlinkDeploymentGetResponse{
			Total: len(items),
			Items: items,
		}, nil
	}
	return model.CrdFlinkDeploymentGetResponse{}, fmt.Errorf("cluster not found")
}

/*
创建资源包括：
 1. deployment x2
 2. service x1
 3. pvc x1
 4. configmap x1
*/
func (s *K8SService) FlinkV12ClustertApply(k8sClusterName string, req model.CreateFlinkV12ClusterRequest) (model.CreateResponse, error) {
	var resp model.CreateResponse
	if io, ok := s.IOs[k8sClusterName]; ok {
		// 1. 初始化所有配置，如果有问题直接报错
		jobDeployment := req.NewJobManagerDeployment()
		createJobD, err := model.NewDeploymentCreateFromMap(jobDeployment)
		if err != nil {
			return resp, err
		}
		taskDeployment := req.NewTaskManagerDeployment()
		createTaskD, err := model.NewDeploymentCreateFromMap(taskDeployment)
		if err != nil {
			return resp, err
		}

		configMapReq := req.NewConfigMap()
		pvcReq := req.NewPVC()
		serviceReq := req.NewService()
		ServiceLB := req.NewLBService()

		// 2. 创建
		// 如果只是辅助资源出错可以正常结束，返回结果标注错误资源和错误信息，后续人工干预
		errors := map[string]string{}
		// pvc
		_, err = io.PvcApply(pvcReq)
		if err != nil {
			return resp, fmt.Errorf("pvc apply error: %v", errors)
		}

		// deployment
		_, err = io.DeploymentCreate(createJobD)
		if err != nil {
			return resp, fmt.Errorf("job deployment apply error: %v", err)
		}
		_, err = io.DeploymentCreate(createTaskD)
		if err != nil {
			return resp, fmt.Errorf("task deployment apply error: %v", err)
		}

		// configmap
		_, err = io.ConfigMapApply(configMapReq)
		if err != nil {
			errors["configmap"] = err.Error()
		}

		// service
		_, err = io.ServiceApply(serviceReq)
		if err != nil {
			errors["service"] = err.Error()
		}
		_, err = io.ServiceApply(ServiceLB)
		if err != nil {
			errors["service-lb"] = err.Error()
		}

		if len(errors) > 0 {
			resp.Result = errors
			return resp, fmt.Errorf("k8s apply error: %v", errors)
		}
		resp.Info = "create deployment*2, configmap, service*2, pvc*1"
		return resp, nil
	}
	return resp, fmt.Errorf("cluster not found")
}

func (s *K8SService) FlinkV12ClusterDelete(k8sClusterName string, req model.DeleteFlinkClusterRequest) error {
	if io, ok := s.IOs[k8sClusterName]; ok {
		// 删除资源
		err := io.DeploymentDelete(tea.StringValue(req.NameSpace), fmt.Sprintf(model.JobManagerDeploymentName, *req.ClusterName))
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("job deployment delete error: %v", err)
			}
		}
		err = io.DeploymentDelete(tea.StringValue(req.NameSpace), fmt.Sprintf(model.TaskManagerDeploymentName, *req.ClusterName))
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("task deployment delete error: %v", err)
			}
		}
		err = io.ConfigMapDelete(tea.StringValue(req.NameSpace), fmt.Sprintf(model.ConfigMapV12Name, *req.ClusterName))
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("configmap delete error: %v", err)
			}
		}
		err = io.ServiceDelete(tea.StringValue(req.NameSpace), fmt.Sprintf(model.JobManagerServiceName, *req.ClusterName))
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("service delete error: %v", err)
			}
		}
		err = io.ServiceDelete(tea.StringValue(req.NameSpace), fmt.Sprintf(model.JobManagerLBServiceName, *req.ClusterName))
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("service delete error: %v", err)
			}
		}
		err = io.PvcDelete(tea.StringValue(req.NameSpace), fmt.Sprintf(model.PvcName, *req.ClusterName))
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("pvc delete error: %v", err)
			}
		}
		return nil
	}
	return fmt.Errorf("cluster not found")
}
