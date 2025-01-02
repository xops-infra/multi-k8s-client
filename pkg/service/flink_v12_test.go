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
		Name:      tea.String("zsj-session-flink"),
		// Owner: tea.String("dingyingjie"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range resp.Items {
		t.Log(item.ClusterName)
		// t.Log(item.Info.GetImages())
		// t.Log(item.Info.GetCreateTime())
		// t.Log(item.Info.GetReplicas())
		t.Log(item.Info.GetResourcesLimitGb())
		t.Log(item.Info.GetResourcesRequestGb())
		t.Log(item.Info.GetRunTime())
		t.Log(item.FlinkConfig)
	}

	t.Log(tea.Prettify(resp.Total))
}

// TEST FlinkV12ClusterCreate
func TestFlinkV12ClusterCreate(t *testing.T) {
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

	resp, err := k8s.FlinkV12ClusterCreate("test", req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("FlinkV12ClusterCreate success", tea.Prettify(resp))
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

func TestFlinkV12ClusterApply(t *testing.T) {
	err := k8s.FlinkV12ClusterApply("test", "flink", "flink-zhoushoujian", model.ApplyFlinkV12ClusterRequest{
		Labels: map[string]string{"owner": "zhoushoujian"},
		FlinkConfiguration: map[string]any{
			"jobmanager.memory.flink.size":         "3072m",
			"jobmanager.memory.jvm-metaspace.size": "1024m",
			"taskmanager.numberOfTaskSlots":        4,
			"taskmanager.memory.managed.size":      "24m",
			"taskmanager.memory.process.size":      "3200m",
			"taskmanager.memory.network.min":       "100MB",
			"taskmanager.memory.network.max":       "301MB",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
