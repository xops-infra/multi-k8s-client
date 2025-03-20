package model_test

import (
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/assert"
	"github.com/xops-infra/multi-k8s-client/pkg/model"
)

func TestCreateFlinkClusterValidate(t *testing.T) {
	tests := []struct {
		name        string
		request     *model.CreateFlinkClusterRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "有效配置",
			request: &model.CreateFlinkClusterRequest{
				ClusterName: tea.String("test-cluster"),
				Submitter:   tea.String("admin"),
				JobManager: &model.Manager{
					Resource: &model.FlinkResource{
						Memory: tea.String("2048Mi"),
						CPU:    tea.String("1"),
					},
				},
				TaskManager: &model.Manager{
					Resource: &model.FlinkResource{
						Memory: tea.String("4096Mi"),
						CPU:    tea.String("2"),
					},
				},
			},
			expectError: false,
		},
		{
			name: "集群名称缺失",
			request: &model.CreateFlinkClusterRequest{
				Submitter: tea.String("admin"),
			},
			expectError: true,
			errorMsg:    "cluster_name is required",
		},
		{
			name: "提交者缺失",
			request: &model.CreateFlinkClusterRequest{
				ClusterName: tea.String("test-cluster"),
			},
			expectError: true,
			errorMsg:    "submitter is required",
		},
		{
			name: "无效的集群名称格式",
			request: &model.CreateFlinkClusterRequest{
				ClusterName: tea.String("Test_Cluster"),
				Submitter:   tea.String("admin"),
			},
			expectError: true,
			errorMsg:    "cluster_name must be a valid kubernetes name",
		},
		{
			name: "JobManager内存单位无效",
			request: &model.CreateFlinkClusterRequest{
				ClusterName: tea.String("test-cluster"),
				Submitter:   tea.String("admin"),
				JobManager: &model.Manager{
					Resource: &model.FlinkResource{
						Memory: tea.String("2048m"),
						CPU:    tea.String("1"),
					},
				},
			},
			expectError: true,
			errorMsg:    "job_manager.resource.memory must use Mi or Gi unit",
		},
		{
			name: "TaskManager内存单位无效",
			request: &model.CreateFlinkClusterRequest{
				ClusterName: tea.String("test-cluster"),
				Submitter:   tea.String("admin"),
				JobManager: &model.Manager{
					Resource: &model.FlinkResource{
						Memory: tea.String("2048Mi"),
						CPU:    tea.String("1"),
					},
				},
				TaskManager: &model.Manager{
					Resource: &model.FlinkResource{
						Memory: tea.String("4096m"),
						CPU:    tea.String("2"),
					},
				},
			},
			expectError: true,
			errorMsg:    "task_manager.resource.memory must use Mi or Gi unit",
		},
		{
			name: "无效的CPU格式",
			request: &model.CreateFlinkClusterRequest{
				ClusterName: tea.String("test-cluster"),
				Submitter:   tea.String("admin"),
				JobManager: &model.Manager{
					Resource: &model.FlinkResource{
						Memory: tea.String("2048Mi"),
						CPU:    tea.String("1x"),
					},
				},
			},
			expectError: true,
			errorMsg:    "job_manager.resource.cpu must be a valid format",
		},
		{
			name: "Job配置缺少JarURI",
			request: &model.CreateFlinkClusterRequest{
				ClusterName: tea.String("test-cluster"),
				Submitter:   tea.String("admin"),
				Job:         &model.Job{},
			},
			expectError: true,
			errorMsg:    "job.jar_url is required when job is provided",
		},
		{
			name: "无效的JAR URL格式",
			request: &model.CreateFlinkClusterRequest{
				ClusterName: tea.String("test-cluster"),
				Submitter:   tea.String("admin"),
				Job: &model.Job{
					JarURI: tea.String("file:///path/to/jar"),
				},
			},
			expectError: true,
			errorMsg:    "job.jar_url must start with",
		},
		{
			name: "无效的升级模式",
			request: &model.CreateFlinkClusterRequest{
				ClusterName: tea.String("test-cluster"),
				Submitter:   tea.String("admin"),
				Job: &model.Job{
					JarURI:      tea.String("local:///path/to/jar"),
					UpgradeMode: tea.String("invalid-mode"),
				},
			},
			expectError: true,
			errorMsg:    "job.upgrade_mode must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
