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

package client

import (
	"flag"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

var kubeconfig = new(string)

func getKubeConfig() *string {
	if *kubeconfig != "" {
		return kubeconfig
	}
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	return kubeconfig
}

func getConfigurationForKubernetesClient() *rest.Config {
	var config *rest.Config
	var err error
	inClusterConfig := os.Getenv("IN_CLUSTER_CONFIG")
	if inClusterConfig == "" || inClusterConfig == "true" {
		config, err = rest.InClusterConfig()
	} else {
		kubeconfig := getKubeConfig()
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
	if err != nil {
		log.Fatalln(err, "Can not get kubernetes config")
		return nil
	}
	return config
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func MakeDynamicClient() dynamic.Interface {
	config := getConfigurationForKubernetesClient()
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalln(err, "Can not get dynamic kubernetes client")
	}
	return client
}

func MakeKubeClientSet() *kubernetes.Clientset {
	config := getConfigurationForKubernetesClient()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err, "Can not get kubernetes client")
	}
	return clientset
}
