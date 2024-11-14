package main

import (
	"context"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/api/entity"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/client"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/config"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/controller"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/server"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"time"
)

func main() {
	// Make a config loader
	cfgLoader := config.GetDefaultEnvConfigLoader()

	// Build a config
	cfg, err := config.NewConfig(cfgLoader)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Easy way to create a kubernetes client if necessary
	kubeClient := client.MakeKubeClientSet()

	// Start DRD server with custom health function inside, which calculates on;y additional health status (fullHealth: false)
	go server.NewServer(cfg).
		WithHealthFunc(func(request entity.HealthRequest) (entity.HealthResponse, error) {
			_, err := kubeClient.CoreV1().Pods("consul-service").Get(context.TODO(), "consul-server-0", metav1.GetOptions{})
			if err != nil {
				log.Printf("Error: %s", err)
			}
			return entity.HealthResponse{Status: entity.DOWN}, nil
		}, false).
		Run()

	// Start DRD controller with external function
	controller.NewController(cfg).
		WithFunc(drFunction).
		WithRetry(3, time.Second*1).
		Run()
}

// DR function
func drFunction(controllerRequest entity.ControllerRequest) (entity.ControllerResponse, error) {
	var configMap v1.ConfigMap
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(controllerRequest.Object, &configMap)
	if err != nil {
		log.Printf("Error: %s", err)
		return entity.ControllerResponse{}, err
	}

	return entity.ControllerResponse{
		SwitchoverState: entity.SwitchoverState{
			Mode:    controllerRequest.Mode,
			Status:  "done",
			Comment: "done",
		},
	}, nil
}
