package usecase

import (
	"errors"
	"fmt"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/api/entity"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/config"
	"strconv"
	"time"
)

const (
	SwitchoverAnnotationKey     = "switchoverRetry"
	CustomResourceNotFoundError = "The custom resource does not contain information about disaster recovery. Error is [%v]"
	updateStatusSleepDuration   = 2 * time.Second
)

func NewSetModeUseCase(crr KubernetesCustomResourceRepo, config config.DisasterRecoveryPath) *SetModeUseCase {
	return &SetModeUseCase{
		crRepo: crr,
		config: config,
	}
}

type SetModeUseCase struct {
	crRepo KubernetesCustomResourceRepo
	config config.DisasterRecoveryPath
}

func (smuc SetModeUseCase) SetDrMode(data entity.RequestData) (entity.SwitchoverState, error) {
	mode := data.Mode
	if mode != entity.ACTIVE && mode != entity.STANDBY && mode != entity.DISABLED {
		return entity.SwitchoverState{Mode: mode,
				Status: entity.FAILED,
				Comment: fmt.Sprintf("'%s' mode is not in the allowed list. Please, use '%s', '%s' or '%s'",
					mode, entity.ACTIVE, entity.STANDBY, entity.DISABLED)},
			fmt.Errorf("illegal mode field value - [%s]", mode)
	}
	drStatus, err := smuc.crRepo.GetDrStatus(smuc.config.StatusPath)
	if err != nil {
		return entity.SwitchoverState{Comment: fmt.Sprintf(CustomResourceNotFoundError, err)}, err
	}

	drMode, err := smuc.crRepo.GetDrMode(smuc.config.ModePath...)
	if err != nil {
		return entity.SwitchoverState{Comment: fmt.Sprintf(CustomResourceNotFoundError, err)}, err
	}

	if drStatus.Status == entity.RUNNING {
		return entity.SwitchoverState{
				Mode:    mode,
				Comment: "The switchover process is in progress. Please, wait until it will be finished"},
			errors.New("switchover process is already in progress")
	}

	if drStatus.Mode == drMode &&
		drStatus.Mode == mode &&
		drStatus.Status == entity.DONE {
		return entity.SwitchoverState{Mode: mode,
			Status:  entity.DONE,
			Comment: "The switchover process has already been done"}, nil
	}

	var noWait bool
	if data.NoWait == nil {
		noWait = true
	} else {
		noWait = *data.NoWait
	}

	err = smuc.crRepo.UpdateStatus(smuc.config.StatusPath,
		entity.SwitchoverState{Mode: drStatus.Mode, Status: entity.QUEUE, Comment: "Switchover is in queue"})
	if err != nil {
		return entity.SwitchoverState{Mode: mode, Comment: err.Error()}, err
	}
	time.Sleep(updateStatusSleepDuration)

	update := entity.ModeDataUpdate{Mode: mode, NoWait: noWait}
	// we should send CR to the reconcile loop if switchover from standby to active was failed
	// and we get the same request again
	if isRetryAction(drMode, drStatus, mode) {
		annotation := make(map[string]string)
		annotation[SwitchoverAnnotationKey] = strconv.Itoa(time.Now().Nanosecond())
		update.Annotation = annotation
	}
	err = smuc.crRepo.UpdateDrMode(smuc.config, update)
	if err != nil {
		return entity.SwitchoverState{Mode: mode, Comment: err.Error()}, err
	}
	return entity.SwitchoverState{Mode: mode}, nil
}

func isRetryAction(drMode string, drStatus entity.SwitchoverState, newMode string) bool {
	return newMode == drMode && drStatus.Status != entity.DONE
}
