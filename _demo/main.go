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

package main

import (
	"context"
	"github.com/Netcracker/disaster-recovery-daemon/api/entity"
	"github.com/Netcracker/disaster-recovery-daemon/client"
	"github.com/Netcracker/disaster-recovery-daemon/config"
	"github.com/Netcracker/disaster-recovery-daemon/controller"
	"github.com/Netcracker/disaster-recovery-daemon/server"
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
