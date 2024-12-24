package service_test

import (
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
)

// TEST FlinkV12ClusterList
func TestFlinkV12ClusterList(t *testing.T) {
	resp, err := k8s.FlinkV12ClusterList("test", model.FilterFlinkV12{
		NameSpace: tea.String("flink"),
		// Name:      tea.String("search-vector"),
		Owner: tea.String("dingyingjie"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range resp.Items {
		t.Log(item.Info.GetVersion())
		t.Log(item.Info.GetImages())
		t.Log(item.Info.GetCreateTime())
		t.Log(item.Info.GetReplicas())
		t.Log(item.Info.GetRunTime())
	}

	t.Log(tea.Prettify(resp.Total))
}

// TEST FlinkV12ClustertApply
func TestFlinkV12ClustertApply(t *testing.T) {
	req := model.CreateFlinkV12ClusterRequest{
		Name:      tea.String("app-session"),
		NameSpace: tea.String("flink"),
		Owner:     tea.String("zhoushoujian"),
		Image:     tea.String("flink:1.12.7"),
		Env:       map[string]string{"ENABLE_BUILT_IN_PLUGINS": "flink-s3-fs-hadoop-1.12.7.jar;flink-s3-fs-presto-1.12.7.jar"},
		LoadBalancer: &model.LoadBalancerRequest{
			Annotations: map[string]string{"service.kubernetes.io/tke-existed-lbid": "lb-xxx"},
			Labels:      map[string]string{"used-by": "ui"},
		},
		JobManager: &model.JobManagerV12{
			PvcSize:      tea.Int(22),
			Resource:     &model.FlinkResource{Memory: tea.String("1Gi"), CPU: tea.String("1")},
			NodeSelector: &map[string]string{"kubernetes.io/os": "linux"},
		},
		TaskManager: &model.TaskManagerV12{
			Nu:           tea.Int(2),
			Resource:     &model.FlinkResource{Memory: tea.String("2Gi"), CPU: tea.String("1")},
			NodeSelector: &map[string]string{"kubernetes.io/os": "linux"},
		},
		FlinkConfigRequest: map[string]any{"taskmanager.numberOfTaskSlots": 2},
	}

	resp, err := k8s.FlinkV12ClustertApply("test", req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("FlinkV12ClusterApply success", tea.Prettify(resp))
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
