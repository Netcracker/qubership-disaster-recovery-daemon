package repo

import (
	"context"
	"github.com/Netcracker/disaster-recovery-daemon/api/entity"
	"github.com/Netcracker/disaster-recovery-daemon/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
	"strconv"
)

func NewKubernetesCustomResourceRepo(client dynamic.Interface,
	crGVR schema.GroupVersionResource,
	name string,
	namespace string) *KubernetesCustomResourceRepo {
	return &KubernetesCustomResourceRepo{
		client:    client,
		crGVR:     crGVR,
		name:      name,
		namespace: namespace,
	}
}

type KubernetesCustomResourceRepo struct {
	client    dynamic.Interface
	crGVR     schema.GroupVersionResource
	name      string
	namespace string
}

func (kcrr KubernetesCustomResourceRepo) GetDrMode(path ...string) (string, error) {
	cr, err := kcrr.client.
		Resource(kcrr.crGVR).
		Namespace(kcrr.namespace).
		Get(context.TODO(), kcrr.name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	mode, _, err := unstructured.NestedString(cr.Object, path...)
	if err != nil {
		return "", err
	}
	return mode, nil
}

func (kcrr KubernetesCustomResourceRepo) GetDrStatus(path config.DisasterRecoveryStatusPath) (entity.SwitchoverState, error) {
	cr, err := kcrr.client.
		Resource(kcrr.crGVR).
		Namespace(kcrr.namespace).
		Get(context.TODO(), kcrr.name, metav1.GetOptions{})
	if err != nil {
		return entity.SwitchoverState{}, err
	}

	drMode, _, err := unstructured.NestedString(cr.Object, path.ModePath...)
	if err != nil {
		return entity.SwitchoverState{}, err
	}
	drStatus, _, err := unstructured.NestedString(cr.Object, path.StatusPath...)
	if err != nil {
		return entity.SwitchoverState{}, err
	}
	var drComment string
	if len(path.CommentPath) > 0 {
		drComment, _, err = unstructured.NestedString(cr.Object, path.CommentPath...)
		if err != nil {
			return entity.SwitchoverState{}, err
		}
	}

	state := entity.SwitchoverState{}
	state.Mode = drMode
	state.Status = drStatus
	if drComment != "" {
		state.Comment = drComment
	}

	return state, err
}

func (kcrr KubernetesCustomResourceRepo) GetResourceVersion() (string, error) {
	cr, err := kcrr.client.
		Resource(kcrr.crGVR).
		Namespace(kcrr.namespace).
		Get(context.TODO(), kcrr.name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	resourceVersion, _, err := unstructured.NestedString(cr.Object, "metadata", "resourceVersion")
	return resourceVersion, err
}

func (kcrr KubernetesCustomResourceRepo) UpdateDrMode(drPathConfig config.DisasterRecoveryPath,
	update entity.ModeDataUpdate) error {
	log.Printf("Update mode '%+v' for resource '%v %s'", update, kcrr.crGVR, kcrr.name)
	cr, err := kcrr.client.
		Resource(kcrr.crGVR).
		Namespace(kcrr.namespace).
		Get(context.TODO(), kcrr.name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	err = unstructured.SetNestedField(cr.Object, update.Mode, drPathConfig.ModePath...)
	if err != nil {
		return err
	}
	var noWait interface{}
	if drPathConfig.NoWaitAsString {
		noWait = strconv.FormatBool(update.NoWait)
	} else {
		noWait = update.NoWait
	}
	err = unstructured.SetNestedField(cr.Object, noWait, drPathConfig.NoWaitPath...)
	if err != nil {
		return err
	}
	if update.Annotation != nil {
		annotations := cr.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		for key, value := range update.Annotation {
			annotations[key] = value
		}
		cr.SetAnnotations(annotations)
	}

	_, err = kcrr.client.
		Resource(kcrr.crGVR).
		Namespace(kcrr.namespace).
		Update(context.TODO(), cr, metav1.UpdateOptions{})
	return err
}

func (kcrr KubernetesCustomResourceRepo) UpdateStatus(drStatusPath config.DisasterRecoveryStatusPath,
	update entity.SwitchoverState) error {
	log.Printf("Update status '%+v' for resource '%v %s'", update, kcrr.crGVR, kcrr.name)
	cr, err := kcrr.client.
		Resource(kcrr.crGVR).
		Namespace(kcrr.namespace).
		Get(context.TODO(), kcrr.name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	err = unstructured.SetNestedField(cr.Object, update.Mode, drStatusPath.ModePath...)
	if err != nil {
		return err
	}
	err = unstructured.SetNestedField(cr.Object, update.Status, drStatusPath.StatusPath...)
	if err != nil {
		return err
	}
	err = unstructured.SetNestedField(cr.Object, update.Comment, drStatusPath.CommentPath...)
	if err != nil {
		return err
	}
	if drStatusPath.TreatStatusAsField {
		_, err = kcrr.client.
			Resource(kcrr.crGVR).
			Namespace(kcrr.namespace).
			Update(context.TODO(), cr, metav1.UpdateOptions{})
	} else {
		_, err = kcrr.client.
			Resource(kcrr.crGVR).
			Namespace(kcrr.namespace).
			UpdateStatus(context.TODO(), cr, metav1.UpdateOptions{})
	}
	return err
}
