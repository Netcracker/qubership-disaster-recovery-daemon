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

package usecase

import (
	"github.com/Netcracker/disaster-recovery-daemon/api/entity"
	"github.com/Netcracker/disaster-recovery-daemon/config"
	"io"
)

type Health interface {
	GetHealth() (entity.HealthResponse, error)
}

type ReadMode interface {
	GetModeAndStatus() (entity.SwitchoverState, error)
}

type SetMode interface {
	SetDrMode(entity.RequestData) (entity.SwitchoverState, error)
}

type KubernetesRepo interface {
	GetServiceStatus(map[string][]string) (string, error)
}

type KubernetesCustomResourceRepo interface {
	GetDrMode(...string) (string, error)
	GetDrStatus(path config.DisasterRecoveryStatusPath) (entity.SwitchoverState, error)
	GetResourceVersion() (string, error)
	UpdateDrMode(config.DisasterRecoveryPath, entity.ModeDataUpdate) error
	UpdateStatus(config.DisasterRecoveryStatusPath, entity.SwitchoverState) error
}

type RestClient interface {
	SendRequest(string, string, io.Reader) (int, []byte, error)
}
