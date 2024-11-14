package app

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/client"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/config"
	v1 "git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/internal/controller/http/v1"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/internal/usecase"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/internal/usecase/repo"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/pkg/httpserver"
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
	restClient := repo.NewRestClient(cfg.HealthConfig.AdditionalHealthStatusConfig.Endpoint, httpClient)

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
