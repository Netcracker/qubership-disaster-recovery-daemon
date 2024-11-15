package controller

import (
	"context"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/Netcracker/disaster-recovery-daemon/api/entity"
	"github.com/Netcracker/disaster-recovery-daemon/client"
	"github.com/Netcracker/disaster-recovery-daemon/config"
	"github.com/Netcracker/disaster-recovery-daemon/internal/usecase"
	"github.com/Netcracker/disaster-recovery-daemon/internal/usecase/repo"
	"github.com/avast/retry-go/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var SwitchoverAnnotationKeyPath = []string{"metadata", "annotations", usecase.SwitchoverAnnotationKey}

type Controller struct {
	controllerFunc   func(request entity.ControllerRequest) (entity.ControllerResponse, error)
	config           *config.Config
	resourceVersion  string
	delay            time.Duration
	attempts         uint
	crKubernetesRepo usecase.KubernetesCustomResourceRepo
}

func NewController(config *config.Config) *Controller {
	return &Controller{config: config, delay: time.Second * 5, attempts: 1}
}

func (ctr *Controller) WithFunc(controllerFunc func(request entity.ControllerRequest) (entity.ControllerResponse, error)) *Controller {
	ctr.controllerFunc = controllerFunc
	return ctr
}

func (ctr *Controller) WithRetry(attempts uint, delay time.Duration) *Controller {
	ctr.attempts = attempts
	ctr.delay = delay
	return ctr
}

func (ctr *Controller) Run() {

	if ctr.controllerFunc == nil {
		log.Panic("Unable to run controller without controller function")
	}

	resource := schema.GroupVersionResource{
		Group:    ctr.config.CustomResourceConfig.Group,
		Version:  ctr.config.CustomResourceConfig.Version,
		Resource: ctr.config.CustomResourceConfig.Resource,
	}

	dynClient := client.MakeDynamicClient()
	ctr.crKubernetesRepo = repo.NewKubernetesCustomResourceRepo(dynClient, resource, ctr.config.Name, ctr.config.Namespace)

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return dynClient.Resource(resource).Namespace(ctr.config.Namespace).List(context.TODO(), metav1.ListOptions{
					FieldSelector: fields.OneTermEqualSelector("metadata.name", ctr.config.CustomResourceConfig.Name).String(),
				})
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				return dynClient.Resource(resource).Namespace(ctr.config.Namespace).Watch(context.TODO(), metav1.ListOptions{
					FieldSelector: fields.OneTermEqualSelector("metadata.name", ctr.config.CustomResourceConfig.Name).String(),
				})
			},
		},
		&unstructured.Unstructured{},
		1*time.Hour,
		cache.Indexers{},
	)

	log.Printf("Controller initiating")
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			ctr.handleEvent(old, new, watch.Modified)
		},
		AddFunc: func(obj interface{}) {
			ctr.handleEvent(nil, obj, watch.Added)
		},
	})
	if err != nil {
		log.Panicf("Cannot register event handler function: %v", err)
		return
	}
	stopCh := make(chan struct{})
	defer close(stopCh)
	log.Printf("Controller started")
	informer.Run(stopCh)
	log.Printf("Controller finished")
}

func (ctr *Controller) handleEvent(old interface{}, new interface{}, eventType watch.EventType) {
	if new == nil {
		log.Printf("DR resource is null")
		return
	}
	newResource := new.(*unstructured.Unstructured)
	controllerRequestNew, err := buildControllerRequest(newResource.UnstructuredContent(), ctr.config)
	if err != nil {
		log.Printf("Cannot deserialize DR resource and build DR request: %v", err)
		return
	}

	if old != nil {
		oldResource := old.(*unstructured.Unstructured)
		controllerRequestOld, err := buildControllerRequest(oldResource.UnstructuredContent(), ctr.config)
		if err != nil {
			log.Printf("Cannot deserialize DR resource and build DR request: %v", err)
			return
		}

		if mapsAreEqual(controllerRequestNew.Object, controllerRequestOld.Object) {
			return
		}

		if controllerRequestNew.Mode == controllerRequestOld.Mode &&
			controllerRequestNew.SwitchoverAnnotation == controllerRequestOld.SwitchoverAnnotation {
			log.Printf("Skip event. Old and New resourses are the same: DR mode '%s', switchoverRetry '%s', state: '%+v'",
				controllerRequestNew.Mode, controllerRequestNew.SwitchoverAnnotation, controllerRequestNew.Status)
			return
		}
	}

	err = retry.Do(func() error {
		return ctr.executeDrFunction(controllerRequestNew, eventType)
	},
		retry.Delay(ctr.delay), retry.Attempts(ctr.attempts), retry.OnRetry(func(n uint, err error) {
			ctr.resourceVersion = ""
		}))

	if err != nil {
		log.Printf("Error occurred during performing DR controller function: %v", err)
		controllerResponse := entity.ControllerResponse{
			SwitchoverState: entity.SwitchoverState{
				Mode:    controllerRequestNew.Mode,
				Status:  entity.FAILED,
				Comment: err.Error(),
			},
		}
		err = ctr.crKubernetesRepo.UpdateStatus(ctr.config.DisasterRecoveryStatusPath, controllerResponse.SwitchoverState)
		if err != nil {
			log.Printf("Error: Cannot update resource status due to: %v", err)
		}
		if err := ctr.updateResourceVersion(ctr.crKubernetesRepo); err != nil {
			log.Printf("Error: Cannot update resource status due to: %v", err)
		}
	}
}

