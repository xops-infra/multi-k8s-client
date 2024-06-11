package model_test

import (
	"fmt"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
)

var req = model.CreateFlinkV12ClusterRequest{
	Name:  tea.String("test-cluster"),
	Owner: tea.String("test-owner"),
	// Image: tea.String("flink:1.12.7"),
	Env: map[string]string{"ENABLE_BUILT_IN_PLUGINS": "flink-s3-fs-hadoop-1.12.7.jar;flink-s3-fs-presto-1.12.7.jar"},
	LoadBalancer: &model.LoadBalancerRequest{
		Annotations: map[string]string{"kubernetes.io/ingress.class": "nginx"},
		Labels:      map[string]string{"app": "flink"},
	},
	JobManager: &model.JobManagerV12{
		PvcSize:  tea.Int(22),
		Resource: &model.FlinkResource{Memory: tea.String("1024m"), CPU: tea.String("1")},
	},
	TaskManager: &model.TaskManagerV12{
		Nu:       tea.Int(10),
		Resource: &model.FlinkResource{Memory: tea.String("2048m"), CPU: tea.String("1")},
	},
	FlinkConfigRequest: map[string]any{"taskmanager.numberOfTaskSlots": 2},
	NodeSelector:       map[string]any{"kubernetes.io/os": "linux"},
}

// TestNewPVC
func TestNewPVC(t *testing.T) {
	pvc := req.NewPVC()
	fmt.Println(tea.Prettify(pvc))
}

// TestNewService
func TestNewService(t *testing.T) {
	service := req.NewService()
	fmt.Println(tea.Prettify(service))
}

// NewLBService
func TestNewLBService(t *testing.T) {
	service := req.NewLBService()
	fmt.Println(tea.Prettify(service))
}

// NewConfigMap
func TestNewConfigMap(t *testing.T) {
	configMap := req.NewConfigMap()
	fmt.Println(tea.Prettify(configMap))
}

// NewJobManagerDeployment
func TestNewJobManagerDeployment(t *testing.T) {
	deployment := req.NewJobManagerDeployment()
	fmt.Println(tea.Prettify(deployment))
	createDeploymentRequest, err := model.NewDeploymentCreateFromMap(deployment)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	fmt.Println(tea.Prettify(createDeploymentRequest))
}

// NewTaskManagerDeployment
func TestNewTaskManagerDeployment(t *testing.T) {
	deployment := req.NewTaskManagerDeployment()
	createDeploymentRequest, err := model.NewDeploymentCreateFromMap(deployment)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	fmt.Println(tea.Prettify(createDeploymentRequest))
}
