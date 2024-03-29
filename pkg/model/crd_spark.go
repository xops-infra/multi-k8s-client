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
	Attempts   int    `json:"attempts"`    // 重试次数
	StartTime  string `json:"start_time"`  // 启动时间 2024-03-29T05:32:42Z
	FinishTime string `json:"finish_time"` // 完成时间 2024-03-29T05:32:42Z
	Age        string `json:"age"`         // 运行时间 1d
}

type Driver struct {
	Cores  int    `json:"cores" binding:"required"`
	Memory string `json:"memory" binding:"required"` // 512m
}

type CreateSparkApplicationRequest struct {
	K8SClusterName *string `json:"k8s_cluster_name" binding:"required"`
	Namespace      *string `json:"namespace" binding:"required"`
	Driver         *Driver `json:"driver" binding:"required"`
	Executor       *Driver `json:"executor" binding:"required"`
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
		  instances: 4
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
			"name": "spark-pi-example",
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
			"executor": map[string]any{},
		},
	}

	return yaml
}

type CreateSparkApplicationResponse struct{}

type DeleteSparkApplicationRequest struct{}
