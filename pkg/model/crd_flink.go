package model

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/spf13/cast"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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
	Annotation     any    `json:"annotation"` // 集群描述信息
	Error          any    `json:"error"`
}

type CrdFlinkDeploymentGetResponse struct {
	Total int                  `json:"total"`
	Items []CrdFlinkDeployment `json:"items"`
}

type CrdFlinkDeployment struct {
	ClusterName  string                 `json:"cluster_name"`
	NameSpace    string                 `json:"namespace"`
	Labels       map[string]string      `json:"labels"`
	Status       any                    `json:"status"`       // 集群状态信息
	Annotation   any                    `json:"annotation"`   // 集群描述信息
	LoadBalancer map[string]string      `json:"loadBalancer"` // 如果创建的时候带了，这里可以查询信息
	Info         CrdFlinkDeploymentInfo `json:"info"`         // 集群额外信息，比如集群版本，启动时间，副本数量，cpu、内存等信息
	FlinkConfig  map[string]any         `json:"flink_config"` // flink Operator 的值是可以是数字比如 slot的数量	 低版本在 k8sconfig 是 string方式
}

// 在 label里面获取 owner
func (s *CrdFlinkDeployment) GetOwner() (string, error) {
	if labels, ok := s.Labels["owner"]; ok {
		return labels, nil
	}
	return "", fmt.Errorf("labels.owner not found")
}

// GetWebUrl 自己 Loadbalancer里面获取的
func (s *CrdFlinkDeployment) GetWebUrl() (string, error) {
	var webUrl string
	for _, v := range s.LoadBalancer {
		webUrl = fmt.Sprintf("%s,%s", webUrl, v)
	}
	webUrl = strings.Trim(webUrl, ",")
	if webUrl != "" {
		return webUrl, nil
	}
	return "", fmt.Errorf("loadBalancer not found")
}

type CrdFlinkDeploymentInfo map[string]string

func (s *CrdFlinkDeploymentInfo) Get(key string) string {
	if _, ok := (*s)[key]; ok {
		return (*s)[key]
	}
	return ""
}

func (s *CrdFlinkDeploymentInfo) GetOk(key string) (string, bool) {
	if _, ok := (*s)[key]; ok {
		return (*s)[key], true
	}
	return "", false
}

// create_time Format("2006-01-02 15:04:05")
func (s *CrdFlinkDeploymentInfo) GetCreateTime() (time.Time, error) {
	if createTime, ok := s.GetOk("create_time"); ok {
		return time.Parse("2006-01-02 15:04:05", createTime)
	}
	return time.Time{}, fmt.Errorf("create_time not found")
}

// getRunTime as 1y31d1h1m1s
func (s *CrdFlinkDeploymentInfo) GetRunTime() (string, error) {
	createTime, err := s.GetCreateTime()
	if err != nil {
		return "", err
	}

	// 计算时间差
	duration := time.Since(createTime)

	// 提取各个时间单位
	years := duration / (365 * 24 * time.Hour) // 估算一年为 365 天
	duration -= years * 365 * 24 * time.Hour

	days := duration / (24 * time.Hour)
	duration -= days * 24 * time.Hour

	hours := duration / time.Hour
	duration -= hours * time.Hour

	minutes := duration / time.Minute
	duration -= minutes * time.Minute

	seconds := duration / time.Second

	// 根据优先级选择两个单位
	result := ""

	if years > 0 {
		result = fmt.Sprintf("%dy", years)
		if days > 0 {
			result += fmt.Sprintf("%dd", days)
		}
	} else if days > 0 {
		result = fmt.Sprintf("%dd", days)
		if hours > 0 {
			result += fmt.Sprintf("%dh", hours)
		}
	} else if hours > 0 {
		result = fmt.Sprintf("%dh", hours)
		if minutes > 0 {
			result += fmt.Sprintf("%dm", minutes)
		}
	} else if minutes > 0 {
		result = fmt.Sprintf("%dm", minutes)
		if seconds > 0 {
			result += fmt.Sprintf("%ds", seconds)
		}
	} else {
		result = fmt.Sprintf("%ds", seconds)
	}

	return result, nil
}

func (s *CrdFlinkDeploymentInfo) GetReplicas() (int32, error) {
	if replicas, ok := s.GetOk("replicas"); ok {
		return cast.ToInt32(replicas), nil
	}
	return 0, fmt.Errorf("replicas not found")
}

// images is string image with ,
func (s *CrdFlinkDeploymentInfo) GetImages() (string, error) {
	if images, ok := s.GetOk("images"); ok {
		return images, nil
	}
	return "", fmt.Errorf("images not found")
}

