package model

import (
	"fmt"
	"math/rand"

	"github.com/alibabacloud-go/tea/tea"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

/*
 * flink1.12.7
 因为 Operator 支持到最低版本是 1.13， 不支持 1.12 版本所以这里使用完整 yaml相关内容来实现
 只抽取必要参数，相对定制化

先创建如下资源
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: flink
  name: flink-configmap-role
rules:
  - apiGroups: ["*"]
    resources: ["configmaps"]
    verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: flink-configmap-role-binding
  namespace: flink
subjects:
  - kind: ServiceAccount
    name: default
    namespace: flink
roleRef:
  kind: Role
  name: flink-configmap-role
  apiGroup: rbac.authorization.k8s.io
*/

const (
	FlinkVersion = "1.12.7"
	// logback-console.xml
	LogbackConsole = `<?xml version="1.0" encoding="UTF-8"?>
    <!--
           ~ Copyright (c) 2022 PatSnap Pte Ltd, All Rights Reserved.
      -->

    <!--
         http://logback.qos.ch/manual/configuration.html
    -->
    <!--configure logback to globally -->
    <configuration debug="${log.config.debug:-false}">
        <contextListener class="ch.qos.logback.classic.jul.LevelChangePropagator"/>

        <!-- common console log pattern -->
        <property name="CONSOLE_LOGGING_PATTERN"
                  value="${log.app.pattern:-%date{ISO8601} [%thread] [%property{flink_job_name}] %-5level %logger{36}.%method:%line - %msg%n}"/>

        <!-- other console log pattern -->
        <if condition='p("log.pattern.type").equals("other")'>
            <then>
                <property name="CONSOLE_LOGGING_PATTERN"
                          value="${log.app.pattern:-%date{ISO8601} [%thread] [${service-name:-unknown}] %-5level %logger{36}.%method:%line - %msg%n}"/>
            </then>
        </if>

        <!-- common log pattern -->
        <property name="APP_LOGGING_PATTERN"
                  value="${log.app.pattern:-%date{ISO8601} [%thread] [%property{flink_job_name}] %-5level %logger{36}.%method:%line - %msg%n}"/>

        <!-- other log pattern -->
        <if condition='p("log.pattern.type").equals("other")'>
            <then>
                <property name="APP_LOGGING_PATTERN"
                          value="${log.app.pattern:-%date{ISO8601} [%thread] [${service-name:-unknown}] %-5level %logger{36}.%method:%line - %msg%n}"/>
            </then>
        </if>

        <!-- allow logging pattern override -->
        <include optional="true" resource="logback-logging-pattern-override.xml"/>

        <appender name="APP_STDOUT" class="ch.qos.logback.core.ConsoleAppender">
            <encoder>
                <pattern>APP ${CONSOLE_LOGGING_PATTERN}</pattern>
            </encoder>
        </appender>

        <appender name="APP_FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
            <file>${log.file}</file>
            <rollingPolicy class="ch.qos.logback.core.rolling.SizeAndTimeBasedRollingPolicy">
                <fileNamePattern>
                    ${log.directory:-/opt/logs/apps}/${log.app.name:-app}.%d{yyyy-MM-dd}.%i.log
                </fileNamePattern>
                <maxFileSize>100MB</maxFileSize>
                <maxHistory>${log.app.maxHistory:-3}</maxHistory>
                <totalSizeCap>20GB</totalSizeCap>
            </rollingPolicy>
            <encoder>
                <pattern>APP ${APP_LOGGING_PATTERN}</pattern>
            </encoder>
        </appender>

        <appender name="APP_FILE_ASYNC" class="ch.qos.logback.classic.AsyncAppender">
            <discardingThreshold>0</discardingThreshold>
            <queueSize>256</queueSize>
            <includeCallerData>true</includeCallerData>
            <appender-ref ref="APP_FILE"/>
        </appender>

        <!-- allow additional logback settings -->
        <include optional="true" resource="logback-overrides.xml"/>

        <root level="${log.app.level:-INFO}">
            <appender-ref ref="APP_STDOUT"/>
            <appender-ref ref="APP_FILE_ASYNC"/>
        </root>

        <jmxConfigurator/>
    </configuration>`
)

var (
	JobManagerServiceName     = "%s-jobmanager-service"
	JobManagerLBServiceName   = "%s-jobmanager-lb-service"
	JobManagerDeploymentName  = "%s-jobmanager"
	TaskManagerDeploymentName = "%s-taskmanager"
	ConfigMapV12Name          = "%s-configmap"
	PvcName                   = "%s-pvc"

	hostLogPath                                = "/mnt/log/%s-flink/"
	rerouceJobManagerDeployment map[string]any = map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			// "name":      "jobmanager",
			"namespace": "default",
		},
		"spec": map[string]any{
			"replicas": 1,
			"selector": map[string]any{
				"matchLabels": map[string]any{
					"component": "jobmanager",
					// "app":       "xxx",
				},
			},
			"template": map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"component": "jobmanager",
						// "app":       "xxx",
					},
				},
				"spec": map[string]any{
					"restartPolicy": "Always",
				},
			},
		},
	}
	resourceTaskManagerDeployment = map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			// "name":      "taskmanager",
			"namespace": "default",
		},
		"spec": map[string]any{
			"replicas": 1,
			"selector": map[string]any{
				"matchLabels": map[string]any{
					"component": "taskmanager",
					// "app":       "xxx",
				},
			},
			"template": map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"component": "taskmanager",
						// "app":       "xxx",
					},
				},
				"spec": map[string]any{
					"restartPolicy": "Always",
				},
			},
		},
	}
)

