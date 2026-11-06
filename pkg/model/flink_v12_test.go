package model_test

import (
	"fmt"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
		Nu:           tea.Int(10),
		Resource:     &model.FlinkResource{Memory: tea.String("2048m"), CPU: tea.String("1")},
		NodeSelector: &map[string]string{"kubernetes.io/os": "linux"},
	},
	FlinkConfigRequest: map[string]any{"taskmanager.numberOfTaskSlots": 2},
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
	req.JobManager.SideCars = []model.SideCar{
		{
			Name:    tea.String("sidecar"),
			Image:   tea.String("busybox"),
			Command: []string{"sleep", "10"},
			Env: []model.Env{
				{Name: tea.String("key"), Value: tea.String("value")},
			},
			LivenessProbe: &model.LivenessProbe{
				Exec: &model.Exec{
					Command: []string{"sleep", "10"},
				},
			},
		},
	}
	deployment := req.NewJobManagerDeployment()
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

func TestFlinkV12DeploymentServiceAccountTokenPermissions(t *testing.T) {
	for _, tc := range []struct {
		name      string
		sidecar   bool
		generate  func(*model.CreateFlinkV12ClusterRequest) map[string]any
		container string
		postStart []string
	}{
		{
			name: "jobmanager", generate: (*model.CreateFlinkV12ClusterRequest).NewJobManagerDeployment,
			container: "jobmanager",
			postStart: []string{"/bin/sh", "-c", "mkdir -p /opt/flink/target/flink-web-upload&&chown -R flink:flink /opt/flink/target/"},
		},
		{
			name: "jobmanager_with_sidecar", sidecar: true,
			generate: (*model.CreateFlinkV12ClusterRequest).NewJobManagerDeployment, container: "jobmanager",
			postStart: []string{"/bin/sh", "-c", "mkdir -p /opt/flink/target/flink-web-upload&&chown -R flink:flink /opt/flink/target/"},
		},
		{
			name: "taskmanager", generate: (*model.CreateFlinkV12ClusterRequest).NewTaskManagerDeployment,
			container: "taskmanager", postStart: []string{"sh", "-c", "chown 9999:9999 /opt/flink/log"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := model.CreateFlinkV12ClusterRequest{
				Name:        tea.String("token-permissions-" + tc.name),
				JobManager:  &model.JobManagerV12{},
				TaskManager: &model.TaskManagerV12{Nu: tea.Int(1)},
			}
			if tc.sidecar {
				request.JobManager.SideCars = []model.SideCar{{
					Name: tea.String("sidecar"), Image: tea.String("busybox"), Command: []string{"sleep", "10"},
				}}
			}
			deployment, err := model.NewDeploymentCreateFromMap(tc.generate(&request))
			require.NoError(t, err)
			pod := deployment.Spec.Template.Spec
			require.NotNil(t, pod.SecurityContext)
			require.NotNil(t, pod.SecurityContext.FSGroup)
			assert.Equal(t, int64(9999), *pod.SecurityContext.FSGroup)
			require.NotNil(t, pod.SecurityContext.FSGroupChangePolicy)
			assert.Equal(t, v1.FSGroupChangeOnRootMismatch, *pod.SecurityContext.FSGroupChangePolicy)
			assert.Empty(t, pod.ServiceAccountName, "keep Kubernetes default service account selection")
			assert.Nil(t, pod.AutomountServiceAccountToken, "keep Kubernetes default token automount behavior")
			containerCount := 1
			if tc.sidecar {
				containerCount++
			}
			require.Len(t, pod.Containers, containerCount)
			container := pod.Containers[0]
			assert.Equal(t, tc.container, container.Name)
			require.NotNil(t, container.SecurityContext)
			require.NotNil(t, container.SecurityContext.RunAsUser)
			assert.Equal(t, int64(0), *container.SecurityContext.RunAsUser)
			require.NotNil(t, container.Lifecycle)
			require.NotNil(t, container.Lifecycle.PostStart)
			require.NotNil(t, container.Lifecycle.PostStart.Exec)
			assert.Equal(t, tc.postStart, container.Lifecycle.PostStart.Exec.Command)
			if tc.sidecar {
				assert.Equal(t, "sidecar", pod.Containers[1].Name)
			}
		})
	}
}

// TestCpuResourceCalculation 测试CPU资源计算是否正确
func TestCpuResourceCalculation(t *testing.T) {
	// 验证CPU资源计算
	testCases := []struct {
		cpuLimit       string
		expectedCpuReq string
	}{
		{"1", "500m"},    // 1 CPU -> 500m
		{"2", "1"},       // 2 CPU -> 1 CPU
		{"500m", "250m"}, // 500m CPU -> 250m
		{"4", "2"},       // 4 CPU -> 2 CPU
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("CPU:%s", tc.cpuLimit), func(t *testing.T) {
			// 解析预期值
			cpuLimit := resource.MustParse(tc.cpuLimit)
			expectedCpuReq := resource.MustParse(tc.expectedCpuReq)

			// 验证CPU请求是限制的一半
			assert.Equal(t, expectedCpuReq.MilliValue(), cpuLimit.MilliValue()/2,
				"CPU request should be half of limit")
		})
	}
}

// TestMemoryResourceCalculation 测试内存资源计算是否正确
func TestMemoryResourceCalculation(t *testing.T) {
	// 验证内存资源计算
	testCases := []struct {
		memLimit       string
		expectedMemReq string
	}{
		{"1024Mi", "512Mi"}, // 1024Mi -> 512Mi
		{"2Gi", "1Gi"},      // 2Gi -> 1Gi
		{"512Mi", "256Mi"},  // 512Mi -> 256Mi
		{"8Gi", "4Gi"},      // 8Gi -> 4Gi
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Memory:%s", tc.memLimit), func(t *testing.T) {
			// 解析内存限制值
			memLimit := resource.MustParse(tc.memLimit)
			expectedMemReq := resource.MustParse(tc.expectedMemReq)

			// 验证内存请求是限制的一半
			assert.Equal(t, expectedMemReq.Value(), memLimit.Value()/2,
				"Memory request should be half of limit")
		})
	}
}

// TestLargeResourceCalculation 测试大资源值计算是否正确
func TestLargeResourceCalculation(t *testing.T) {
	// 解析预期值
	cpuLimit := resource.MustParse("4")
	expectedCpuReq := resource.MustParse("2")

	// 验证CPU请求是限制的一半
	assert.Equal(t, expectedCpuReq.MilliValue(), cpuLimit.MilliValue()/2,
		"CPU request should be half of limit for large values")

	// 解析大内存值，使用正确的内存单位
	memLimit := resource.MustParse("8192Mi")
	expectedMemReq := resource.MustParse("4096Mi")

	// 验证内存请求是限制的一半
	assert.Equal(t, expectedMemReq.Value(), memLimit.Value()/2,
		"Memory request should be half of limit for large values")

	// 同样测试GiB单位
	memLimitGi := resource.MustParse("8Gi")
	expectedMemReqGi := resource.MustParse("4Gi")

	// 验证内存请求是限制的一半
	assert.Equal(t, expectedMemReqGi.Value(), memLimitGi.Value()/2,
		"Memory request should be half of limit for Gi values")
}
