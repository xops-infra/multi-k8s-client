package model

// https://github.com/kubeflow/spark-operator/blob/master/docs/quick-start-guide.md

type CrdSparkApplicationGetResponse struct {
	Items []CrdSparkApplication `json:"items"`
	Total int                   `json:"total"`
}

type CrdSparkApplication struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Status     string `json:"status"`      // 状态 COMPLETED,
	Attempts   int64  `json:"attempts"`    // 重试次数
	StartTime  string `json:"start_time"`  // 启动时间 2024-03-29T05:32:42Z
	FinishTime string `json:"finish_time"` // 完成时间 2024-03-29T05:32:42Z
	Age        string `json:"age"`         // 运行时间 1d
}

type Driver struct {
	Cores     int               `json:"cores" `     // default:1
	CoreLimit string            `json:"coreLimit" ` // default:1200m
	Memory    string            `json:"memory" `    // default:512m
	Labels    map[string]string `json:"labels" `
}

type Executor struct {
	Cores     int               `json:"cores" `     // default:2
	Instances int               `json:"instances" ` // default:2
	Memory    string            `json:"memory" `    // default:512m
	Labels    map[string]string `json:"labels" `
}

type CreateSparkApplicationRequest struct {
	Name                *string   `json:"name" binding:"required"` // spark-pi-example
	Namespace           *string   `json:"namespace"`
	Type                *string   `json:"type" `                 // Scala
	Mode                *string   `json:"mode" `                 // cluster
	Image               *string   `json:"image"`                 // apache/spark-py:v3.2.1
	SparkVersion        *string   `json:"spark_version"`         // 3.2.1
	MainClass           *string   `json:"main_class"`            // org.apache.spark.examples.SparkPi
	MainApplicationFile *string   `json:"main_application_file"` // local:///opt/spark/examples/jars/spark-examples_2.12-3.2.1.jar
	RestartPolicy       *string   `json:"restart_policy"`        // Never
	Driver              *Driver   `json:"driver"`
	Executor            *Executor `json:"executor"`
	EnableMonitoring    bool      `json:"enable_monitoring"` // prometheus port 8090
}