func (s *CrdFlinkDeploymentInfo) GetVersion() (string, error) {
	if version, ok := s.GetOk("version"); ok {
		return version, nil
	}
	return "", fmt.Errorf("version not found")
}

type CreateFlinkClusterRequest struct {
	NameSpace          *string              `json:"namespace" default:"default"`
	ClusterName        *string              `json:"cluster_name" binding:"required"` // metadata.name，必须符合k8s标准不支持中文，下划线等
	Image              *string              `json:"image" default:"flink:1.17"`
	Version            *string              `json:"version" default:"v1_17"`
	ServiceAccount     *string              `json:"service_account" default:"flink"`
	FlinkConfiguration map[string]any       `json:"flink_configuration"`              // flink配置,键值对的方式比如: {"taskmanager.numberOfTaskSlots": "2"}
	EnableFluentit     *bool                `json:"enable_fluentbit" default:"false"` // sidecar fluentbit
	Env                []Env                `json:"env"`                              // 环境变量,同时给JM和TM设置环境变量
	TaskManager        *Manager             `json:"task_manager"`
	JobManager         *Manager             `json:"job_manager"`
	Job                *Job                 `json:"job"`          // 如果没有该字段则创建 Session集群，如果有该字段则创建Application集群。
	Submitter          *string              `json:"submitter"`    // 提交人
	Labels             map[string]string    `json:"labels"`       // 自定义标签
	LoadBalancer       *LoadBalancerRequest `json:"loadBalancer"` // 配置相关 annotations启用云主机负载均衡,nil不会启用
}

/*
kind: Service
apiVersion: v1
metadata:

	name: flink-application-aiops-1-rest
	namespace: flink
	uid: 7b877cf1-112e-4d97-b2d9-b2b058b7f47c
	resourceVersion: '15803757841'
	creationTimestamp: '2024-11-06T07:46:17Z'
	labels:
	  app: flink-application-aiops-1
	  type: flink-native-kubernetes

spec:

	ports:
	  - name: rest
	    protocol: TCP
	    port: 8081
	    targetPort: 8081
	selector:
	  app: flink-application-aiops-1
	  component: jobmanager
	  type: flink-native-kubernetes
*/
func (c *CreateFlinkClusterRequest) NewLBService() ApplyServiceRequest {
	// 随机生成30000-32767端口
	min := 30000
	max := 32767
	randPort := rand.Intn(max-min+1) + min

	req := ApplyServiceRequest{
		Name:      tea.String(fmt.Sprintf(JobManagerLBServiceName, *c.ClusterName)),
		Namespace: c.NameSpace,
		Spec: &ServiceSpec{
			Selector: map[string]string{"app": *c.ClusterName, "component": "jobmanager", "type": "flink-native-kubernetes"},
			Ports: []Port{
				{
					Name:       tea.String("rest"),
					Protocol:   tea.String("TCP"),
					Port:       tea.Int32(int32(randPort)),
					TargetPort: tea.Int32(8081),
					NodePort:   tea.Int32(int32(randPort)),
				},
			},
			Type: tea.String("LoadBalancer"),
		},
	}

	if c.LoadBalancer != nil {
		req.Annotations = c.LoadBalancer.Annotations
		if c.LoadBalancer.Labels != nil {
			req.Labels = c.LoadBalancer.Labels
		}
	}
	if c.Submitter != nil {
		if req.Labels == nil {
			req.Labels = map[string]string{}
		}
		req.Labels["owner"] = *c.Submitter
		req.Labels["app"] = *c.ClusterName
	}
	return req
}

