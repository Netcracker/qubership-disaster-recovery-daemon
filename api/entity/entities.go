package entity

import "k8s.io/apimachinery/pkg/watch"

const (
	ACTIVE          = "active"
	DISABLED        = "disable"
	STANDBY         = "standby"
	DEGRADED        = "degraded"
	QUEUE           = "queue"
	RUNNING         = "running"
	DONE            = "done"
	DOWN            = "down"
	UP              = "up"
	FAILED          = "failed"
	DeploymentType  = "deployment"
	StatefulsetType = "statefulset"
)

type SwitchoverState struct {
	Mode    string `json:"mode"`
	Status  string `json:"status,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type RequestData struct {
	Mode   string `json:"mode"`
	NoWait *bool  `json:"no-wait,omitempty"`
}

type ControllerRequest struct {
	RequestData
	SwitchoverAnnotation string                 `json:"switchoverAnnotation,omitempty"`
	Status               SwitchoverState        `json:"status"`
	EventType            watch.EventType        `json:"eventType"`
	Object               map[string]interface{} `json:"object"`
}

type ControllerResponse struct {
	SwitchoverState
}

type HealthRequest struct {
	Mode       string
	FullHealth bool
}

type HealthResponse struct {
	Status  string `json:"status"`
	Comment string `json:"comment,omitempty"`
}

type ModeDataUpdate struct {
	Mode       string
	NoWait     bool
	Annotation map[string]string
}

type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
