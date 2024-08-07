package service_test

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

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
	_k8s, err := service.NewK8SService([]model.Cluster{
		{
			KubePath: tea.String("~/.kube/config"),
			Name:     tea.String("test"),       // 集群名称
			Alias:    tea.String("test_alias"), // 集群别名
		}, {
			KubePath: tea.String("~/.kube/config"),
			Name:     tea.String("testa"),
			Alias:    tea.String("test_alias_a"),
		},
	})
	if err != nil {
		panic(err)
	}

	k8s = _k8s
}

// TEST GetK8SCluster
func TestGetK8SCluster(t *testing.T) {
	clusterNames := k8s.GetK8SCluster()
	t.Logf("clusterNames: %v", tea.Prettify(clusterNames))
}

// TEST CrdFlinkDeploymentGet
func TestCrdFlinkDeploymentGet(t *testing.T) {

	resp, err := k8s.CrdFlinkDeploymentList("test", model.Filter{
		NameSpace: tea.String("flink"),
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
	s3Path := fmt.Sprintf("%s/%s/%s/flink-data/", os.Getenv("S3_BUCKET"), "test", time.Now().Format("20060102150405"))
	req := model.CreateFlinkClusterRequest{
		ClusterName: tea.String("flink-application-13-" + generateRandomString(6)),
		Image:       tea.String("flink:1.13"),
		Version:     tea.String("v1_13"),
		Submitter:   tea.String("xops"),
		NameSpace:   tea.String("flink"),
		Env: []model.Env{
			{
				Name:  tea.String("ENABLE_BUILT_IN_PLUGINS"),
				Value: tea.String("flink-s3-fs-hadoop-1.13.6.jar;flink-s3-fs-presto-1.13.6.jar"),
			},
		},
		TaskManager: &model.Manager{
			NodeSelector: &map[string]string{"env": "flink"},
			Resource:     &model.FlinkResource{Memory: tea.String("1024m"), CPU: tea.String("2")},
		},
		JobManager: &model.Manager{
			NodeSelector: &map[string]string{"env": "flink"},
			Resource:     &model.FlinkResource{Memory: tea.String("2048m"), CPU: tea.String("1")},
		},
		FlinkConfiguration: map[string]any{
			"taskmanager.numberOfTaskSlots":          "2",
			"jobmanager.execution.failover-strategy": "region",
			"state.savepoints.dir":                   fmt.Sprintf("s3a://%s/savepoints", strings.TrimSuffix(s3Path, "/")),
			"state.checkpoints.dir":                  fmt.Sprintf("s3a://%s/checkpoints", strings.TrimSuffix(s3Path, "/")),
			"high-availability":                      "org.apache.flink.kubernetes.highavailability.KubernetesHaServicesFactory",
			"high-availability.storageDir":           fmt.Sprintf("s3a://%s/ha", strings.TrimSuffix(s3Path, "/")),
			"classloader.resolve-order":              "parent-first",
			"fs.s3a.access.key":                      os.Getenv("AWS_ACCESS_KEY_ID"),
			"fs.s3a.secret.key":                      os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"fs.s3a.endpoint":                        "cos.ap-shanghai.myqcloud.com",
		},
		Job: &model.Job{
			Parallelism: tea.Int32(2),
			JarURI:      tea.String("local:///opt/flink/examples/streaming/StateMachineExample.jar"),
		},
	}
	resp, err := k8s.CrdFlinkDeploymentApply("dev", req)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("resp: %v", resp)
}

// TEST CrdFlinkDeploymentDelete
func TestCrdFlinkDeploymentDelete(t *testing.T) {

	req := model.DeleteFlinkClusterRequest{
		ClusterName: tea.String("flink-application-13-csdtr8"),
	}
	err := k8s.CrdFlinkDeploymentDelete("dev", req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func createSessionFlinkCluster() error {
	// Create session cluster first
	req := model.CreateFlinkClusterRequest{
		ClusterName: tea.String("flink-session"),
		Image:       tea.String("flink:1.17"),
	}
	resp, err := k8s.CrdFlinkDeploymentApply("test", req)
	if err != nil {
		return err
	}
	fmt.Printf("resp: %v", resp)
	return nil
}

// TEST CrdFlinkSessionJobGet
func TestCrdFlinkSessionJobGet(t *testing.T) {
	resp, err := k8s.CrdFlinkSessionJobList("dev", model.Filter{
		NameSpace: tea.String("flink"),
		// LabelSelector: tea.String("target.session=flink-session"),
		// FieldSelector: tea.String("metadata.name=flink-session-job-3"),
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
		NameSpace:     tea.String("default"),
		SubmitJobName: tea.String("flink-session-job-13"),
		ClusterName:   tea.String("flink-session-13"),
		Job: &model.Job{
			JarURI:      tea.String("https://repo1.maven.org/maven2/org/apache/flink/flink-examples-streaming_2.12/1.13.6/flink-examples-streaming_2.12-1.13.6-TopSpeedWindowing.jar"),
			Parallelism: tea.Int32(2),
			UpgradeMode: tea.String("stateless"),
		},
		Submitter: tea.String("xops"),
	}
	sessionJobResp, sessionJobErr := k8s.CrdFlinkSessionJobSubmit("test", sessionJobReq)
	if sessionJobErr != nil {
		t.Fatal(sessionJobErr)
	}
	t.Log("Submitted session job successfully", tea.Prettify(sessionJobResp))
}

// TEST CrdFlinkSessionJobDelete
func TestCrdFlinkSessionJobDelete(t *testing.T) {

	req := model.DeleteFlinkSessionJobRequest{
		ClusterName: tea.String("flink-session"),
		JobName:     tea.String("flink-session-job-4"),
	}
	err := k8s.CrdFlinkSessionJobDelete("test", req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

// CrdSparkApplicationList
func TestCrdSparkApplicationList(t *testing.T) {
	resp, err := k8s.CrdSparkApplicationList("test", model.Filter{
		// NameSpace: tea.String("default"),
		// FieldSelector: tea.String("metadata.name=spark-pi-example"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("List SparkApplication success", tea.Prettify(resp))
}

// CrdSparkApplicationGet
func TestCrdSparkApplicationGet(t *testing.T) {
	resp, err := k8s.CrdSparkApplicationGet("test", "default", "spark-pi-example")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Get SparkApplication success", tea.Prettify(resp.Spec))
}

// CrdFlinkDeploymentRestart
func TestCrdFlinkDeploymentRestart(t *testing.T) {

	req := model.RestartFlinkClusterRequest{
		ClusterName: tea.String("flink-sync"),
		NameSpace:   tea.String("flink"),
		Type:        model.FlinkTypeTM,
	}
	err := k8s.CrdFlinkDeploymentRestart("test", req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