type FilterFlinkV12 struct {
	NameSpace *string `json:"namespace" default:"default"`
	Owner     *string `json:"owner"`
	Name      *string `json:"name"`
}

type LoadBalancerRequest struct {
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

type JobManagerV12 struct {
	Resource     *FlinkResource     `json:"resource"`
	NodeSelector *map[string]string `json:"node_selector"`
	PvcSize      *int               `json:"pvc_size" default:"10"`
	SideCars     []SideCar          `json:"side_cars"`
}

type SideCar struct {
	Name          *string          `json:"name" binding:"required"`
	Image         *string          `json:"image" binding:"required"`
	Command       []string         `json:"command" binding:"required"`
	Env           []Env            `json:"env"`
	VolumeMounts  []map[string]any `json:"volume_mounts"` // example, [{"name": "xxx", "mountPath": "/xxx"}]
	LivenessProbe *LivenessProbe   `json:"liveness_probe"`
}

type TaskManagerV12 struct {
	Resource     *FlinkResource     `json:"resource"`
	NodeSelector *map[string]string `json:"node_selector"`
	Nu           *int               `json:"nu" default:"5"`
}

type CreateFlinkV12ClusterRequest struct {
	Name               *string              `json:"name" binding:"required"`
	NameSpace          *string              `json:"namespace" default:"flink"`
	Owner              *string              `json:"owner" binding:"required"`
	Image              *string              `json:"image" default:"flink:1.12.7"` // 镜像
	Env                map[string]string    `json:"env"`
	LoadBalancer       *LoadBalancerRequest `json:"loadBalancer"` // 配置相关 annotations启用云主机负载均衡
	TaskManager        *TaskManagerV12      `json:"taskManager"`
	JobManager         *JobManagerV12       `json:"jobManager"`
	FlinkConfigRequest map[string]any       `json:"flinkConfigRequest"` // flink-conf.yaml 的具体配置，example：{"key":"key","value":"value"}
	// NodeSelector       map[string]any       `json:"nodeSelector"`       // {"env":"flink"}
}

// 主要组装 Name和 Size
func (c *CreateFlinkV12ClusterRequest) NewPVC() ApplyPvcRequest {
	req := ApplyPvcRequest{
		Name:        tea.String(fmt.Sprintf(PvcName, *c.Name)),
		Namespace:   c.NameSpace,
		StorageSize: c.JobManager.PvcSize,
	}
	if c.Owner != nil {
		req.Label = map[string]string{"owner": *c.Owner, "app": *c.Name}
	}
	return req
}

func (c *CreateFlinkV12ClusterRequest) NewService() ApplyServiceRequest {
	req := ApplyServiceRequest{
		Name:      tea.String(fmt.Sprintf(JobManagerServiceName, *c.Name)),
		Namespace: c.NameSpace,
		Spec: &ServiceSpec{
			Selector: map[string]string{"app": *c.Name, "component": "jobmanager"},
			Ports: []Port{
				{
					Name:     tea.String("rpc"),
					Protocol: tea.String("TCP"),
					Port:     tea.Int32(6123),
				},
				{
					Name:     tea.String("webui"),
					Protocol: tea.String("TCP"),
					Port:     tea.Int32(8081),
				}, {
					Name:     tea.String("blob-service"),
					Protocol: tea.String("TCP"),
					Port:     tea.Int32(6124),
				},
			},
			Type: tea.String("ClusterIP"),
		},
	}

	if c.Owner != nil {
		req.Label = map[string]string{"owner": *c.Owner, "app": *c.Name}
	}
	return req
}

// Port 不指定自动分配
func (c *CreateFlinkV12ClusterRequest) NewLBService() ApplyServiceRequest {
	// 随机生成30000-32767端口
	min := 30000
	max := 32767
	randPort := rand.Intn(max-min+1) + min

	req := ApplyServiceRequest{
		Name:      tea.String(fmt.Sprintf(JobManagerLBServiceName, *c.Name)),
		Namespace: c.NameSpace,
		Spec: &ServiceSpec{
			Selector: map[string]string{"app": *c.Name, "component": "jobmanager"},
			Ports: []Port{
				{
					Name:       tea.String("webui"),
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
		req.Label = c.LoadBalancer.Labels
		req.Annotations = c.LoadBalancer.Annotations
	}
	return req
}

func (c *CreateFlinkV12ClusterRequest) NewConfigMap() ApplyConfigMapRequest {
	req := ApplyConfigMapRequest{
		Namespace: c.NameSpace,
		Name:      tea.String(fmt.Sprintf(ConfigMapV12Name, *c.Name)),
	}
	if c.Owner != nil {
		req.Labels = map[string]string{"owner": *c.Owner, "app": *c.Name}
	}

	defaultConfig := map[string]any{
		"taskmanager.rpc.port":        6122,
		"jobmanager.rpc.port":         6123,
		"blob.server.port":            6124,
		"queryable-state.proxy.ports": 6125,
		"parallelism.default":         1,

		"taskmanager.numberOfTaskSlots":        2,
		"taskmanager.memory.managed.size":      "24m",
		"taskmanager.memory.task.heap.size":    "3200m",
		"jobmanager.memory.flink.size":         "2048m",
		"jobmanager.memory.jvm-metaspace.size": "2048m",
		"web.upload.dir":                       "/opt/flink/target",
		"jobmanager.rpc.address":               fmt.Sprintf(JobManagerServiceName, *c.Name),
	}
	if c.FlinkConfigRequest != nil {
		for k, v := range c.FlinkConfigRequest {
			defaultConfig[k] = v
		}
	}
	req.Data = map[string]string{
		"flink-conf.yaml":     toString(defaultConfig),
		"logback-console.xml": LogbackConsole,
	}
	return req
}

func toString(value map[string]any) string {
	b, err := yaml.Marshal(value)
	if err != nil {
		return ""
	}
	return string(b)
}

func (c *CreateFlinkV12ClusterRequest) NewJobManagerDeployment() map[string]any {
	yaml := rerouceJobManagerDeployment
	if yaml["metadata"].(map[string]any)["labels"] == nil {
		yaml["metadata"].(map[string]any)["labels"] = map[string]string{}
	}
	yaml["metadata"].(map[string]any)["labels"].(map[string]string)["app"] = *c.Name
	yaml["metadata"].(map[string]any)["name"] = fmt.Sprintf(JobManagerDeploymentName, *c.Name)

	if c.Owner != nil {
		yaml["metadata"].(map[string]any)["labels"].(map[string]string)["owner"] = *c.Owner
	}

	if c.NameSpace != nil {
		yaml["metadata"].(map[string]any)["namespace"] = *c.NameSpace
	}

	// app
	yaml["spec"].(map[string]any)["selector"].(map[string]any)["matchLabels"].(map[string]any)["app"] = *c.Name
	yaml["spec"].(map[string]any)["template"].(map[string]any)["metadata"].(map[string]any)["labels"].(map[string]any)["app"] = *c.Name
	// volumes 组装
	yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["volumes"] = []map[string]any{
		{
			"name": "flink-log",
			"hostPath": map[string]any{
				"path": fmt.Sprintf(hostLogPath, *c.Name),
			},
		},
		{
			"name": "flink-target-pvc",
			"persistentVolumeClaim": map[string]any{
				"claimName": fmt.Sprintf(PvcName, *c.Name),
			},
		},
		{
			"name": "flink-config",
			"configMap": map[string]any{
				"name": fmt.Sprintf(ConfigMapV12Name, *c.Name),
				"items": []map[string]any{
					{
						"key":  "flink-conf.yaml",
						"path": "flink-conf.yaml",
					}, {
						"key":  "logback-console.xml",
						"path": "logback-console.xml",
					},
				},
			},
		},
	}
	// strategy
	yaml["spec"].(map[string]any)["strategy"].(map[string]any)["type"] = "Recreate"

	jobContainer := map[string]any{
		"name":  "jobmanager",
		"image": FlinkVersion,
		"args": []string{
			"jobmanager",
		},
		"env": []map[string]string{},
		"ports": []map[string]any{
			{
				"containerPort": 6123,
				"name":          "rpc",
			},
			{
				"containerPort": 8081,
				"name":          "webui",
			},
			{
				"containerPort": 6124,
				"name":          "blob-service",
			},
		},
		"livenessProbe": map[string]any{
			"tcpSocket": map[string]any{
				"port": 6123,
			},
			"initialDelaySeconds": 30,
			"periodSeconds":       60,
		},
		"securityContext": map[string]any{
			"runAsUser": 0,
		},
		"lifecycle": map[string]any{
			"postStart": map[string]any{
				"exec": map[string]any{
					"command": []string{
						"/bin/sh",
						"-c",
						"mkdir -p /opt/flink/target/flink-web-upload&&chown -R flink:flink /opt/flink/target/",
					},
				},
			},
		},
		"volumeMounts": []map[string]any{
			{
				"mountPath": "/opt/flink/log",
				"name":      "flink-log",
			},
			{
				"mountPath": "/opt/flink/target",
				"name":      "flink-target-pvc",
			},
			{
				"mountPath": "/opt/flink/conf",
				"name":      "flink-config",
			},
		},
	}
	if c.Image != nil {
		jobContainer["image"] = *c.Image
	}
	if c.JobManager != nil && c.JobManager.Resource != nil {
		jobContainer["resources"] = map[string]any{
			"requests": v1.ResourceList(v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(*c.JobManager.Resource.CPU),
				v1.ResourceMemory: resource.MustParse(*c.JobManager.Resource.Memory),
			}),
			"limits": v1.ResourceList(v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(*c.JobManager.Resource.CPU),
				v1.ResourceMemory: resource.MustParse(*c.JobManager.Resource.Memory),
			}),
		}
	}

	if c.Env != nil {
		for k, v := range c.Env {
			jobContainer["env"] = append(jobContainer["env"].([]map[string]string), map[string]string{
				"name":  k,
				"value": v,
			})
		}
	}

	yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"] = []map[string]any{jobContainer}

	// nodeSelector 组装
	if c.JobManager.NodeSelector != nil {
		yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["nodeSelector"] = c.JobManager.NodeSelector
	}

	// support sidecar
	if c.JobManager != nil && len(c.JobManager.SideCars) > 0 {
		for _, sideCar := range c.JobManager.SideCars {
			_sideCar := map[string]any{
				"name":    sideCar.Name,
				"image":   sideCar.Image,
				"command": sideCar.Command,
			}
			if sideCar.Env != nil {
				_sideCar["env"] = sideCar.Env
			}

			if sideCar.VolumeMounts != nil && len(sideCar.VolumeMounts) > 0 {
				_sideCar["volumeMounts"] = sideCar.VolumeMounts
			}

			if sideCar.LivenessProbe != nil {
				_sideCar["livenessProbe"] = sideCar.LivenessProbe
			}

			yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"] = append(
				yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"].([]map[string]any),
				_sideCar,
			)
		}
	}

	return yaml
}

func (c *CreateFlinkV12ClusterRequest) NewTaskManagerDeployment() map[string]any {
	yaml := resourceTaskManagerDeployment

	if yaml["metadata"].(map[string]any)["labels"] == nil {
		yaml["metadata"].(map[string]any)["labels"] = map[string]string{}
	}
	yaml["metadata"].(map[string]any)["labels"].(map[string]string)["app"] = *c.Name
	yaml["metadata"].(map[string]any)["name"] = fmt.Sprintf(TaskManagerDeploymentName, *c.Name)

	if c.Owner != nil {
		yaml["metadata"].(map[string]any)["labels"].(map[string]string)["owner"] = *c.Owner
	}

	if c.NameSpace != nil {
		yaml["metadata"].(map[string]any)["namespace"] = *c.NameSpace
	}
	yaml["spec"].(map[string]any)["replicas"] = c.TaskManager.Nu
	// containers 组装
	taskManagerContainer := map[string]any{
		"name":  "taskmanager",
		"image": FlinkVersion,
		"args": []string{
			"taskmanager",
		},
		"ports": []map[string]any{
			{
				"containerPort": 6122,
				"name":          "rpc",
			},
			{
				"containerPort": 6125,
				"name":          "query-state",
			},
		},
		"livenessProbe": map[string]any{
			"tcpSocket": map[string]any{
				"port": 6122,
			},
			"initialDelaySeconds": 30,
			"periodSeconds":       60,
		},
		"env": []map[string]string{},
		"securityContext": map[string]any{
			"runAsUser": 0,
		},
		"lifecycle": map[string]any{
			"postStart": map[string]any{
				"exec": map[string]any{
					"command": []string{
						"sh",
						"-c",
						"chown 9999:9999 /opt/flink/log",
					},
				},
			},
		},

		"volumeMounts": []map[string]any{
			{
				"mountPath": "/opt/flink/log",
				"name":      "flink-log",
			},
			{
				"mountPath": "/opt/flink/conf",
				"name":      "flink-config",
			},
		},
	}
	if c.Image != nil {
		taskManagerContainer["image"] = *c.Image
	}
	if c.Env != nil {
		for k, v := range c.Env {
			taskManagerContainer["env"] = append(taskManagerContainer["env"].([]map[string]string), map[string]string{
				"name":  k,
				"value": v,
			})
		}
	}
	if c.TaskManager != nil && c.TaskManager.Resource != nil {
		taskManagerContainer["resources"] = map[string]any{
			"requests": v1.ResourceList(v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("100m"), // 默认少点，避免资源浪费
				v1.ResourceMemory: resource.MustParse(*c.TaskManager.Resource.Memory),
			}),
			"limits": v1.ResourceList(v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(*c.TaskManager.Resource.CPU),
				v1.ResourceMemory: resource.MustParse(*c.TaskManager.Resource.Memory),
			}),
		}
	}

	yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"] = []map[string]any{taskManagerContainer}

	// volume 组装
	yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["volumes"] = []map[string]any{
		{
			"name": "flink-log",
			"hostPath": map[string]any{
				"path": fmt.Sprintf(hostLogPath, *c.Name),
			},
		},
		{
			"name": "flink-config",
			"configMap": map[string]any{
				"name": fmt.Sprintf(ConfigMapV12Name, *c.Name),
				"items": []map[string]any{
					{
						"key":  "flink-conf.yaml",
						"path": "flink-conf.yaml",
					}, {
						"key":  "logback-console.xml",
						"path": "logback-console.xml",
					},
				},
			},
		},
	}

	// nodeSelector 组装
	if c.TaskManager.NodeSelector != nil {
		yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["nodeSelector"] = c.TaskManager.NodeSelector
	}
	return yaml
}