func (ctr *Controller) executeDrFunction(controllerRequest entity.ControllerRequest, eventType watch.EventType) error {
	resourceVersion, err := ctr.crKubernetesRepo.GetResourceVersion()
	if err != nil {
		log.Printf("Cannot obtain resource status version: %v", err)
		return err
	}
	if controllerRequest.Mode == controllerRequest.Status.Mode && controllerRequest.Status.Status == entity.DONE {
		log.Printf("Current DR mode is already '%s' and finished successfully", controllerRequest.Mode)
		return nil
	}
	if resourceVersion == ctr.resourceVersion {
		log.Println("Incoming DR resource does not contain changes")
		return nil
	}
	if err := ctr.updateResourceVersion(ctr.crKubernetesRepo); err != nil {
		return err
	}

	log.Printf("New incoming DR request with mode '%s', current status is '%s' ", controllerRequest.Mode, controllerRequest.Status.Status)
	switchoverState := entity.SwitchoverState{
		Mode:    controllerRequest.Mode,
		Status:  entity.RUNNING,
		Comment: "Switchover is in progress",
	}
	err = ctr.crKubernetesRepo.UpdateStatus(ctr.config.DisasterRecoveryStatusPath, switchoverState)
	if err != nil {
		log.Printf("Cannot update resource status due to: %v", err)
		return err
	}
	if err := ctr.updateResourceVersion(ctr.crKubernetesRepo); err != nil {
		return err
	}

	controllerRequest.EventType = eventType
	controllerResponse, err := ctr.controllerFunc(controllerRequest)
	if err != nil {
		log.Printf("Error occurred during execution of DR controller function: %v", err)
		return err
	}
	log.Printf("Switchover finished, status: '%s', comment: '%s'", controllerResponse.Status, controllerResponse.Comment)
	err = ctr.crKubernetesRepo.UpdateStatus(ctr.config.DisasterRecoveryStatusPath, controllerResponse.SwitchoverState)
	if err != nil {
		log.Printf("Cannot update resource status due to: %v", err)
		return err
	}
	if err := ctr.updateResourceVersion(ctr.crKubernetesRepo); err != nil {
		return err
	}
	return nil
}

func (ctr *Controller) updateResourceVersion(crRepo usecase.KubernetesCustomResourceRepo) error {
	resourceVersion, err := crRepo.GetResourceVersion()
	if err != nil {
		log.Printf("Cannot obtain resource status version: %v", err)
		return err
	}
	ctr.resourceVersion = resourceVersion
	return nil
}

func buildControllerRequest(object map[string]interface{}, cfg *config.Config) (entity.ControllerRequest, error) {
	drMode, _, err := unstructured.NestedString(object, cfg.ModePath...)
	if err != nil {
		return entity.ControllerRequest{}, err
	}
	var noWait bool
	if cfg.NoWaitAsString {
		noWaitString, _, err := unstructured.NestedString(object, cfg.NoWaitPath...)
		if err != nil {
			return entity.ControllerRequest{}, err
		}
		noWait, err = strconv.ParseBool(noWaitString)
		if err != nil {
			return entity.ControllerRequest{}, err
		}
	} else {
		noWait, _, err = unstructured.NestedBool(object, cfg.NoWaitPath...)
		if err != nil {
			return entity.ControllerRequest{}, err
		}
	}
	statusMode, _, err := unstructured.NestedString(object, cfg.StatusPath.ModePath...)
	if err != nil {
		return entity.ControllerRequest{}, err
	}
	statusStatus, _, err := unstructured.NestedString(object, cfg.StatusPath.StatusPath...)
	if err != nil {
		return entity.ControllerRequest{}, err
	}
	statusComment, _, err := unstructured.NestedString(object, cfg.StatusPath.CommentPath...)
	if err != nil {
		return entity.ControllerRequest{}, err
	}

	SwitchoverAnnotation, _, err := unstructured.NestedString(object, SwitchoverAnnotationKeyPath...)
	if err != nil {
		return entity.ControllerRequest{}, err
	}

	controllerRequest := entity.ControllerRequest{
		RequestData: entity.RequestData{
			Mode:   drMode,
			NoWait: &noWait,
		},
		SwitchoverAnnotation: SwitchoverAnnotation,
		Status: entity.SwitchoverState{
			Mode:    statusMode,
			Status:  statusStatus,
			Comment: statusComment,
		},
		Object: object,
	}
	return controllerRequest, nil
}

func mapsAreEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for key, valA := range a {
		valB, ok := b[key]
		if !ok || !reflect.DeepEqual(valA, valB) {
			return false
		}
	}
	return true
}
