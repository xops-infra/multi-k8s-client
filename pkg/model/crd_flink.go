package model

type CreateFlinkRequest struct {
	NameSpace      *string      `json:"nameSpace" default:"default"`
	ClusterName    *string      `json:"clusterName" binding:"required"`
	Image          *string      `json:"image" default:"flink:1.17"`
	Version        *string      `json:"version" default:"v1_17"`
	ServiceAccount *string      `json:"serviceAccount" default:"flink"`
	TaskManager    *TaskManager `json:"taskManager"`
	JobManager     *JobManager  `json:"jobManager"`
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
func (req *CreateFlinkRequest) ToYaml() map[string]any {
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
			// "job": map[string]interface{}{
			// 	"jarURI":      "local:///opt/flink/examples/streaming/StateMachineExample.jar",
			// 	"parallelism": 2,
			// 	"upgradeMode": "stateless",
			// },
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
		yaml["spec"].(map[string]interface{})["job"] = req.Job
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

type Job struct {
	JarURI      *string `json:"jarURI" binding:"required"`
	Parallelism *int32  `json:"parallelism" binding:"required"`
	UpgradeMode *string `json:"upgradeMode" binding:"required"` // stateless or stateful
}

type FlinkSessionJobRequest struct {
	JobName     *string `json:"jobName" binding:"required"`
	ClusterName *string `json:"clusterName" binding:"required"` // session集群名称
	Job         *Job    `json:"job" binding:"required"`
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
func (req *FlinkSessionJobRequest) ToYaml() map[string]any {
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
	if req.JobName != nil {
		yaml["metadata"].(map[string]interface{})["name"] = *req.JobName
	}
	if req.ClusterName != nil {
		yaml["spec"].(map[string]interface{})["deploymentName"] = *req.ClusterName
	}
	if req.Job != nil {
		yaml["spec"].(map[string]interface{})["job"] = req.Job
	}
	// fmt.Println(tea.Prettify(yaml))
	return yaml
}

type DeleteFlinkClusterRequest struct {
	ClusterName *string `json:"clusterName" binding:"required"`
	NameSpace   *string `json:"nameSpace" default:"default"`
	Name        *string `json:"name" binding:"required"`
}

type DeleteFlinkJobRequest struct {
	ClusterName *string `json:"clusterName" binding:"required"`
	NameSpace   *string `json:"nameSpace" default:"default"`
	Name        *string `json:"name" binding:"required"`
}
