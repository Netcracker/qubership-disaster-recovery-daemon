// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
