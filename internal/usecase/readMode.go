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
	"github.com/Netcracker/qubership-disaster-recovery-daemon/api/entity"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/config"
)

func NewReadModeUseCase(crr KubernetesCustomResourceRepo, config config.DisasterRecoveryPath) *ReadModeUseCase {
	return &ReadModeUseCase{
		crRepo: crr,
		config: config,
	}
}

type ReadModeUseCase struct {
	crRepo KubernetesCustomResourceRepo
	config config.DisasterRecoveryPath
}

func (rmuc ReadModeUseCase) GetModeAndStatus() (entity.SwitchoverState, error) {
	return rmuc.crRepo.GetDrStatus(rmuc.config.StatusPath)
}