type Env struct {
	Name  *string `json:"name"`
	Value *string `json:"value"`
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
			"name":      "basic-example",
			"namespace": "default",
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
				"podTemplate": map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"sdk":   "multi-k8s-client",
							"owner": tea.StringValue(req.Submitter),
						},
					},
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
						"labels": map[string]interface{}{
							"sdk":   "multi-k8s-client",
							"owner": tea.StringValue(req.Submitter),
						},
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

	if req.EnableFluentit != nil || len(req.Env) != 0 {

		mainContainer := map[string]any{
			"name": "flink-main-container",
			"volumeMounts": []map[string]interface{}{
				{
					"mountPath": "/opt/flink/log",
					"name":      "flink-logs",
				},
			},
		}
		fluentitContainer := map[string]any{
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
		}
		if len(req.Env) != 0 {
			mainContainer["env"] = req.Env
		}

		containers := []map[string]any{
			mainContainer,
		}
		if tea.BoolValue(req.EnableFluentit) {
			containers = append(containers, fluentitContainer)
		}
		yaml["spec"].(map[string]interface{})["podTemplate"] = map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name": "pod-template",
			},
			"spec": map[string]interface{}{
				"containers": containers,
				"volumes": []map[string]interface{}{
					{
						"name": "flink-logs",
						"emptyDir": map[string]interface{}{
							"medium": "Memory",
						},
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
	if req.TaskManager != nil {
		if req.TaskManager.Resource != nil {
			// 转换cpu 的 str到这里 int类型
			cpu := cast.ToInt(*req.JobManager.Resource.CPU)
			yaml["spec"].(map[string]interface{})["taskManager"].(map[string]interface{})["resource"] = map[string]interface{}{
				"memory": req.JobManager.Resource.Memory,
				"cpu":    cpu,
			}
		}
		if req.TaskManager.NodeSelector != nil {
			yaml["spec"].(map[string]interface{})["taskManager"].(map[string]interface{})["podTemplate"] = map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "task-manager-pod-template",
				},
				"spec": map[string]interface{}{},
			}
			yaml["spec"].(map[string]interface{})["taskManager"].(map[string]interface{})["podTemplate"].(map[string]interface{})["spec"].(map[string]interface{})["nodeSelector"] = *req.TaskManager.NodeSelector
		}
	}
	if req.JobManager != nil {
		if req.JobManager.Resource != nil {
			// 转换cpu 的 str到这里 int类型
			cpu := cast.ToInt(*req.JobManager.Resource.CPU)
			yaml["spec"].(map[string]interface{})["jobManager"].(map[string]interface{})["resource"] = map[string]interface{}{
				"memory": req.JobManager.Resource.Memory,
				"cpu":    cpu,
			}
		}
		if req.JobManager.NodeSelector != nil {
			yaml["spec"].(map[string]interface{})["jobManager"].(map[string]interface{})["podTemplate"] = map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "job-manager-pod-template",
				},
				"spec": map[string]interface{}{},
			}
			yaml["spec"].(map[string]interface{})["jobManager"].(map[string]interface{})["podTemplate"].(map[string]interface{})["spec"].(map[string]interface{})["nodeSelector"] = *req.JobManager.NodeSelector
		}
	}
	if req.Job != nil {
		yaml["spec"].(map[string]interface{})["job"] = req.Job.ToYaml()
	}
	if req.ServiceAccount != nil {
		yaml["spec"].(map[string]interface{})["serviceAccount"] = *req.ServiceAccount
	}
	if req.Labels != nil {
		yaml["metadata"].(map[string]interface{})["labels"] = req.Labels
	} else {
		yaml["metadata"].(map[string]interface{})["labels"] = map[string]string{}
	}
	if req.Submitter != nil {
		yaml["metadata"].(map[string]interface{})["labels"].(map[string]string)["owner"] = *req.Submitter
	}

	return yaml
}

type Manager struct {
	Resource     *FlinkResource     `json:"resource"`
	NodeSelector *map[string]string `json:"node_selector"`
}

type FlinkResource struct {
	Memory *string `json:"memory" default:"2048m"`
	CPU    *string `json:"cpu" default:"1"`
}

// yaml 定义的Json 不用_规范
type Job struct {
	JarURI      *string  `json:"jar_url"` // jar包路径，application模式必须是local方式将包打包到镜像配合image去做;session模式必须是http方式
	Parallelism *int32   `json:"parallelism"`
	UpgradeMode *string  `json:"upgrade_mode"` // stateless or stateful
	Args        []string `json:"args"`         // 启动参数 --arg1=value1
	EntryClass  *string  `json:"entry_class"`  // 主类
}

func (j *Job) ToYaml() map[string]any {
	yaml := map[string]any{
		"parallelism": 2,
		"upgradeMode": "stateless",
	}
	if j.UpgradeMode != nil {
		yaml["upgradeMode"] = *j.UpgradeMode
	}
	if j.JarURI != nil {
		yaml["jarURI"] = *j.JarURI
	}
	if j.Parallelism != nil {
		yaml["parallelism"] = *j.Parallelism
	}
	if j.Args != nil {
		yaml["args"] = j.Args
	}
	if j.EntryClass != nil {
		yaml["entryClass"] = *j.EntryClass
	}
	return yaml
}

