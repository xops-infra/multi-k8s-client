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

// TEST FlinkV12ClustertApply
func TestFlinkV12ClustertApply(t *testing.T) {
	req := model.CreateFlinkV12ClusterRequest{
		Name:      tea.String("app-session"),
		NameSpace: tea.String("flink"),
		Owner:     tea.String("zhangsan"),
		Image:     tea.String("flink:1.12.7"),
		Env:       map[string]string{"ENABLE_BUILT_IN_PLUGINS": "flink-s3-fs-hadoop-1.12.7.jar;flink-s3-fs-presto-1.12.7.jar"},
		LoadBalancer: &model.LoadBalancerRequest{
			Annotations: map[string]string{"kubernetes.io/ingress.class": "nginx"},
			Labels:      map[string]string{"app": "flink"},
		},
		JobManager: &model.JobManagerV12{
			PvcSize: tea.Int(22),
			// Resource: &model.FlinkResource{Memory: tea.String("1024m"), CPU: tea.String("1")},
		},
		TaskManager: &model.TaskManagerV12{
			Nu: tea.Int(2),
			// Resource: &model.FlinkResource{Memory: tea.String("2048m"), CPU: tea.String("1")},
		},
		FlinkConfigRequest: map[string]any{"taskmanager.numberOfTaskSlots": 2},
		NodeSelector:       map[string]any{"kubernetes.io/os": "linux"},
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
