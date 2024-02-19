package model

import "github.com/alibabacloud-go/tea/tea"

type CrdFlinkSessionJobGetResponse struct {
	Total int                      `json:"total"`
	Items []CrdFlinkSessionJobItem `json:"items"`
}

type CrdFlinkSessionJobItem struct {
	ClusterName    string `json:"cluster_name"`
	NameSpace      string `json:"namespace"`
	SubmitJobName  string `json:"submit_job_name"` // 用户提交制定的job名称
	JobName        string `json:"job_name"`
	JobId          string `json:"job_id"`
	Job            any    `json:"job"`
	Status         string `json:"status"`
	LifecycleState string `json:"lifecycle_state"`
}

type CrdFlinkDeploymentGetResponse struct {
	Total int                  `json:"total"`
	Items []CrdFlinkDeployment `json:"items"`
}

type CrdFlinkDeployment struct {
	ClusterName string `json:"cluster_name"`
	NameSpace   string `json:"namespace"`
}

type CreateFlinkClusterRequest struct {
	K8SClusterName *string      `json:"k8s_cluster_name" binding:"required"` // 初始化的k8s集群名称
	NameSpace      *string      `json:"namespace" default:"default"`
	ClusterName    *string      `json:"cluster_name" binding:"required"` // metadata.name
	Image          *string      `json:"image" default:"flink:1.17"`
	Version        *string      `json:"version" default:"v1_17"`
	ServiceAccount *string      `json:"service_account" default:"flink"`
	TaskManager    *TaskManager `json:"task_manager"`
	JobManager     *JobManager  `json:"job_manager"`
	Job            *Job         `json:"job"` // 如果没有该字段则创建 Session集群，如果有该字段则创建Application集群。
}

