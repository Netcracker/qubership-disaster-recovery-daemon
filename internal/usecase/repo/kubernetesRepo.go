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

package repo

import (
	"context"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/api/entity"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appsv1clients "k8s.io/client-go/kubernetes/typed/apps/v1"
)

func NewKubernetesRepo(clientSet *kubernetes.Clientset, namespace string) *KubernetesRepo {
	return &KubernetesRepo{
		deploymentsClient:  clientSet.AppsV1().Deployments(namespace),
		statefulSetsClient: clientSet.AppsV1().StatefulSets(namespace),
	}
}

type KubernetesRepo struct {
	deploymentsClient  appsv1clients.DeploymentInterface
	statefulSetsClient appsv1clients.StatefulSetInterface
}

func (kr KubernetesRepo) GetServiceStatus(services map[string][]string) (status string, err error) {
	var readyServiceNumber int
	var allServiceNumber int
	for serviceType, names := range services {
		var tmpReadyServiceNumber int
		var tmpAllServiceNumber int
		switch serviceType {
		case entity.DeploymentType:
			tmpReadyServiceNumber, tmpAllServiceNumber, err = kr.getDeploymentData(names)
		case entity.StatefulsetType:
			tmpReadyServiceNumber, tmpAllServiceNumber, err = kr.getStatefulSetData(names)
		}
		if err != nil {
			return
		}
		readyServiceNumber += tmpReadyServiceNumber
		allServiceNumber += tmpAllServiceNumber
	}
	if readyServiceNumber == 0 {
		status = entity.DOWN
	} else if readyServiceNumber < allServiceNumber {
		status = entity.DEGRADED
	} else {
		status = entity.UP
	}
	return
}

func (kr KubernetesRepo) getDeploymentData(names []string) (readyDeploymentsNumber,
	allDeploymentsNumber int, err error) {
	var deployments []appsv1.Deployment
	for _, name := range names {
		deployment, getErr := kr.deploymentsClient.Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			err = getErr
			return
		}
		deployments = append(deployments, *deployment)
	}
	allDeploymentsNumber = len(deployments)
	for _, deployment := range deployments {
		if kr.isDeploymentReady(deployment) {
			readyDeploymentsNumber += 1
		}
	}
	return
}

func (kr KubernetesRepo) isDeploymentReady(deployment appsv1.Deployment) bool {
	availableReplicas := min(deployment.Status.ReadyReplicas, deployment.Status.UpdatedReplicas)
	return *deployment.Spec.Replicas == availableReplicas && *deployment.Spec.Replicas != 0
}

func (kr KubernetesRepo) getStatefulSetData(names []string) (readyStatefulSetsNumber,
	allStatefulSetsNumber int, err error) {
	var statefulSets []appsv1.StatefulSet
	for _, name := range names {
		sts, getErr := kr.statefulSetsClient.Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			err = getErr
			return
		}
		statefulSets = append(statefulSets, *sts)
	}
	allStatefulSetsNumber = len(statefulSets)
	for _, statefulSet := range statefulSets {
		if kr.isStatefulSetReady(statefulSet) {
			readyStatefulSetsNumber += 1
		}
	}
	return
}

func (kr KubernetesRepo) isStatefulSetReady(statefulSet appsv1.StatefulSet) bool {
	availableReplicas := min(statefulSet.Status.ReadyReplicas, statefulSet.Status.UpdatedReplicas)
	return *statefulSet.Spec.Replicas == availableReplicas && *statefulSet.Spec.Replicas != 0
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
