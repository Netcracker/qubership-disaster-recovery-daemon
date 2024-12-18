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
	"encoding/json"
	"fmt"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/api/entity"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/config"
	"log"
	"net/http"
	"strings"
)

func NewHealthUseCase(kr KubernetesRepo,
	crr KubernetesCustomResourceRepo,
	config config.HealthConfig,
	restClient RestClient) *HealthUseCase {
	return &HealthUseCase{
		k8sRepo:    kr,
		crRepo:     crr,
		config:     config,
		restClient: restClient,
	}
}

type HealthUseCase struct {
	k8sRepo    KubernetesRepo
	crRepo     KubernetesCustomResourceRepo
	config     config.HealthConfig
	restClient RestClient
}

func (hus HealthUseCase) GetHealth() (entity.HealthResponse, error) {
	drStatus, err := hus.crRepo.GetDrStatus(hus.config.DisasterRecoveryStatusPath)
	if err != nil {
		return entity.HealthResponse{}, err
	}
	mode := strings.ToLower(drStatus.Mode)
	if hus.config.AdditionalHealthStatusConfig.FullHealthEnabled && hus.isCustomHealthNeeded() {
		return hus.getCustomHealth(mode)
	}
	switch mode {
	case entity.ACTIVE:
		return hus.getServicesHealth(hus.config.ActiveMainServices, hus.config.ActiveAdditionalServices, mode)
	case entity.STANDBY:
		return hus.getServicesHealth(hus.config.StandbyMainServices, hus.config.StandbyAdditionalServices, mode)
	case entity.DISABLED:
		return hus.getServicesHealth(hus.config.DisableMainServices, hus.config.DisableAdditionalServices, mode)
	default:
		return entity.HealthResponse{}, fmt.Errorf("can't perform health check for the disaster recovery mode - [%s]", mode)
	}
}

func (hus HealthUseCase) getServicesHealth(mainServices map[string][]string,
	additionalServices map[string][]string,
	mode string) (entity.HealthResponse, error) {
	if mainServices == nil {
		return entity.HealthResponse{Status: entity.UP}, nil
	}
	mainServiceStatus, err := hus.k8sRepo.GetServiceStatus(mainServices)
	if err != nil {
		return entity.HealthResponse{}, err
	}
	var additionalServiceStatus string
	if additionalServices != nil {
		additionalServiceStatus, err = hus.k8sRepo.GetServiceStatus(additionalServices)
		if err != nil {
			return entity.HealthResponse{}, err
		}
	}
	additionalHealthStatus := entity.UP
	if hus.isCustomHealthNeeded() {
		healthResponse, err := hus.getCustomHealth(mode)
		if err != nil {
			additionalHealthStatus = entity.DOWN
		} else {
			additionalHealthStatus = healthResponse.Status
		}
	}
	status := getServiceState(mainServiceStatus, additionalServiceStatus, additionalHealthStatus)
	return entity.HealthResponse{Status: status}, nil
}

func (hus HealthUseCase) getCustomHealth(mode string) (entity.HealthResponse, error) {
	healthRequest := entity.HealthRequest{Mode: mode, FullHealth: hus.config.AdditionalHealthStatusConfig.FullHealthEnabled}

	if hus.config.AdditionalHealthStatusConfig.HealthFunc != nil {
		return hus.config.AdditionalHealthStatusConfig.HealthFunc(healthRequest)
	}
	if hus.config.AdditionalHealthStatusConfig.Endpoint != "" {
		return hus.getExternalEndpointHealth(healthRequest), nil
	}
	return entity.HealthResponse{}, fmt.Errorf("can not evaluate health status for empty health function and health endpoint")
}

func (hus HealthUseCase) getExternalEndpointHealth(healthRequest entity.HealthRequest) entity.HealthResponse {
	path := fmt.Sprintf("?mode=%s&fullHealth=%t", healthRequest.Mode, healthRequest.FullHealth)
	statusCode, responseBody, err := hus.restClient.SendRequest(http.MethodGet, path, nil)
	if err == nil && statusCode == 200 {
		response := entity.StatusResponse{}
		if err := json.Unmarshal(responseBody, &response); err != nil {
			log.Println(`Can not evaluate status from additional health endpoint`)
			return entity.HealthResponse{Status: entity.DOWN}
		}
		status := strings.ToLower(response.Status)
		if status != entity.UP && status != entity.DEGRADED && status != entity.DOWN {
			log.Printf("Error! Status response from external full health endpoint must be up, degraded or down. But %s was given ", response.Status)
			return entity.HealthResponse{Status: entity.DOWN}
		}
		if response.Message != "" {
			log.Printf(`Additional service full health status is "%s" with message: "%s"}`, response.Status, response.Message)
		}
		return entity.HealthResponse{Status: response.Status}
	} else {
		log.Printf(`Can not get full health status from additional health endpoint with statusCode: "%d" and error: "%s" `, statusCode, err)
		return entity.HealthResponse{Status: entity.DOWN}
	}
}

func getServiceState(mainServiceState string, additionalServiceState string, additionalHealthStatus string) string {
	if mainServiceState == entity.UP && ((additionalServiceState == entity.DEGRADED || additionalServiceState == entity.DOWN) || (additionalHealthStatus == entity.DEGRADED || additionalHealthStatus == entity.DOWN)) {
		return entity.DEGRADED
	} else {
		return mainServiceState
	}
}

func (hus HealthUseCase) isCustomHealthNeeded() bool {
	return hus.config.AdditionalHealthStatusConfig.HealthFunc != nil || hus.config.AdditionalHealthStatusConfig.Endpoint != ""
}