type CreateFlinkSessionJobRequest struct {
	NameSpace     *string `json:"namespace"`                          // 默认是default
	SubmitJobName *string `json:"submit_job_name" binding:"required"` // 提交job名称,实际集群会自动产生一个 job_name ，防止冲突这里叫submit_job_name
	ClusterName   *string `json:"cluster_name" binding:"required"`    // session集群名称 spec.deploymentName
	Job           *Job    `json:"job" binding:"required"`
	Submitter     *string `json:"submitter"` // 提交人
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
		},
	} // default
	if req.SubmitJobName != nil {
		yaml["metadata"].(map[string]interface{})["name"] = tea.StringValue(req.SubmitJobName)
	}
	if req.ClusterName != nil {
		yaml["spec"].(map[string]interface{})["deploymentName"] = tea.StringValue(req.ClusterName)
	}
	if req.Submitter != nil {
		yaml["metadata"].(map[string]interface{})["labels"] = map[string]interface{}{}
		yaml["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["owner"] = tea.StringValue(req.Submitter)
	}
	if req.Job != nil {
		yaml["spec"].(map[string]interface{})["job"] = req.Job.ToYaml()
	}
	// fmt.Println(tea.Prettify(yaml))
	return yaml
}

type DeleteFlinkClusterRequest struct {
	NameSpace   *string `json:"namespace" default:"default"`
	ClusterName *string `json:"cluster_name" binding:"required"`
}

type DeleteFlinkSessionJobRequest struct {
	ClusterName *string `json:"cluster_name" binding:"required"` // flink集群名称
	NameSpace   *string `json:"namespace" default:"default"`
	JobName     *string `json:"job_name" binding:"required"`
}

type FlinkType string

const (
	FlinkTypeJM  FlinkType = "JM" // jobmanager
	FlinkTypeTM  FlinkType = "TM" // taskmanager
	FlinkTypeALL FlinkType = "ALL"
)

type RestartFlinkClusterRequest struct {
	ClusterName *string   `json:"cluster_name" binding:"required"` // flink集群名称
	NameSpace   *string   `json:"namespace" default:"default"`
	Type        FlinkType `json:"type" binding:"required"` // JM/TM/ALL
}

type CrdFlinkTMScaleRequest struct {
	ClusterName *string `json:"cluster_name" binding:"required"` // flink集群名称
	NameSpace   *string `json:"namespace" default:"default"`
	Replicas    *int32  `json:"replicas" binding:"required"` // 调整后的 TM 数量
}

// 获取创建时间，replicas，更新时间
func GetInfoFromItem(item unstructured.Unstructured) CrdFlinkDeploymentInfo {
	data := make(map[string]string, 0)
	for k, v := range item.GetLabels() {
		data[k] = v
	}
	data["create_time"] = item.GetCreationTimestamp().Local().Format("2006-01-02 15:04:05")
	if r, ok := item.Object["spec"].(map[string]any)["replicas"]; ok {
		data["replicas"] = fmt.Sprintf("%v", r)
	}
	data["images"] = GetFlinkImageFromItem(item)
	data["version"] = GetFlinkVersionFromItem(item)
	return data
}

func GetFlinkConfigFromItem(item unstructured.Unstructured) map[string]any {
	if config, ok := item.Object["spec"].(map[string]any)["flinkConfiguration"]; ok {
		return config.(map[string]any)
	}
	return nil
}

func GetFlinkImageFromItem(item unstructured.Unstructured) string {
	if image, ok := item.Object["spec"].(map[string]any)["image"]; ok {
		return image.(string)
	}
	return ""
}

func GetFlinkVersionFromItem(item unstructured.Unstructured) string {
	// flinkVersion
	if version, ok := item.Object["spec"].(map[string]any)["flinkVersion"]; ok {
		return version.(string)
	}
	return ""
}

func GetInfoFromDeploymentForV12(item v1.Deployment) CrdFlinkDeploymentInfo {
	data := make(map[string]string, 0)
	data["create_time"] = item.GetCreationTimestamp().Local().Format("2006-01-02 15:04:05")
	data["replicas"] = fmt.Sprintf("%d", tea.Int32Value(item.Spec.Replicas))
	for k, v := range item.Labels {
		data[k] = v
	}
	images := ""
	for _, v := range item.Spec.Template.Spec.Containers {
		images = fmt.Sprintf("%s,%s", images, v.Image)
	}
	data["images"] = strings.Trim(images, ",")

	// get flink version from image tag
	for _, v := range item.Spec.Template.Spec.Containers {
		if strings.Contains(v.Image, "flink") {
			data["version"] = strings.Split(v.Image, ":")[1]
			break
		}
	}
	return data
}
