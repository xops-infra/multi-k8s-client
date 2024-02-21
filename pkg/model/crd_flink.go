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
	ClusterName                string `json:"cluster_name"`
	NameSpace                  string `json:"namespace"`
	ClusterInfo                any    `json:"cluster_info"`
	JobManagerDeploymentStatus any    `json:"job_manager_deployment_status"`
	JobStatus                  any    `json:"status"`     // 集群状态信息
	Annotation                 any    `json:"annotation"` // 集群描述信息
	Error                      any    `json:"error"`      // 集群错误信息
}

type CreateFlinkClusterRequest struct {
	K8SClusterName     *string        `json:"k8s_cluster_name" binding:"required"` // 初始化的k8s集群名称
	NameSpace          *string        `json:"namespace" default:"default"`
	ClusterName        *string        `json:"cluster_name" binding:"required"` // metadata.name
	Image              *string        `json:"image" default:"flink:1.17"`
	Version            *string        `json:"version" default:"v1_17"`
	ServiceAccount     *string        `json:"service_account" default:"flink"`
	FlinkConfiguration map[string]any `json:"flink_configuration"`             // flink配置,键值对的方式比如: {"taskmanager.numberOfTaskSlots": "2"}
	EnableFluentit     *bool          `json:"enableFluentbit" default:"false"` // sidecar fluentbit
	TaskManager        *TaskManager   `json:"task_manager"`
	JobManager         *JobManager    `json:"job_manager"`
	Job                *Job           `json:"job"` // 如果没有该字段则创建 Session集群，如果有该字段则创建Application集群。
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

		state.savepoints.dir: file:///flink-data/savepoints
		state.checkpoints.dir: file:///flink-data/checkpoints
		high-availability: org.apache.flink.kubernetes.highavailability.KubernetesHaServicesFactory
		high-availability.storageDir: file:///flink-data/ha

		kubernetes.operator.periodic.savepoint.interval: 6h # 该操作员还支持通过以下配置选项定期触发保存点，该选项可以在每个作业级别进行配置
	serviceAccount: flink
	podTemplate:
		apiVersion: v1
		kind: Pod
		metadata:
		name: pod-template
		spec:
		containers:
			# Do not change the main container name
			- name: flink-main-container
			volumeMounts:
				- mountPath: /opt/flink/log
				name: flink-logs
			# Sample sidecar container
			- name: fluentbit
			image: fluent/fluent-bit:1.8.12-debug
			command: [ 'sh','-c','/fluent-bit/bin/fluent-bit -i tail -p path=/flink-logs/*.log -p multiline.parser=java -o stdout' ]
			volumeMounts:
				- mountPath: /flink-logs
				name: flink-logs
		volumes:
			- name: flink-logs
			emptyDir: { }
	jobManager:
		resource:
			memory: "2048m"
			cpu: 1
	taskManager:
		resource:
			memory: "2048m"
			cpu: 1
		podTemplate:
			apiVersion: v1
			kind: Pod
			metadata:
				name: task-manager-pod-template
			spec:
				initContainers:
				# Sample sidecar container
				- name: busybox
					image: busybox:1.35.0
					command: [ 'sh','-c','echo hello from task manager' ]
	job:
	  jarURI: local:///opt/flink/examples/streaming/StateMachineExample.jar
	  parallelism: 2  # 2 task managers
	  upgradeMode: stateless # last-state,stateless,savepoint
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
				// "state.savepoints.dir":          "file:///flink-data/savepoints",
				// "state.checkpoints.dir":         "file:///flink-data/checkpoints",
				// "high-availability":             "org.apache.flink.kubernetes.highavailability.KubernetesHaServicesFactory",
				// "high-availability.storageDir":  "file:///flink-data/ha",
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
				"podTemplate": map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name": "task-manager-pod-template",
					},
					"spec": map[string]interface{}{
						"initContainers": []map[string]interface{}{
							{
								"name":  "busybox",
								"image": "busybox:1.35.0",
								"command": []string{
									"sh",
									"-c",
									"echo hello from task manager",
								},
							},
						},
					},
				},
			},
		},
	} // default
	if tea.BoolValue(req.EnableFluentit) {
		yaml["spec"].(map[string]interface{})["podTemplate"] = map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name": "pod-template",
			},
			"spec": map[string]interface{}{
				"containers": []map[string]interface{}{
					{
						"name": "flink-main-container",
						"volumeMounts": []map[string]interface{}{
							{
								"mountPath": "/opt/flink/log",
								"name":      "flink-logs",
							},
						},
					},
					{
						"name":  "fluentbit",
						"image": "fluent/fluent-bit:1.8.12-debug",
						"command": []string{
							"sh",
							"-c",
							"/fluent-bit/bin/fluent-bit -i tail -p path=/flink-logs/*.log -p multiline.parser=java -o stdout",
						},
						"volumeMounts": []map[string]interface{}{
							{
								"mountPath": "/flink-logs",
								"name":      "flink-logs",
							},
						},
					},
				},
				"volumes": []map[string]interface{}{
					{
						"name":     "flink-logs",
						"emptyDir": map[string]interface{}{},
					},
				},
			},
		}
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
	if req.FlinkConfiguration != nil {
		yaml["spec"].(map[string]interface{})["flinkConfiguration"] = req.FlinkConfiguration
	}
	if req.TaskManager != nil && req.TaskManager.Resource != nil {
		yaml["spec"].(map[string]interface{})["taskManager"].(map[string]interface{})["resource"] = req.TaskManager.Resource
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
		if req.Job.Args != nil {
			yaml["spec"].(map[string]interface{})["job"].(map[string]interface{})["args"] = req.Job.Args
		}
	}
	if req.ServiceAccount != nil {
		yaml["spec"].(map[string]interface{})["serviceAccount"] = *req.ServiceAccount
	}
	return yaml
}

type TaskManager struct {
	Resource *Resource `json:"resource"`
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
	JarURI      *string  `json:"jarURI"` // jar包路径，application模式必须是local方式将包打包到镜像配合image去做;session模式必须是http方式
	Parallelism *int32   `json:"parallelism"`
	UpgradeMode *string  `json:"upgradeMode"` // stateless or stateful
	Args        []string `json:"args"`        // 启动参数
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
	} // default
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
