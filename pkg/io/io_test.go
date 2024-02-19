package io_test

import (
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/joho/godotenv"
	"github.com/xops-infra/multi-k8s-client/pkg/io"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
)

var client model.K8SIO

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	client, err = io.NewK8SClient(model.Cluster{
		KubePath: tea.String("~/.kube/config"),
	})
	if err != nil {
		panic(err)
	}

}

func TestK8SPod(t *testing.T) {

	var podName string
	{
		// List
		pods, err := client.PodList("default")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("List Pod success", len(pods.Items))
		if len(pods.Items) == 0 {
			t.Fatal("Pod not found")
			return
		}
		podName = pods.Items[0].Name
	}
	{
		// Get
		pod, err := client.PodGet("default", podName)
		if err != nil {
			t.Fatal(err)
		}
		// fmt.Println(tea.Prettify(pod))
		t.Log("Get Pod success", pod.Name)
	}
}

// TestK8SRbac
func TestK8SRbac(t *testing.T) {
	{
		// Rbac
		roles, err := client.RbacList("default")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("List Rbac success", len(roles.Items))
		for i := range roles.Items {
			t.Log(roles.Items[i].Name)
		}
	}
}

// FlinkApplication Create
func TestCrdFlinkDeploymentApplyApplication(t *testing.T) {
	req := model.CreateFlinkClusterRequest{
		ClusterName: tea.String("application-cluster"),
		Job: &model.Job{
			JarURI:      tea.String("local:///opt/flink/examples/streaming/StateMachineExample.jar"),
			Parallelism: tea.Int32(2),
			UpgradeMode: tea.String("stateless"),
		},
	}

	resp, err := client.CrdFlinkDeploymentApply("", req.ToYaml())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Start FlinkApplication success", resp)
}

// FlinkSession Create
func TestCrdFlinkDeploymentApplySession(t *testing.T) {
	req := model.CreateFlinkClusterRequest{
		ClusterName: tea.String("session-cluster"),
	}

	resp, err := client.CrdFlinkDeploymentApply("", req.ToYaml())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Start FlinkSession success", resp)
	// kubectl port-forward svc/session-cluster-rest 8081
}

// FlinkSessionJob
func TestCrdFlinkSessionJobSubmit(t *testing.T) {
	req := model.CreateFlinkSessionJobRequest{
		JobName:     tea.String("test-job"),
		ClusterName: tea.String("session-cluster"),
		Job: &model.Job{
			JarURI:      tea.String("https://repo1.maven.org/maven2/org/apache/flink/flink-examples-streaming_2.12/1.16.1/flink-examples-streaming_2.12-1.16.1-TopSpeedWindowing.jar"),
			Parallelism: tea.Int32(2),
			UpgradeMode: tea.String("stateless"),
		},
	}
	resp, err := client.CrdFlinkSessionJobSubmit("", req.ToYaml())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("CrdFlinkSessionJobSubmit success", resp)
}

// CrdFlinkSessionJobDelete
func TestCrdFlinkSessionJobDelete(t *testing.T) {
	cluster := []string{"test-job"}
	for _, i := range cluster {
		err := client.CrdFlinkSessionJobDelete("", i)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("CrdFlinkSessionJobDelete %s success", i)
	}
}

// FlinkDeployment
func TestCrdFlinkDeploymentDelete(t *testing.T) {
	cluster := []string{"session-cluster"}
	for _, i := range cluster {
		err := client.CrdFlinkDeploymentDelete("", i)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("CrdFlinkDeploymentDelete %s success", i)
	}
}