/*
apiVersion: "sparkoperator.k8s.io/v1beta2"
kind: SparkApplication
metadata:

	name: spark-pi-321
	namespace: default

spec:

		type: Scala
		mode: cluster
		image: "apache/spark-py:v3.2.1"
		imagePullPolicy: Always
		mainClass: org.apache.spark.examples.SparkPi
		mainApplicationFile: "local:///opt/spark/examples/jars/spark-examples_2.12-3.2.1.jar"
		sparkVersion: "3.2.1"
		restartPolicy:
		  type: Never
		volumes:
		  - name: "test-volume"
		    hostPath:
		      path: "/tmp"
		      type: Directory

		driver:
		  cores: 1
		  coreLimit: "1200m"
		  memory: "512m"
		  labels:
		    version: 3.2.1
		  serviceAccount: spark
		  volumeMounts:
		    - name: "test-volume"
		      mountPath: "/tmp"
		monitoring:
	      exposeDriverMetrics: true
	       exposeExecutorMetrics: true
	      prometheus:
	        jmxExporterJar: "/prometheus/jmx_prometheus_javaagent-0.11.0.jar"
	        port: 8090
		executor:
		  cores: 2
		  instances: 2
		  memory: "512m"
		  labels:
		    version: 3.2.1
		  volumeMounts:
		    - name: "test-volume"
		      mountPath: "/tmp"
*/
func (req *CreateSparkApplicationRequest) ToYaml() map[string]any {
	// default
	yaml := map[string]any{
		"apiVersion": "sparkoperator.k8s.io/v1beta2",
		"kind":       "SparkApplication",
		"metadata": map[string]any{
			"namespace": "default",
		},
		"spec": map[string]any{
			"type":                "Scala",
			"mode":                "cluster",
			"image":               "apache/spark-py:v3.2.1",
			"imagePullPolicy":     "Always",
			"mainClass":           "org.apache.spark.examples.SparkPi",
			"mainApplicationFile": "local:///opt/spark/examples/jars/spark-examples_2.12-3.2.1.jar",
			"sparkVersion":        "3.2.1",
			"restartPolicy": map[string]any{
				"type": "Never",
			},
			"volumes": []map[string]any{
				{
					"name": "test-volume",
					"hostPath": map[string]any{
						"path": "/tmp",
						"type": "Directory",
					},
				},
			},
			"driver": map[string]any{
				"cores":          1,
				"coreLimit":      "1200m",
				"memory":         "512m",
				"labels":         map[string]any{"version": "3.2.1"},
				"serviceAccount": "spark",
				"volumeMounts": []map[string]any{
					{
						"name":      "test-volume",
						"mountPath": "/tmp",
					},
				},
			},
			"executor": map[string]any{
				"cores":     2,
				"instances": 2,
				"memory":    "512m",
				"labels":    map[string]any{"version": "3.2.1"},
				"volumeMounts": []map[string]any{
					{
						"name":      "test-volume",
						"mountPath": "/tmp",
					},
				},
			},
		},
	}
	if req.Name != nil {
		yaml["metadata"].(map[string]any)["name"] = *req.Name
		yaml["spec"].(map[string]any)["driver"].(map[string]any)["labels"].(map[string]any)["app"] = *req.Name + "-driver"
		yaml["spec"].(map[string]any)["executor"].(map[string]any)["labels"].(map[string]any)["app"] = *req.Name + "-executor"
	}
	if req.Namespace != nil {
		yaml["metadata"].(map[string]any)["namespace"] = *req.Namespace
	}
	if req.Type != nil {
		yaml["spec"].(map[string]any)["type"] = *req.Type
	}
	if req.Mode != nil {
		yaml["spec"].(map[string]any)["mode"] = *req.Mode
	}
	if req.Image != nil {
		yaml["spec"].(map[string]any)["image"] = *req.Image
	}

	if req.MainClass != nil {
		yaml["spec"].(map[string]any)["mainClass"] = *req.MainClass
	}
	if req.MainApplicationFile != nil {
		yaml["spec"].(map[string]any)["mainApplicationFile"] = *req.MainApplicationFile
	}
	if req.RestartPolicy != nil {
		yaml["spec"].(map[string]any)["restartPolicy"].(map[string]any)["type"] = *req.RestartPolicy
	}

	if req.Driver != nil && req.Driver.Cores != 0 {
		yaml["spec"].(map[string]any)["driver"].(map[string]any)["cores"] = req.Driver.Cores
	}
	if req.Driver != nil && req.Driver.CoreLimit != "" {
		yaml["spec"].(map[string]any)["driver"].(map[string]any)["coreLimit"] = req.Driver.CoreLimit
	}
	if req.Driver != nil && req.Driver.Memory != "" {
		yaml["spec"].(map[string]any)["driver"].(map[string]any)["memory"] = req.Driver.Memory
	}
	if req.Driver != nil && len(req.Driver.Labels) > 0 {
		for k, v := range req.Driver.Labels {
			yaml["spec"].(map[string]any)["driver"].(map[string]any)["labels"].(map[string]any)[k] = v
		}
	}

	if req.Executor != nil && req.Executor.Cores != 0 {
		yaml["spec"].(map[string]any)["executor"].(map[string]any)["cores"] = req.Executor.Cores
	}
	if req.Executor != nil && req.Executor.Instances != 0 {
		yaml["spec"].(map[string]any)["executor"].(map[string]any)["instances"] = req.Executor.Instances
	}
	if req.Executor != nil && req.Executor.Memory != "" {
		yaml["spec"].(map[string]any)["executor"].(map[string]any)["memory"] = req.Executor.Memory
	}
	if req.Executor != nil && len(req.Executor.Labels) > 0 {
		for k, v := range req.Executor.Labels {
			yaml["spec"].(map[string]any)["executor"].(map[string]any)["labels"].(map[string]any)[k] = v
		}
	}
	if req.EnableMonitoring {
		yaml["spec"].(map[string]any)["monitoring"] = map[string]any{
			"exposeDriverMetrics":   true,
			"exposeExecutorMetrics": true,
			"prometheus": map[string]any{
				"jmxExporterJar": "/prometheus/jmx_prometheus_javaagent-0.11.0.jar",
				"port":           8090,
			},
		}
	}
	if req.SparkVersion != nil {
		yaml["spec"].(map[string]any)["sparkVersion"] = *req.SparkVersion
		yaml["spec"].(map[string]any)["executor"].(map[string]any)["labels"].(map[string]any)["version"] = *req.SparkVersion
		yaml["spec"].(map[string]any)["driver"].(map[string]any)["labels"].(map[string]any)["version"] = *req.SparkVersion
	}

	return yaml
}

type DeleteSparkApplicationRequest struct {
	Namespace *string `json:"namespace" binding:"required"`
	Name      *string `json:"name" binding:"required"`
}
