package io_test

import (
	"strings"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/joho/godotenv"
	"github.com/xops-infra/multi-k8s-client/pkg/io"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var client model.K8SIO

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	kubeConfig := "~/.kube/config"
	client, err = io.NewK8SClient(kubeConfig)
	if err != nil {
		panic(err)
	}

	if strings.HasPrefix(kubeConfig, "~/") {
		kubeConfig = strings.Replace(kubeConfig, "~/", homedir.HomeDir()+"/", 1)
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err)
	}
	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	dynamicClient, err = dynamic.NewForConfig(config)
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
	client, err := io.NewK8SClient("~/.kube/config")
	if err != nil {
		t.Fatal(err)
	}
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

func TestCrdApplyFlinkDeployment(t *testing.T) {
	req := model.CreateFlinkRequest{
		ClusterName: tea.String("application-cluster"),
		Job: &model.Job{
			JarURI:      tea.String("local:///opt/flink/examples/streaming/StateMachineExample.jar"),
			Parallelism: tea.Int32(2),
			UpgradeMode: tea.String("stateless"),
		},
	}

	resp, err := client.CrdApplyFlinkDeployment("", req.ToYaml())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Start FlinkApplication success", resp)
}

func TestApplyFlinkDeploymentSession(t *testing.T) {
	req := model.CreateFlinkRequest{
		ClusterName: tea.String("session-cluster"),
	}

	resp, err := client.CrdApplyFlinkDeployment("", req.ToYaml())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Start FlinkSession success", resp)
	// kubectl port-forward svc/session-cluster-rest 8081
}

func TestCrdSubmitFlinkSessionJob(t *testing.T) {
	req := model.FlinkSessionJobRequest{
		JobName:     tea.String("test-job"),
		ClusterName: tea.String("session-cluster"),
		Job: &model.Job{
			JarURI:      tea.String("https://repo1.maven.org/maven2/org/apache/flink/flink-examples-streaming_2.12/1.16.1/flink-examples-streaming_2.12-1.16.1-TopSpeedWindowing.jar"),
			Parallelism: tea.Int32(2),
			UpgradeMode: tea.String("stateless"),
		},
	}
	resp, err := client.CrdSubmitFlinkSessionJob("", req.ToYaml())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("CrdSubmitFlinkSessionJob success", resp)
}

// FlinkDeployment
func TestCrdDeleteFlinkDeployment(t *testing.T) {
	cluster := []string{"application-1", "application-session-1"}
	for _, i := range cluster {
		err := client.CrdDeleteFlinkDeployment("", i)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("CrdDeleteFlinkDeployment %s success", i)
	}
}
