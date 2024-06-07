package service

import (
	"fmt"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
)

// 查询 flinkNamespace 下的所有 deployment
func (s *K8SService) FlinkV12ClusterList(k8sClusterName string, filter model.Filter) (model.CrdFlinkDeploymentGetResponse, error) {
	if io, ok := s.IOs[k8sClusterName]; ok {
		resp, err := io.DeploymentList(filter)
		if err != nil {
			return model.CrdFlinkDeploymentGetResponse{}, err
		}
		var items []model.CrdFlinkDeployment
		for _, item := range resp.Items {
			// fmt.Println(tea.Prettify(item.Object))
			items = append(items, model.CrdFlinkDeployment{
				ClusterName: item.GetName(),
				NameSpace:   item.GetNamespace(),
				Annotation:  item.GetAnnotations(),
			})
		}
		return model.CrdFlinkDeploymentGetResponse{
			Total: len(resp.Items),
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
	if io, ok := s.IOs[k8sClusterName]; ok {
		namespace := "flink"
		if req.NameSpace != nil {
			namespace = *req.NameSpace
		}
		// pvc
		_, err := io.PvcApply(model.ApplyPvcRequest{
			Name:        req.Name,
			Namespace:   tea.String(namespace),
			Label:       map[string]string{"owner": *req.Owner},
			StorageSize: req.JobManager.PvcSize,
		})
		if err != nil {
			return model.CreateResponse{}, err
		}

		// service

	}
	return model.CreateResponse{}, fmt.Errorf("cluster not found")
}

func (s *K8SService) FlinkV12ClusterDelete(k8sClusterName string, req model.DeleteFlinkClusterRequest) error {

	return nil
}
