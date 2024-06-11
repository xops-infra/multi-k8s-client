package model

import (
	"fmt"

	"github.com/alibabacloud-go/tea/tea"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

/*
 * flink1.12.7
 因为 Operator 支持到最低版本是 1.13， 不支持 1.12 版本所以这里使用完整 yaml相关内容来实现
 只抽取必要参数，相对定制化
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
	jobManagerServiceName                    = "%s-jobmanager-service"
	jobManagerDeploymentName                 = "%s-jobmanager"
	taskManagerDeploymentName                = "%s-taskmanager"
	configMapV12Name                         = "%s-configmap"
	pvcName                                  = "%s-pvc"
	hostLogPath                              = "/mnt/log/%s-flink/"
	resourceFlinkV12ConfigMap map[string]any = map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata":   map[string]any{
			// "name":      "flink-configuration",
		},
		"data": map[string]any{
			"logback.xml": LogbackConsole,
		},
	}
	resourcePVC map[string]any = map[string]any{
		"apiVersion": "v1",
		"kind":       "PersistentVolumeClaim",
		"metadata":   map[string]any{
			// "name":      "pvc",
		},
		"spec": map[string]any{
			"accessModes": []string{
				"ReadWriteOnce",
			},
			"resources": map[string]any{
				"requests": map[string]any{
					"storage": "10Gi", // default 10Gi
				},
			},
		},
	}
	rerouceJobManagerService map[string]any = map[string]any{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata":   map[string]any{
			// "name":      "jobmanager",
		},
		"spec": map[string]any{
			"ports": []map[string]any{
				{
					"port": 6123,
					"name": "rpc",
				},
				{
					"port": 8081,
					"name": "webui",
				},
				{
					"port": 6124,
					"name": "blob-service",
				},
			},
			"selector": map[string]any{
				"component": "jobmanager",
				// "app":       "xxx",
			},
			"type": "ClusterIP",
		},
	}
	rerouceJobManagerLBService map[string]any = map[string]any{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata":   map[string]any{
			// "name":      "jobmanager",
			// "annotations": map[string]any{},
		},
		"spec": map[string]any{
			"ports": []map[string]any{
				{
					// "port": 38081,
					"targetPort": 8081,
					"name":       "webui",
				},
			},
			"selector": map[string]any{
				"component": "jobmanager",
				// "app":       "xxx",
			},
			"type": "LoadBalancer",
		},
	}
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
	resourceClusterRole = map[string]any{
		"apiVersion": "rbac.authorization.k8s.io/v1",
		"kind":       "ClusterRole",
		"metadata": map[string]any{
			"name": "flink-cluster-role",
		},
		"rules": []map[string]any{
			{
				"apiGroups": []string{
					"*",
				},
				"resources": []string{
					"configmaps",
				},
				"verbs": []string{
					"*",
				},
			},
		},
	}
	resourceClusterRoleBinding = map[string]any{
		"apiVersion": "rbac.authorization.k8s.io/v1",
		"kind":       "ClusterRoleBinding",
		"metadata": map[string]any{
			"name": "flink-cluster-role-binding",
		},
		"roleRef": map[string]any{
			"apiGroup": "rbac.authorization.k8s.io",
			"kind":     "ClusterRole",
			"name":     "flink-cluster-role",
		},
		"subjects": []map[string]any{
			// {
			// 	"kind":      "User",
			// 	"name":      "system:anonymous",
			// 	"namespace": "flink",
			// },
		},
	}
)

type LoadBalancerRequest struct {
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

type JobManagerV12 struct {
	Resource     *FlinkResource     `json:"resource"`
	NodeSelector *map[string]string `json:"node_selector"`
	PvcSize      *int               `json:"pvcSize" default:"10"`
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
	NodeSelector       map[string]any       `json:"nodeSelector"`       // {"env":"flink"}
}

// 主要组装 Name和 Size
func (c *CreateFlinkV12ClusterRequest) NewPVC() map[string]any {
	yaml := resourcePVC
	yaml["metadata"].(map[string]any)["name"] = fmt.Sprintf(pvcName, *c.Name)
	if c.JobManager != nil && c.JobManager.PvcSize != nil {
		yaml["spec"].(map[string]any)["resources"].(map[string]any)["requests"].(map[string]any)["storage"] = fmt.Sprintf("%dGi", *c.JobManager.PvcSize)
	}
	if c.Owner != nil {
		if yaml["metadata"].(map[string]any)["labels"] == nil {
			yaml["metadata"].(map[string]any)["labels"] = map[string]string{}
		}
		yaml["metadata"].(map[string]any)["labels"].(map[string]string)["owner"] = *c.Owner

	}
	return yaml
}

func (c *CreateFlinkV12ClusterRequest) NewService() map[string]any {
	yaml := rerouceJobManagerService
	yaml["metadata"].(map[string]any)["name"] = fmt.Sprintf(jobManagerServiceName, *c.Name)
	yaml["spec"].(map[string]any)["selector"].(map[string]any)["app"] = *c.Name

	if c.Owner != nil {
		if yaml["metadata"].(map[string]any)["labels"] == nil {
			yaml["metadata"].(map[string]any)["labels"] = map[string]string{}
		}
		yaml["metadata"].(map[string]any)["labels"].(map[string]string)["owner"] = *c.Owner
	}
	return yaml
}

// Port 不指定自动分配
func (c *CreateFlinkV12ClusterRequest) NewLBService() map[string]any {
	if c.LoadBalancer == nil {
		return nil
	}

	yaml := rerouceJobManagerLBService
	yaml["metadata"].(map[string]any)["name"] = fmt.Sprintf("%s-lb-service", *c.Name)
	yaml["spec"].(map[string]any)["selector"].(map[string]any)["app"] = *c.Name
	if c.LoadBalancer.Annotations != nil {
		yaml["metadata"].(map[string]any)["annotations"] = c.LoadBalancer.Annotations
	}
	if c.LoadBalancer.Labels != nil {
		yaml["metadata"].(map[string]any)["labels"] = c.LoadBalancer.Labels
	}

	if c.Owner != nil {
		if yaml["metadata"].(map[string]any)["labels"] == nil {
			yaml["metadata"].(map[string]any)["labels"] = map[string]string{}
		}
		yaml["metadata"].(map[string]any)["labels"].(map[string]string)["owner"] = *c.Owner
	}

	return yaml
}

func (c *CreateFlinkV12ClusterRequest) NewConfigMap() map[string]any {
	yaml := resourceFlinkV12ConfigMap
	yaml["metadata"].(map[string]any)["name"] = fmt.Sprintf(configMapV12Name, *c.Name)
	if c.Owner != nil {
		if yaml["metadata"].(map[string]any)["labels"] == nil {
			yaml["metadata"].(map[string]any)["labels"] = map[string]string{}
		}
		yaml["metadata"].(map[string]any)["labels"].(map[string]string)["owner"] = *c.Owner
	}

	defaultConfig := map[string]any{
		"taskmanager.rpc.port":        6122,
		"jobmanager.rpc.port":         6123,
		"blob.server.port":            6124,
		"queryable-state.proxy.ports": 6125,
		"parallelism.default":         1,

		"taskmanager.numberOfTaskSlots":        2,
		"taskmanager.memory.flink.size":        "2048m",
		"taskmanager.memory.task.heap.size":    "3000m",
		"web.upload.dir":                       "/opt/flink/target",
		"jobmanager.rpc.address":               fmt.Sprintf(jobManagerServiceName, *c.Name),
		"jobmanager.memory.flink.size":         "2048m",
		"jobmanager.memory.jvm-metaspace.size": "2048m",
	}
	if c.FlinkConfigRequest != nil {
		for k, v := range c.FlinkConfigRequest {
			defaultConfig[k] = v
		}
	}
	yaml["data"].(map[string]any)["flink-conf.yaml"] = tea.Prettify(defaultConfig)
	return yaml
}

func (c *CreateFlinkV12ClusterRequest) NewJobManagerDeployment() map[string]any {
	yaml := rerouceJobManagerDeployment
	if yaml["metadata"].(map[string]any)["labels"] == nil {
		yaml["metadata"].(map[string]any)["labels"] = map[string]string{}
	}
	yaml["metadata"].(map[string]any)["labels"].(map[string]string)["app"] = *c.Name
	yaml["metadata"].(map[string]any)["name"] = fmt.Sprintf(jobManagerDeploymentName, *c.Name)

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
				"claimName": fmt.Sprintf(pvcName, *c.Name),
			},
		},
		{
			"name": "flink-config",
			"configMap": map[string]any{
				"name": fmt.Sprintf(configMapV12Name, *c.Name),
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
	cpuR := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse(*c.JobManager.Resource.CPU),
		v1.ResourceMemory: resource.MustParse(*c.JobManager.Resource.Memory),
	}
	if c.JobManager.Resource != nil {
		jobContainer["resources"] = map[string]any{
			"requests": v1.ResourceList(cpuR),
			"limits":   v1.ResourceList(cpuR),
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

	sideCarContainer := map[string]any{
		"name":  "busybox",
		"image": "busybox",
		"command": []string{
			"sh",
			"-c",
			"echo hello from sidecar",
		},
	}
	yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"] = []map[string]any{jobContainer, sideCarContainer}

	// nodeSelector 组装
	if c.NodeSelector != nil {
		yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["nodeSelector"] = c.NodeSelector
	}

	return yaml
}

func (c *CreateFlinkV12ClusterRequest) NewTaskManagerDeployment() map[string]any {
	yaml := resourceTaskManagerDeployment

	if yaml["metadata"].(map[string]any)["labels"] == nil {
		yaml["metadata"].(map[string]any)["labels"] = map[string]string{}
	}
	yaml["metadata"].(map[string]any)["labels"].(map[string]string)["app"] = *c.Name
	yaml["metadata"].(map[string]any)["name"] = fmt.Sprintf(taskManagerDeploymentName, *c.Name)

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
			"exec": map[string]any{
				"tcpSocket": map[string]any{
					"port": 6122,
				},
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
		cpuR := v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(*c.TaskManager.Resource.CPU),
			v1.ResourceMemory: resource.MustParse(*c.TaskManager.Resource.Memory),
		}
		taskManagerContainer["resources"] = map[string]any{
			"requests": v1.ResourceList(cpuR),
			"limits":   v1.ResourceList(cpuR),
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
				"name": fmt.Sprintf(configMapV12Name, *c.Name),
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
	yaml["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["nodeSelector"] = c.NodeSelector
	return yaml
}
