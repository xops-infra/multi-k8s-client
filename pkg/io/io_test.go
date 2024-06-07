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

// CrdFlinkDeploymentList
func TestCrdFlinkDeploymentList(t *testing.T) {
	resp, err := client.CrdFlinkDeploymentList(model.Filter{
		NameSpace: tea.String("flink"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("List FlinkDeployment success", tea.Prettify(resp.Items))

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
		SubmitJobName: tea.String("test-job"),
		ClusterName:   tea.String("session-cluster"),
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

// CrdSparkApplicationList
func TestCrdSparkApplicationList(t *testing.T) {
	resp, err := client.CrdSparkApplicationList(model.Filter{
		// NameSpace: tea.String("default"),
		// FieldSelector: tea.String("metadata.name=spark-pi-example"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("List SparkApplication success", tea.Prettify(resp))
}

// CrdSparkApplicationApply
func TestCrdSparkApplicationApply(t *testing.T) {
	req := model.CreateSparkApplicationRequest{
		Name: tea.String("spark-pi-example"),
	}
	resp, err := client.CrdSparkApplicationApply(req.ToYaml())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Start SparkApplication success", resp)
}

// CrdSparkApplicationDelete
func TestCrdSparkApplicationDelete(t *testing.T) {
	err := client.CrdSparkApplicationDelete("default", "spark-pi-example")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Delete SparkApplication success")
}

// PvcList
func TestPvcList(t *testing.T) {
	resp, err := client.PvcList(model.Filter{
		NameSpace: tea.String("default"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range resp.Items {
		i.ManagedFields = nil
		t.Log(tea.Prettify(i))
	}
	t.Logf("List Pvc success %d", len(resp.Items))
}

// PvcApply
func TestPvcApply(t *testing.T) {
	req := model.ApplyPvcRequest{
		Name:        tea.String("demo-pvc"),
		Owner:       tea.String("demo-owner"),
		StorageSize: tea.Int(10),
	}

	resp, err := client.PvcApply(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("PvcApply success", resp)
}

// PvcDelete
func TestPvcDelete(t *testing.T) {
	err := client.PvcDelete("default", "demo-pvc")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("PvcDelete success")
}
