package usecase

import (
	"github.com/Netcracker/disaster-recovery-daemon/api/entity"
	"github.com/Netcracker/disaster-recovery-daemon/config"
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
