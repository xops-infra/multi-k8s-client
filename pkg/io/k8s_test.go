package io_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/xops-infra/multi-k8s-client/pkg/io"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/test/integration/fixtures"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientset *kubernetes.Clientset
var dynamicClient dynamic.Interface

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

func TestServerUp(t *testing.T) {
	tearDown, _, _, err := fixtures.StartDefaultServerWithClients(t)
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()
}

func TestCrdDeployment(t *testing.T) {
	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "demo-deployment",
			},
			"spec": map[string]interface{}{
				"replicas": 2,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "demo",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "demo",
						},
					},

					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"name":  "web",
								"image": "nginx:1.12",
								"ports": []map[string]interface{}{
									{
										"name":          "http",
										"protocol":      "TCP",
										"containerPort": 80,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := dynamicClient.Resource(deploymentRes).Namespace(apiv1.NamespaceDefault).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetName())
}

func TestCrd(t *testing.T) {
	flinkDeploymentRes := io.GetGVR("flink.apache.org", "v1beta1", "flinkdeployments")

	flinkDeployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "flink.apache.org/v1beta1",
			"kind":       "FlinkDeployment",
			"metadata": map[string]interface{}{
				"name": "basic-example-session-1",
			},
			"spec": map[string]interface{}{
				"image":        "flink:1.17",
				"flinkVersion": "v1_17",
				"flinkConfiguration": map[string]interface{}{
					"taskmanager.numberOfTaskSlots": "2",
				},
				"serviceAccount": "flink",
				"jobManager": map[string]interface{}{
					"resource": map[string]interface{}{
						"memory": "2048m",
						"cpu":    1,
					},
				},
				"taskManager": map[string]interface{}{
					"resource": map[string]interface{}{
						"memory": "2048m",
						"cpu":    1,
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating FlinkDeployment...")
	result, err := dynamicClient.Resource(flinkDeploymentRes).Create(context.TODO(), flinkDeployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created FlinkDeployment %q.\n", result.GetName())
}
