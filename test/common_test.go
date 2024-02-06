package test

import (
	"testing"

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
	kubeConfig := "~/.kube/config"
	client, err = io.NewK8SClient(kubeConfig)
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
