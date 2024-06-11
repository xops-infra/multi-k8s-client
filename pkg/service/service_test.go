package service_test

import (
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
	"github.com/xops-infra/multi-k8s-client/pkg/service"
)

var k8s model.K8SContract

func init() {
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	panic(err)
	// }

	k8s = service.NewK8SService(model.K8SConfig{
		Clusters: map[string]model.Cluster{
			"test": {
				KubePath: tea.String("~/.kube/config"),
			},
		},
	})
}

// TEST FlinkV12ClusterList
func TestFlinkV12ClusterList(t *testing.T) {
	resp, err := k8s.FlinkV12ClusterList("test", model.FilterFlinkV12{
		NameSpace: tea.String("flink"),
		Name:      tea.String("app-session"),
		// Owner: tea.String("dingyingjie"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range resp.(map[string]any)["items"].(map[string][]model.CrdFlinkDeployment) {
		for _, item := range v {
			t.Log(k, tea.Prettify(item.Status))
		}
	}
	t.Log(tea.Prettify(resp.(map[string]any)["total"]))
}

// TEST FlinkV12ClustertApply
func TestFlinkV12ClustertApply(t *testing.T) {
	req := model.CreateFlinkV12ClusterRequest{
		Name:      tea.String("app-session"),
		NameSpace: tea.String("flink"),
		Owner:     tea.String("zhangsan"),
		Image:     tea.String("flink:1.12.7"),
		Env:       map[string]string{"ENABLE_BUILT_IN_PLUGINS": "flink-s3-fs-hadoop-1.12.7.jar;flink-s3-fs-presto-1.12.7.jar"},
		LoadBalancer: &model.LoadBalancerRequest{
			Annotations: map[string]string{"service.kubernetes.io/loadbalance-id": "lb-xxx"},
			Labels:      map[string]string{"used-by": "ui"},
		},
		JobManager: &model.JobManagerV12{
			PvcSize:  tea.Int(22),
			Resource: &model.FlinkResource{Memory: tea.String("1Gi"), CPU: tea.String("1")},
		},
		TaskManager: &model.TaskManagerV12{
			Nu:       tea.Int(2),
			Resource: &model.FlinkResource{Memory: tea.String("2Gi"), CPU: tea.String("1")},
		},
		FlinkConfigRequest: map[string]any{"taskmanager.numberOfTaskSlots": 2},
		NodeSelector:       map[string]any{"env": "flink"},
	}

	_, err := k8s.FlinkV12ClustertApply("test", req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("FlinkV12ClusterApply success")
}

// TEST FlinkV12ClusterDelete
func TestFlinkV12ClusterDelete(t *testing.T) {
	err := k8s.FlinkV12ClusterDelete("test", model.DeleteFlinkClusterRequest{
		ClusterName: tea.String("app-session"),
		NameSpace:   tea.String("flink"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("FlinkV12ClusterDelete success")
}
