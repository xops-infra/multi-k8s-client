package test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

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
				KubeConfig: tea.String(os.Getenv("KUBE_CONFIG")),
			},
		},
	})
}

// TEST GetK8SCluster
func TestGetK8SCluster(t *testing.T) {
	clusterNames, err := k8s.GetK8SCluster()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("clusterNames: %v", clusterNames)
}

// TEST CrdFlinkDeploymentGet
func TestCrdFlinkDeploymentGet(t *testing.T) {

	resp, err := k8s.CrdFlinkDeploymentList("test", model.Filter{
		// FieldSelector: tea.String("metadata.name=flink-session-17,metadata.namespace=default"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(tea.Prettify(resp))
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// TEST CrdFlinkDeploymentApply
func TestCrdFlinkDeploymentApply(t *testing.T) {

	req := model.CreateFlinkClusterRequest{
		K8SClusterName: tea.String("test"),
		ClusterName:    tea.String("flink-application-13-" + generateRandomString(6)),
		Image:          tea.String("flink:1.13"),
		Version:        tea.String("v1_13"),
		Creater:        tea.String("xops"),
		FlinkConfiguration: map[string]any{
			"taskmanager.numberOfTaskSlots": "2",
			"state.savepoints.dir":          "file:///opt/flink/flink-data/savepoints",
			"state.checkpoints.dir":         "file:///opt/flink/flink-data/checkpoints",
			"high-availability":             "org.apache.flink.kubernetes.highavailability.KubernetesHaServicesFactory",
			"high-availability.storageDir":  "file:///opt/flink/flink-data/ha",
			"classloader.resolve-order":     "parent-first",
			"fs.s3a.access.key":             "minio",
			"fs.s3a.secret.key":             "minio123",
		},
		Job: &model.Job{
			Args:        []string{"-p", "4"},
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
		ClusterName:    tea.String("flink-application-cluster"),
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
		ClusterName:    tea.String("flink-session"),
		Image:          tea.String("flink:1.17"),
	}
	resp, err := k8s.CrdFlinkDeploymentApply(req)
	if err != nil {
		return err
	}
	fmt.Printf("resp: %v", resp)
	return nil
}

// TEST CrdFlinkSessionJobGet
func TestCrdFlinkSessionJobGet(t *testing.T) {
	resp, err := k8s.CrdFlinkSessionJobList("test", model.Filter{
		LabelSelector: tea.String("target.session=flink-session"),
		FieldSelector: tea.String("metadata.name=flink-session-job-3"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(tea.Prettify(resp))
}

// TEST CrdFlinkSessionJobSubmit
func TestCrdFlinkSessionJobSubmit(t *testing.T) {
	// createSessionFlinkCluster()
	// Submit session job
	sessionJobReq := model.CreateFlinkSessionJobRequest{
		K8SClusterName: tea.String("test"),
		NameSpace:      tea.String("default"),
		SubmitJobName:  tea.String("flink-session-job-4"),
		ClusterName:    tea.String("flink-session"),
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
		ClusterName:    tea.String("flink-session"),
		JobName:        tea.String("flink-session-job-1"),
	}
	err := k8s.CrdFlinkSessionJobDelete(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
