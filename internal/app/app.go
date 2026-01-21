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

package app

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/client"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/config"
	v1 "github.com/Netcracker/qubership-disaster-recovery-daemon/internal/controller/http/v1"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/internal/usecase"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/internal/usecase/repo"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/pkg/httpserver"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"os"
)

func Run(cfg *config.Config) {
	serviceGVR := schema.GroupVersionResource{
		Group:    cfg.Group,
		Version:  cfg.Version,
		Resource: cfg.Resource}
	dynClient := client.MakeDynamicClient()
	clientSet := client.MakeKubeClientSet()
	httpClient := configureClient(fmt.Sprintf("%s/ca.crt", cfg.CertsPath))
	kubernetesRepo := repo.NewKubernetesRepo(clientSet, cfg.Namespace)
	crKubernetesRepo := repo.NewKubernetesCustomResourceRepo(dynClient, serviceGVR, cfg.Name, cfg.Namespace)
	restClient := repo.NewRestClient(cfg.AdditionalHealthStatusConfig.Endpoint, httpClient)

	healthUseCase := usecase.NewHealthUseCase(kubernetesRepo, crKubernetesRepo, cfg.HealthConfig, restClient)
	readStateUseCase := usecase.NewReadModeUseCase(crKubernetesRepo, cfg.DisasterRecoveryPath)
	setModeUseCase := usecase.NewSetModeUseCase(crKubernetesRepo, cfg.DisasterRecoveryPath)

	authenticator := v1.NewTokenReviewAuthenticator(clientSet, cfg.AuthConfig)

	serverHandler := v1.NewServerHandler(authenticator)
	serverHandler.NewHealthRoute(readStateUseCase)
	serverHandler.NewHealthzRoute(healthUseCase)
	serverHandler.NewReadModeRoute(readStateUseCase)
	serverHandler.NewUpdateModeRoute(setModeUseCase)
	httpHandler := serverHandler.BuildHandler()
	_ = httpserver.StartServer(httpHandler, cfg.ServerConfig)
}

func configureClient(certificateFilePath string) http.Client {
	httpClient := http.Client{}
	if _, err := os.Stat(certificateFilePath); errors.Is(err, os.ErrNotExist) {
		return httpClient
	}
	caCert, err := os.ReadFile(certificateFilePath)
	if err != nil {
		return httpClient
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}
	return httpClient
}
