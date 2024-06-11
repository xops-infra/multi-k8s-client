package io_test

import (
	"fmt"
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
		Label:       map[string]string{"app": "demo-pvc"},
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

// ServiceList
func TestServiceList(t *testing.T) {
	resp, err := client.ServiceList(model.Filter{
		NameSpace: tea.String("default"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range resp.Items {
		i.ManagedFields = nil
		t.Log(tea.Prettify(i))
	}
	t.Logf("List Service success %d", len(resp.Items))
}

// ServiceApply
func TestServiceApply(t *testing.T) {
	req := model.ApplyServiceRequest{
		Name:      tea.String("demo-service"),
		Namespace: tea.String("default"),
		Label:     map[string]string{"app": "demo-service", "owner": "demo"},
		Spec: &model.ServiceSpec{
			Type:     tea.String("ClusterIP"),
			Selector: map[string]string{"app": "demo-service"},
			Ports: []model.Port{{Name: tea.String("http"),
				Port: tea.Int32(80), TargetPort: tea.Int32(8080),
				Protocol: tea.String("TCP")}},
		},
		Annotations: map[string]string{"owner": "demo"},
	}
	fmt.Println(tea.Prettify(req))
	resp, err := client.ServiceApply(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ServiceApply success", resp)
}

// ServiceDelete
func TestServiceDelete(t *testing.T) {
	err := client.ServiceDelete("default", "demo-service")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ServiceDelete success")
}

// ConfigMapList
func TestConfigmapList(t *testing.T) {
	resp, err := client.ConfigMapList(model.Filter{
		NameSpace: tea.String("default"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range resp.Items {
		i.ManagedFields = nil
		t.Log(tea.Prettify(i))
	}
	t.Logf("List Configmap success %d", len(resp.Items))
}

// ConfigMapApply
func TestConfigmapApply(t *testing.T) {
	req := model.ApplyConfigMapRequest{
		Namespace: tea.String("default"),
		Name:      tea.String("demo-configmap"),
		Labels:    map[string]string{"app": "demo-configmap", "owner": "demo"},
		Data: map[string]string{"abc.config": `asdasd
		dddd
		asd
		xxx
		`, "xxx.yaml": "value2"},
	}
	resp, err := client.ConfigMapApply(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ConfigmapApply success", resp)
}

// ConfigMapDelete
func TestConfigmapDelete(t *testing.T) {
	err := client.ConfigMapDelete("default", "demo-configmap")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ConfigmapDelete success")
}

// DeploymentList
func TestDeploymentList(t *testing.T) {
	resp, err := client.DeploymentList(model.Filter{
		NameSpace: tea.String("default"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range resp.Items {
		i.ManagedFields = nil
		t.Log(tea.Prettify(i))
	}
	t.Logf("List Deployment success %d", len(resp.Items))
}

// DeploymentCreate
func TestDeploymentCreate(t *testing.T) {
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
			PvcSize:  tea.Int(22),
			Resource: &model.FlinkResource{Memory: tea.String("1024m"), CPU: tea.String("1")},
		},
		TaskManager: &model.TaskManagerV12{
			Nu:       tea.Int(2),
			Resource: &model.FlinkResource{Memory: tea.String("2048m"), CPU: tea.String("1")},
		},
		FlinkConfigRequest: map[string]any{"taskmanager.numberOfTaskSlots": 2},
		NodeSelector:       map[string]any{"kubernetes.io/os": "linux"},
	}
	jobDy := req.NewJobManagerDeployment()
	// fmt.Println(tea.Prettify(dy))
	createDeploymentRequest, err := model.NewDeploymentCreateFromMap(jobDy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	fmt.Println(tea.Prettify(createDeploymentRequest))
	resp, err := client.DeploymentCreate(createDeploymentRequest)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("DeploymentCreate success", resp)
}

// DeploymentApply
func TestDeploymentApply(t *testing.T) {
}

// DeploymentDelete
func TestDeploymentDelete(t *testing.T) {
	err := client.DeploymentDelete("flink", "app-session-jobmanager")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("DeploymentDelete success")
}
