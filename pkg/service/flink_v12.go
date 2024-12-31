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
			var clustername, flinkType string
			if strings.HasSuffix(item.GetName(), "-jobmanager") {
				clustername = strings.TrimSuffix(item.GetName(), "-jobmanager")
				flinkType = "jobmanager"
			} else if strings.HasSuffix(item.GetName(), "-taskmanager") {
				clustername = strings.TrimSuffix(item.GetName(), "-taskmanager")
				flinkType = "taskmanager"
			}
			if clustername == "" {
				continue
			}

			// 获取 deployment 类型
			if c, ok := clusterMap[clustername]; ok {
				// 已经存在的集群，因为有 2 种 deployment, 丰富数据
				c.Status.(map[string]any)[flinkType] = item.Status
				c.Annotation.(map[string]any)[flinkType] = item.GetAnnotations()
				if flinkType == "taskmanager" {
					// 只记录 taskmanager 配置信息
					for k, v := range model.GetInfoFromDeploymentForV12(item) {
						c.Info[k] = v
					}
				}
				continue
			}

			var clusterItem = model.CrdFlinkDeployment{
				NameSpace:   item.GetNamespace(),
				ClusterName: clustername,
				Annotation: map[string]any{
					flinkType: item.GetAnnotations(),
				},
				Status: map[string]any{
					flinkType: item.Status,
				},
				Labels:       item.GetLabels(), // 集群创建时的标签都一样这里就取第一个
				LoadBalancer: map[string]string{},
				Info:         model.CrdFlinkDeploymentInfo{},
			}
			clusterMap[clustername] = clusterItem
		}

		// 转换
		items := make([]model.CrdFlinkDeployment, 0)
		for _, v := range clusterMap {
			flinkconfig := make(map[string]any, 0)
			// 获取 flink configmap 内容
			flinkConfigs, err := io.ConfigMapList(model.Filter{
				NameSpace:     filter.NameSpace,
				LabelSelector: tea.String(fmt.Sprintf("app=%s", v.ClusterName)),
				// FieldSelector: tea.String(fmt.Sprintf("metadata.name=%s", fmt.Sprintf(model.ConfigMapV12Name, v.ClusterName))),
			})
			if err != nil {
				return model.CrdFlinkDeploymentGetResponse{}, fmt.Errorf("get flink v12 configmap error: %v", err)
			}
			for _, cv := range flinkConfigs.Items {
				// 提取 flink-conf.yaml key 内容
				if d, ok := cv.Data["flink-conf.yaml"]; ok {
					_flinkconfig, err := model.ConvertYamlToMap(d)
					if err != nil {
						return model.CrdFlinkDeploymentGetResponse{}, fmt.Errorf("get flink v12 configmap error: %v", err)
					}
					flinkconfig = _flinkconfig
					break
				}
			}
			v.FlinkConfig = flinkconfig

			// 增加 LoadBlance 连接信息
			lbResp, err := io.ServiceList(model.Filter{
				NameSpace:     tea.String(v.NameSpace),
				FieldSelector: tea.String(fmt.Sprintf("metadata.name=%s-jobmanager-lb-service", v.ClusterName)), // app-session-jobmanager-lb-service
			})
			if err == nil {
				for k, item := range lbResp.Items {
					v.Status.(map[string]any)[fmt.Sprintf("loadbalance-%d", k)] = fmt.Sprintf("%s:%d", item.Status.LoadBalancer.Ingress[0].IP, item.Spec.Ports[0].Port)
					v.LoadBalancer[fmt.Sprintf("loadbalance-%d", k)] = fmt.Sprintf("%s:%d", item.Status.LoadBalancer.Ingress[0].IP, item.Spec.Ports[0].Port)
				}
			}
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
func (s *K8SService) FlinkV12ClusterCreate(k8sClusterName string, req model.CreateFlinkV12ClusterRequest) (model.CreateResponse, error) {
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
		// 优化打印 LB 的请求公网地址+端口，创建的时候看不到，改到查询里面展示

		if len(errors) > 0 {
			resp.Result = errors
			return resp, fmt.Errorf("k8s apply error: %v", errors)
		}
		resp.Result = "create deployment*2, configmap, service*2, pvc*1 success"
		return resp, nil
	}
	return resp, fmt.Errorf("cluster not found")
}

func (s *K8SService) FlinkV12ClusterApply(k8sClusterName, namespace, clusterName string, req model.ApplyFlinkV12ClusterRequest) error {
	if io, ok := s.IOs[k8sClusterName]; ok {
		// 涉及到 deployment和 configmap更新
		// 1. 更新 deployment
		if namespace == "" {
			namespace = "default"
		}

		if req.Labels != nil {
			_, err := io.DeploymentApply(model.ApplyDeploymentRequest{
				ClusterName: tea.String(fmt.Sprintf(model.JobManagerDeploymentName, clusterName)),
				Namespace:   tea.String(namespace),
				Labels:      req.Labels,
			})
			if err != nil {
				return fmt.Errorf("job deployment apply error: %v", err)
			}
			_, err = io.DeploymentApply(model.ApplyDeploymentRequest{
				ClusterName: tea.String(fmt.Sprintf(model.TaskManagerDeploymentName, clusterName)),
				Namespace:   tea.String(namespace),
				Labels:      req.Labels,
			})
			if err != nil {
				return fmt.Errorf("task deployment apply error: %v", err)
			}
		}
		if req.FlinkConfiguration != nil {
			// 2. 更新 configmap
			_, err := io.ConfigMapApply(model.ApplyConfigMapRequest{
				Name:      tea.String(fmt.Sprintf(model.ConfigMapV12Name, clusterName)),
				Namespace: tea.String(namespace),
				Labels:    req.Labels,
				Data: map[string]string{
					"flink-conf.yaml":     model.ToString(req.FlinkConfiguration),
					"logback-console.xml": model.LogbackConsole,
				},
			})
			if err != nil {
				return fmt.Errorf("configmap apply error: %v", err)
			}
		}

		// 3. 更新 service
		// TODO: invalid: spec.ports: Required value ???
		// _, err = io.ServiceApply(model.ApplyServiceRequest{
		// 	Name:      tea.String(fmt.Sprintf(model.JobManagerServiceName, clusterName)),
		// 	Namespace: tea.String(namespace),
		// 	Labels:    req.Labels,
		// })
		// if err != nil {
		// 	return fmt.Errorf("service apply error: %v", err)
		// }

		return nil
	}
	return fmt.Errorf("cluster not found")
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

		// 删除配置，衍生的资源也会被删除
		/*
			labels:
			   app: flink-session
			   configmap-type: high-availability
			   type: flink-native-kubernetes
		*/
		err = io.ConfigMapDelete(tea.StringValue(req.NameSpace), fmt.Sprintf(model.ConfigMapV12Name, *req.ClusterName))
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("configmap delete error: %v", err)
			}
		}
		resp, err := io.ConfigMapList(model.Filter{
			NameSpace:     tea.String("flink"),
			LabelSelector: tea.String(fmt.Sprintf("app=%s,configmap-type=high-availability,type=flink-native-kubernetes", *req.ClusterName)),
		})
		if err != nil {
			return fmt.Errorf("list configmaps error: %v", err)
		}
		for _, item := range resp.Items {
			err = io.ConfigMapDelete(tea.StringValue(req.NameSpace), item.GetName())
			if err != nil {
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
