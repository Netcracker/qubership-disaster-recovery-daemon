package usecase

import (
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/api/entity"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/config"
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
