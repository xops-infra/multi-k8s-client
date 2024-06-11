package model

import (
	"encoding/json"
	"fmt"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/mitchellh/mapstructure"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type EmptyDir struct{}

type HostPath struct {
	Path *string `json:"path"`
}

type PersistentVolumeClaim struct {
	ClaimName *string `json:"claim_name"`
	ReadOnly  *bool   `json:"read_only"`
}

type Item struct {
	Key  *string `json:"key"`
	Path *string `json:"path"`
}
type ConfigMap struct {
	Name  *string `json:"name"`
	Items []Item  `json:"items"`
}

type Volume struct {
	Name                  *string                `json:"name" binding:"required"`
	EmptyDir              *EmptyDir              `json:"empty_dir"`
	HostPath              *HostPath              `json:"host_path"`
	PersistentVolumeClaim *PersistentVolumeClaim `json:"persistent_volume_claim"`
	ConfigMap             *ConfigMap             `json:"config_map"`
}

type ResourceList v1.ResourceList

type Resource struct {
	Limits   ResourceList `json:"limits"`
	Requests ResourceList `json:"requests"`
}

type ContainerPort struct {
	Name          *string `json:"name"`
	ContainerPort *int32  `json:"container_port"`
}

type Exec v1.ExecAction
type HTTPGet v1.HTTPGetAction
type TCPSocket v1.TCPSocketAction
type GRPCAction v1.GRPCAction

type LivenessProbe struct {
	InitialDelaySeconds *int32      `json:"initial_delay_seconds"`
	TimeoutSeconds      *int32      `json:"timeout_seconds"`
	PeriodSeconds       *int32      `json:"period_seconds"`
	SuccessThreshold    *int32      `json:"success_threshold"`
	FailureThreshold    *int32      `json:"failure_threshold"`
	Exec                *Exec       `json:"exec"`
	HTTPGet             *HTTPGet    `json:"http_get"`
	TCPSocket           *TCPSocket  `json:"tcp_socket"`
	GRPC                *GRPCAction `json:"grpc"`
}

type Container struct {
	Name          *string           `json:"name" binding:"required"`
	Image         *string           `json:"image" binding:"required"`
	Args          []string          `json:"args"`
	Resource      Resource          `json:"resource"`
	Ports         []ContainerPort   `json:"ports"`
	Env           map[string]string `json:"env"`
	LivenessProbe *LivenessProbe    `json:"liveness_probe"`
}

type Spec struct {
	NodeSelector map[string]string `json:"node_selector"`
	Volumes      []Volume          `json:"volumes"`
	Containers   []Container       `json:"containers"`
}

type ApplyDeploymentRequest struct {
	Name        *string           `json:"name" binding:"required"`
	Namespace   *string           `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Replicas    *int32            `json:"replicas" default:"1"`
	Spec        *Spec             `json:"spec"`
}

func (req *ApplyDeploymentRequest) NewApplyDeployment() (*appsv1.DeploymentApplyConfiguration, error) {
	if req.Name == nil {
		return nil, fmt.Errorf("name is required")
	}
	if req.Namespace == nil {
		req.Namespace = tea.String(v1.NamespaceDefault)
	}
	deployment := appsv1.Deployment(*req.Name, *req.Namespace).WithLabels(req.Labels).WithAnnotations(req.Annotations)
	// Spec volumes
	var volumes []*corev1.VolumeApplyConfiguration
	for _, v := range req.Spec.Volumes {
		val := corev1.Volume().WithName(*v.Name)
		if v.EmptyDir != nil {
			val = val.WithEmptyDir(corev1.EmptyDirVolumeSource())
		}
		if v.HostPath != nil {
			val = val.WithHostPath(corev1.HostPathVolumeSource().WithPath(*v.HostPath.Path))
		}
		if v.PersistentVolumeClaim != nil {
			val = val.WithPersistentVolumeClaim(corev1.PersistentVolumeClaimVolumeSource().WithClaimName(*v.PersistentVolumeClaim.ClaimName).WithReadOnly(*v.PersistentVolumeClaim.ReadOnly))
		}
		if v.ConfigMap != nil {
			var items []*corev1.KeyToPathApplyConfiguration
			for _, item := range v.ConfigMap.Items {
				if item.Key == nil || item.Path == nil {
					return nil, fmt.Errorf("key or path is required")
				}
				items = append(items, corev1.KeyToPath().WithKey(*item.Key).WithPath(*item.Path))
			}
			val = val.WithConfigMap(corev1.ConfigMapVolumeSource().WithName(*v.ConfigMap.Name).WithItems(items...))
		}
		volumes = append(volumes, val)
	}
	// spec containers
	var containers []*corev1.ContainerApplyConfiguration
	for _, c := range req.Spec.Containers {
		container := corev1.Container().WithName(*c.Name).WithImage(*c.Image)
		if c.Args != nil {
			container = container.WithArgs(c.Args...)
		}
		if c.Resource.Limits != nil {
			container = container.WithResources(corev1.ResourceRequirements().WithLimits(v1.ResourceList(c.Resource.Limits)))
		}
		if c.Resource.Requests != nil {
			container = container.WithResources(corev1.ResourceRequirements().WithRequests(v1.ResourceList(c.Resource.Requests)))
		}
		var ports []*corev1.ContainerPortApplyConfiguration
		if c.Ports != nil {
			for _, p := range c.Ports {
				port := corev1.ContainerPort().WithName(*p.Name).WithContainerPort(*p.ContainerPort)
				ports = append(ports, port)
			}
			container = container.WithPorts(ports...)
		}
		if c.Env != nil {
			for k, v := range c.Env {
				container = container.WithEnv(corev1.EnvVar().WithName(k).WithValue(v))
			}
		}
		if c.LivenessProbe != nil {
			if c.LivenessProbe.Exec != nil {
				container = container.WithLivenessProbe(corev1.Probe().WithExec(corev1.ExecAction().WithCommand(c.LivenessProbe.Exec.Command...)))
			}
			if c.LivenessProbe.HTTPGet != nil {
				container = container.WithLivenessProbe(corev1.Probe().WithHTTPGet(corev1.HTTPGetAction().WithPath(c.LivenessProbe.HTTPGet.Path).WithPort(c.LivenessProbe.HTTPGet.Port).WithScheme(v1.URIScheme(c.LivenessProbe.HTTPGet.Scheme))))
			}
			if c.LivenessProbe.TCPSocket != nil {
				container = container.WithLivenessProbe(corev1.Probe().WithTCPSocket(corev1.TCPSocketAction().WithPort(c.LivenessProbe.TCPSocket.Port)))
			}
			if c.LivenessProbe.GRPC != nil {
				container = container.WithLivenessProbe(corev1.Probe().WithGRPC(corev1.GRPCAction().WithPort(c.LivenessProbe.GRPC.Port)))
			}
		}
		containers = append(containers, container)
	}

	// Spec nodeSelector
	podSpec := corev1.PodSpec().WithNodeSelector(req.Spec.NodeSelector).WithVolumes(volumes...).WithContainers(containers...)
	template := corev1.PodTemplateSpec().WithName(*req.Name).WithLabels(req.Labels).WithAnnotations(req.Annotations).WithSpec(podSpec)
	// Spec
	deployment.Spec = appsv1.DeploymentSpec().WithReplicas(1).WithTemplate(template)
	if req.Replicas != nil {
		deployment.Spec.Replicas = req.Replicas
	}

	return deployment, nil
}

func NewDeploymentApplyFromMap(data map[string]any) (*appsv1.DeploymentApplyConfiguration, error) {
	var deployment appsv1.DeploymentApplyConfiguration

	if err := mapstructure.Decode(data, &deployment); err != nil {
		return nil, err
	}
	return &deployment, nil
}

func NewDeploymentCreateFromMap(data map[string]any) (*appv1.Deployment, error) {
	var deployment appv1.Deployment

	dataj, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(dataj, &deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (req *ApplyDeploymentRequest) ToApplyOptions() metav1.ApplyOptions {
	return metav1.ApplyOptions{
		FieldManager: "multi-k8s-client",
	}
}

func (req *ApplyDeploymentRequest) ToCreateOptions() metav1.CreateOptions {
	return metav1.CreateOptions{
		FieldManager: "multi-k8s-client",
	}
}