/*
apiVersion: flink.apache.org/v1beta1
kind: FlinkDeployment
metadata:

	name: basic-example

spec:

	image: flink:1.17
	flinkVersion: v1_17
	flinkConfiguration:
	  taskmanager.numberOfTaskSlots: "2"
	serviceAccount: flink
	jobManager:
	  resource:
	    memory: "2048m"
	    cpu: 1
	taskManager:
	  resource:
	    memory: "2048m"
	    cpu: 1
	job:
	  jarURI: local:///opt/flink/examples/streaming/StateMachineExample.jar
	  parallelism: 2  # 2 task managers
	  upgradeMode: stateless
*/
func (req *CreateFlinkClusterRequest) ToYaml() map[string]any {
	yaml := map[string]interface{}{
		"apiVersion": "flink.apache.org/v1beta1",
		"kind":       "FlinkDeployment",
		"metadata": map[string]interface{}{
			"name": "basic-example",
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
	}
	if req.ClusterName != nil {
		yaml["metadata"].(map[string]interface{})["name"] = *req.ClusterName
	}
	if req.NameSpace != nil {
		yaml["metadata"].(map[string]interface{})["namespace"] = *req.NameSpace
	}
	if req.Image != nil {
		yaml["spec"].(map[string]interface{})["image"] = *req.Image
	}
	if req.Version != nil {
		yaml["spec"].(map[string]interface{})["flinkVersion"] = *req.Version
	}
	if req.TaskManager != nil && req.TaskManager.Resource != nil {
		yaml["spec"].(map[string]interface{})["taskManager"].(map[string]interface{})["resource"] = req.TaskManager.Resource
	}
	if req.TaskManager != nil && req.TaskManager.NumberOfTaskSlots != nil {
		yaml["spec"].(map[string]interface{})["flinkConfiguration"].(map[string]interface{})["taskmanager.numberOfTaskSlots"] = *req.TaskManager.NumberOfTaskSlots
	}
	if req.JobManager != nil && req.JobManager.Resource != nil {
		yaml["spec"].(map[string]interface{})["jobManager"].(map[string]interface{})["resource"] = req.JobManager.Resource
	}
	if req.Job != nil {
		yaml["spec"].(map[string]interface{})["job"] = map[string]interface{}{
			"jarURI":      "local:///opt/flink/examples/streaming/StateMachineExample.jar",
			"parallelism": 2,
			"upgradeMode": "stateless",
		}
		if req.Job.UpgradeMode == nil {
			yaml["spec"].(map[string]interface{})["job"].(map[string]interface{})["upgradeMode"] = "stateless"
		}
		if req.Job.JarURI != nil {
			yaml["spec"].(map[string]interface{})["job"].(map[string]interface{})["jarURI"] = *req.Job.JarURI
		}
		if req.Job.Parallelism != nil {
			yaml["spec"].(map[string]interface{})["job"].(map[string]interface{})["parallelism"] = *req.Job.Parallelism
		}

	}
	if req.ServiceAccount != nil {
		yaml["spec"].(map[string]interface{})["serviceAccount"] = *req.ServiceAccount
	}
	return yaml
}

type TaskManager struct {
	Resource          *Resource `json:"resource"`
	NumberOfTaskSlots *string   `json:"numberOfTaskSlots" default:"2"`
}

type JobManager struct {
	Resource *Resource `json:"resource"`
}

type Resource struct {
	Memory *string `json:"memory" default:"2048m"`
	CPU    *int32  `json:"cpu" default:"1"`
}

// yaml 定义的Json 不用_规范
type Job struct {
	JarURI      *string `json:"jarURI"` // jar包路径，application模式必须是local方式将包打包到镜像配合image去做;session模式必须是http方式
	Parallelism *int32  `json:"parallelism"`
	UpgradeMode *string `json:"upgradeMode"` // stateless or stateful
}

type CreateFlinkClusterResponse struct {
	Result any    `json:"result"`
	Info   string `json:"info"`
}

type CreateFlinkSessionJobRequest struct {
	K8SClusterName *string `json:"k8s_cluster_name" binding:"required"` // k8s集群名称
	NameSpace      *string `json:"namespace"`                           // 默认是default
	SubmitJobName  *string `json:"submit_job_name" binding:"required"`  // 提交job名称,实际集群会自动产生一个 job_name ，防止冲突这里叫submit_job_name
	ClusterName    *string `json:"cluster_name" binding:"required"`     // session集群名称 spec.deploymentName
	Job            *Job    `json:"job" binding:"required"`
}

/*
apiVersion: flink.apache.org/v1beta1
kind: FlinkSessionJob
metadata:

	name: basic-session-job-example

spec:

	deploymentName: basic-example-session
	job:
	  jarURI: https://repo1.maven.org/maven2/org/apache/flink/flink-examples-streaming_2.12/1.16.1/flink-examples-streaming_2.12-1.16.1-TopSpeedWindowing.jar
	  parallelism: 4
	  upgradeMode: stateless
*/
func (req *CreateFlinkSessionJobRequest) ToYaml() map[string]any {
	yaml := map[string]any{
		"apiVersion": "flink.apache.org/v1beta1",
		"kind":       "FlinkSessionJob",
		"metadata": map[string]interface{}{
			"name": "basic-session-job-example",
		},
		"spec": map[string]interface{}{
			"deploymentName": "basic-example-session",
			"job": map[string]interface{}{
				"jarURI":      "https://repo1.maven.org/maven2/org/apache/flink/flink-examples-streaming_2.12/1.16.1/flink-examples-streaming_2.12-1.16.1-TopSpeedWindowing.jar",
				"parallelism": 4,
				"upgradeMode": "stateless",
			},
		},
	}
	if req.SubmitJobName != nil {
		yaml["metadata"].(map[string]interface{})["name"] = tea.StringValue(req.SubmitJobName)
	}
	if req.ClusterName != nil {
		yaml["spec"].(map[string]interface{})["deploymentName"] = tea.StringValue(req.ClusterName)
	}
	if req.Job != nil {
		if req.Job.UpgradeMode == nil {
			yaml["spec"].(map[string]interface{})["job"].(map[string]interface{})["upgradeMode"] = "stateless"
		}
		if req.Job.JarURI != nil {
			yaml["spec"].(map[string]interface{})["job"].(map[string]interface{})["jarURI"] = *req.Job.JarURI
		}
		if req.Job.Parallelism != nil {
			yaml["spec"].(map[string]interface{})["job"].(map[string]interface{})["parallelism"] = *req.Job.Parallelism
		}
	}
	// fmt.Println(tea.Prettify(yaml))
	return yaml
}

type DeleteFlinkClusterRequest struct {
	K8SClusterName *string `json:"k8s_cluster_name" binding:"required"` // k8s集群名称
	NameSpace      *string `json:"namespace" default:"default"`
	ClusterName    *string `json:"cluster_name" binding:"required"`
}

type DeleteFlinkSessionJobRequest struct {
	K8SClusterName *string `json:"k8s_cluster_name" binding:"required"` // k8s集群名称
	ClusterName    *string `json:"cluster_name" binding:"required"`     // flink集群名称
	NameSpace      *string `json:"namespace" default:"default"`
	JobName        *string `json:"job_name" binding:"required"`
}
