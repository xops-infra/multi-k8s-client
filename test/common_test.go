package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/joho/godotenv"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
	"github.com/xops-infra/multi-k8s-client/pkg/service"
)

var k8s model.K8SContract

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	k8s = service.NewK8SService(model.K8SConfig{
		Clusters: map[string]model.Cluster{
			"test": {
				KubeConfig: tea.String(os.Getenv("KUBECONFIG")),
			},
		},
	})
}

// TEST CrdFlinkDeploymentApply
func TestCrdFlinkDeploymentApply(t *testing.T) {

	req := model.CreateFlinkClusterRequest{
		K8SClusterName: tea.String("test"),
		ClusterName:    tea.String("flink-application-cluster"),
		Image:          tea.String("flink:1.17"),
		Job: &model.Job{
			Parallelism: tea.Int32(4),
			JarURI:      tea.String("local:///opt/flink/examples/streaming/StateMachineExample.jar"),
		},
	}
	resp, err := k8s.CrdFlinkDeploymentApply(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("resp: %v", resp)
}

// TEST CrdFlinkDeploymentDelete
func TestCrdFlinkDeploymentDelete(t *testing.T) {

	req := model.DeleteFlinkClusterRequest{
		K8SClusterName: tea.String("test"),
		Name:           tea.String("flink-application-cluster"),
	}
	err := k8s.CrdFlinkDeploymentDelete(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func createSessionFlinkCluster() error {
	// Create session cluster first
	req := model.CreateFlinkClusterRequest{
		K8SClusterName: tea.String("test"),
		ClusterName:    tea.String("session-cluster"),
		Image:          tea.String("flink:1.17"),
	}
	resp, err := k8s.CrdFlinkDeploymentApply(req)
	if err != nil {
		return err
	}
	fmt.Printf("resp: %v", resp)
	return nil
}

// TEST CrdFlinkSessionJobSubmit
func TestCrdFlinkSessionJobSubmit(t *testing.T) {
	// createSessionFlinkCluster()
	// Submit session job
	sessionJobReq := model.CreateFlinkSessionJobRequest{
		K8SClusterName: tea.String("test"),
		JobName:        tea.String("example-job-1"),
		ClusterName:    tea.String("session-cluster"),
		Job: &model.Job{
			JarURI:      tea.String("https://repo1.maven.org/maven2/org/apache/flink/flink-examples-streaming_2.12/1.16.1/flink-examples-streaming_2.12-1.16.1-TopSpeedWindowing.jar"),
			Parallelism: tea.Int32(2),
			UpgradeMode: tea.String("stateless"),
		},
	}
	sessionJobResp, sessionJobErr := k8s.CrdFlinkSessionJobSubmit(sessionJobReq)
	if sessionJobErr != nil {
		t.Fatal(sessionJobErr)
	}
	t.Log("Submitted session job successfully", tea.Prettify(sessionJobResp))
}

// TEST CrdFlinkSessionJobDelete
func TestCrdFlinkSessionJobDelete(t *testing.T) {

	req := model.DeleteFlinkSessionJobRequest{
		K8SClusterName: tea.String("test"),
		ClusterName:    tea.String("session-cluster"),
		JobName:        tea.String("example-job"),
	}
	err := k8s.CrdFlinkSessionJobDelete(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
