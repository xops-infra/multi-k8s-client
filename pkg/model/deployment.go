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
	ClusterName *string           `json:"cluster_name" binding:"required"`
	Namespace   *string           `json:"namespace"`
	Labels      map[string]string `json:"labels"`
}

func (req *ApplyDeploymentRequest) NewApplyDeployment() (*appsv1.DeploymentApplyConfiguration, error) {
	if req.ClusterName == nil {
		return nil, fmt.Errorf("name is required")
	}
	if req.Namespace == nil {
		req.Namespace = tea.String(v1.NamespaceDefault)
	}
	deployment := appsv1.Deployment(*req.ClusterName, *req.Namespace).WithLabels(req.Labels)

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
		Force:        true,
	}
}

func (req *ApplyDeploymentRequest) ToCreateOptions() metav1.CreateOptions {
	return metav1.CreateOptions{
		FieldManager: "multi-k8s-client",
	}
}
