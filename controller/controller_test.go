package controller

import (
	"fmt"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/api/entity"
	repoConfig "git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/config"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/internal/usecase"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"strconv"
	"testing"
	"time"
)

const NoChanged = ""

var emptyControllerFunc = func(request entity.ControllerRequest) (entity.ControllerResponse, error) {
	return entity.ControllerResponse{}, nil
}

type TestEnvProvider struct {
	envs map[string]string
}

func (t TestEnvProvider) GetEnv(key, fallback string) string {
	if value, ok := t.envs[key]; ok {
		return value
	}
	return fallback
}

type TestCustomResourceRepo struct {
	ResourceVersion int
	ResultStatus    entity.SwitchoverState
}

func (t *TestCustomResourceRepo) GetDrMode(...string) (string, error) {
	return "", nil
}

func (t *TestCustomResourceRepo) GetDrStatus(path repoConfig.DisasterRecoveryStatusPath) (entity.SwitchoverState, error) {
	return t.ResultStatus, nil
}

func (t *TestCustomResourceRepo) GetResourceVersion() (string, error) {
	return strconv.Itoa(t.ResourceVersion), nil
}

func (t *TestCustomResourceRepo) UpdateDrMode(repoConfig.DisasterRecoveryPath, entity.ModeDataUpdate) error {
	return nil
}

func (t *TestCustomResourceRepo) UpdateStatus(path repoConfig.DisasterRecoveryStatusPath, state entity.SwitchoverState) error {
	t.ResultStatus = state
	t.ResourceVersion++
	return nil
}

func TestController_handleNullEvents(t *testing.T) {
	ctr := buildController(emptyControllerFunc)
	ctr.handleEvent(nil, nil, watch.Added)
	ctr.handleEvent(nil, nil, watch.Modified)
	ctr.handleEvent(nil, nil, watch.Deleted)

	state, _ := ctr.crKubernetesRepo.GetDrStatus(ctr.config.StatusPath)
	version, _ := ctr.crKubernetesRepo.GetResourceVersion()
	assert.Equalf(t, NoChanged, state.Mode, "Mode shouldn't change")
	assert.Equalf(t, NoChanged, state.Status, "Status shouldn't change")
	assert.Equalf(t, "1", version, "Version shouldn't change")
}

func TestController_handleAddSuccessful(t *testing.T) {
	ctr := buildController(buildControllerFunc(entity.ACTIVE, entity.DONE, false))
	newResource := buildCustomResource(entity.ACTIVE, "", "", "")

	ctr.handleEvent(nil, newResource, watch.Added)

	state, _ := ctr.crKubernetesRepo.GetDrStatus(ctr.config.StatusPath)
	version, _ := ctr.crKubernetesRepo.GetResourceVersion()
	assert.Equalf(t, entity.ACTIVE, state.Mode, "Mode should change")
	assert.Equalf(t, entity.DONE, state.Status, "Status should change")
	assert.Equalf(t, "3", version, "Version should change")
}

func TestController_handleAddFailed(t *testing.T) {
	ctr := buildController(buildControllerFunc(entity.ACTIVE, entity.FAILED, false))
	newResource := buildCustomResource(entity.ACTIVE, "", "", "")

	ctr.handleEvent(nil, newResource, watch.Added)

	state, _ := ctr.crKubernetesRepo.GetDrStatus(ctr.config.StatusPath)
	version, _ := ctr.crKubernetesRepo.GetResourceVersion()
	assert.Equalf(t, entity.ACTIVE, state.Mode, "Mode should change")
	assert.Equalf(t, entity.FAILED, state.Status, "Status should change")
	assert.Equalf(t, "3", version, "Version should change")
}

func TestController_handleAddError(t *testing.T) {
	ctr := buildController(buildControllerFunc(entity.ACTIVE, entity.DONE, true))
	newResource := buildCustomResource(entity.ACTIVE, "", "", "")

	ctr.handleEvent(nil, newResource, watch.Added)

	state, _ := ctr.crKubernetesRepo.GetDrStatus(ctr.config.StatusPath)
	version, _ := ctr.crKubernetesRepo.GetResourceVersion()
	assert.Equalf(t, entity.ACTIVE, state.Mode, "Mode should change")
	assert.Equalf(t, entity.FAILED, state.Status, "Status should change")
	assert.Equalf(t, "4", version, "Version should change and include 1 retry")
}

func TestController_handleAddEmpty(t *testing.T) {
	ctr := buildController(buildControllerFunc(entity.ACTIVE, entity.DONE, false))
	newResource := &unstructured.Unstructured{}

	ctr.handleEvent(nil, newResource, watch.Added)

	state, _ := ctr.crKubernetesRepo.GetDrStatus(ctr.config.StatusPath)
	version, _ := ctr.crKubernetesRepo.GetResourceVersion()
	assert.Equalf(t, NoChanged, state.Mode, "Mode shouldn't change")
	assert.Equalf(t, NoChanged, state.Status, "Status shouldn't change")
	assert.Equalf(t, "1", version, "Version shouldn't change")
}

func TestController_handleSuccessfulModify(t *testing.T) {
	ctr := buildController(buildControllerFunc(entity.ACTIVE, entity.DONE, false))
	oldResource := buildCustomResource(entity.STANDBY, entity.STANDBY, entity.QUEUE, "")
	newResource := buildCustomResource(entity.ACTIVE, entity.STANDBY, entity.QUEUE, "")

	ctr.handleEvent(oldResource, newResource, watch.Modified)

	state, _ := ctr.crKubernetesRepo.GetDrStatus(ctr.config.StatusPath)
	version, _ := ctr.crKubernetesRepo.GetResourceVersion()
	assert.Equalf(t, entity.ACTIVE, state.Mode, "Mode should change")
	assert.Equalf(t, entity.DONE, state.Status, "Status should change")
	assert.Equalf(t, "3", version, "Version should change")
}

func TestController_handleFailedModify(t *testing.T) {
	ctr := buildController(buildControllerFunc(entity.ACTIVE, entity.FAILED, false))
	oldResource := buildCustomResource(entity.STANDBY, entity.STANDBY, entity.QUEUE, "")
	newResource := buildCustomResource(entity.ACTIVE, entity.STANDBY, entity.QUEUE, "")

	ctr.handleEvent(oldResource, newResource, watch.Modified)

	state, _ := ctr.crKubernetesRepo.GetDrStatus(ctr.config.StatusPath)
	version, _ := ctr.crKubernetesRepo.GetResourceVersion()
	assert.Equalf(t, entity.ACTIVE, state.Mode, "Mode should change")
	assert.Equalf(t, entity.FAILED, state.Status, "Status should change")
	assert.Equalf(t, "3", version, "Version should change")
}

func TestController_handleModifyOnlyQueue(t *testing.T) {
	ctr := buildController(buildControllerFunc(entity.ACTIVE, entity.DONE, false))
	oldResource := buildCustomResource(entity.ACTIVE, entity.ACTIVE, entity.DONE, "")
	newResource := buildCustomResource(entity.ACTIVE, entity.ACTIVE, entity.QUEUE, "")

	ctr.handleEvent(oldResource, newResource, watch.Modified)

	state, _ := ctr.crKubernetesRepo.GetDrStatus(ctr.config.StatusPath)
	version, _ := ctr.crKubernetesRepo.GetResourceVersion()
	assert.Equalf(t, NoChanged, state.Mode, "Mode shouldn't change")
	assert.Equalf(t, NoChanged, state.Status, "Status shouldn't change")
	assert.Equalf(t, "1", version, "Version shouldn't change")
}

func TestController_handleModifyOnlyQueueForFailed(t *testing.T) {
	ctr := buildController(buildControllerFunc(entity.ACTIVE, entity.DONE, false))
	oldResource := buildCustomResource(entity.ACTIVE, entity.ACTIVE, entity.FAILED, "")
	newResource := buildCustomResource(entity.ACTIVE, entity.ACTIVE, entity.QUEUE, "")

	ctr.handleEvent(oldResource, newResource, watch.Modified)

	state, _ := ctr.crKubernetesRepo.GetDrStatus(ctr.config.StatusPath)
	version, _ := ctr.crKubernetesRepo.GetResourceVersion()
	assert.Equalf(t, NoChanged, state.Mode, "Mode shouldn't change")
	assert.Equalf(t, NoChanged, state.Status, "Status shouldn't change")
	assert.Equalf(t, "1", version, "Version shouldn't change")
}

func TestController_handleModifyAnnotationForFailed(t *testing.T) {
	ctr := buildController(buildControllerFunc(entity.ACTIVE, entity.DONE, false))
	oldResource := buildCustomResource(entity.ACTIVE, entity.ACTIVE, entity.QUEUE, "")
	newResource := buildCustomResource(entity.ACTIVE, entity.ACTIVE, entity.QUEUE, "1")

	ctr.handleEvent(oldResource, newResource, watch.Modified)

	state, _ := ctr.crKubernetesRepo.GetDrStatus(ctr.config.StatusPath)
	version, _ := ctr.crKubernetesRepo.GetResourceVersion()
	assert.Equalf(t, entity.ACTIVE, state.Mode, "Mode should change")
	assert.Equalf(t, entity.DONE, state.Status, "Status should change")
	assert.Equalf(t, "3", version, "Version should change")
}

func buildController(controllerFunc func(request entity.ControllerRequest) (entity.ControllerResponse, error)) *Controller {
	envs := make(map[string]string)
	envs["DISASTER_RECOVERY_MODE_PATH"] = "data.mode"
	envs["DISASTER_RECOVERY_NOWAIT_AS_STRING"] = "true"
	envs["DISASTER_RECOVERY_NOWAIT_PATH"] = "data.noWait"
	envs["DISASTER_RECOVERY_STATUS_COMMENT_PATH"] = "data.status_comment"
	envs["DISASTER_RECOVERY_STATUS_MODE_PATH"] = "data.status_mode"
	envs["DISASTER_RECOVERY_STATUS_STATUS_PATH"] = "data.status_status"
	envs["HEALTH_MAIN_SERVICES_ACTIVE"] = "statefulset test"
	envs["IN_CLUSTER_CONFIG"] = "false"
	envs["NAMESPACE"] = "test"
	envs["RESOURCE_FOR_DR"] = `"" v1 configmaps test`
	envs["TREAT_STATUS_AS_FIELD"] = "true"
	envs["USE_DEFAULT_PATHS"] = "false"

	cfgLoader := repoConfig.NewEnvConfigLoader(TestEnvProvider{envs: envs})
	config, _ := repoConfig.NewConfig(cfgLoader)
	crRepo := TestCustomResourceRepo{ResourceVersion: 1, ResultStatus: entity.SwitchoverState{}}
	ctr := &Controller{
		delay:            1 * time.Second,
		attempts:         2,
		config:           config,
		crKubernetesRepo: usecase.KubernetesCustomResourceRepo(&crRepo),
		controllerFunc:   controllerFunc,
	}
	return ctr
}

func buildControllerFunc(resultMode string, resultStatus string, isError bool) func(request entity.ControllerRequest) (entity.ControllerResponse, error) {
	var err error
	if isError {
		err = fmt.Errorf("test error")
	}
	return func(request entity.ControllerRequest) (entity.ControllerResponse, error) {
		return entity.ControllerResponse{SwitchoverState: entity.SwitchoverState{Mode: resultMode, Status: resultStatus}}, err
	}
}

func buildCustomResource(mode string, currentMode string, currentStatus string, switchoverAnnotation string) *unstructured.Unstructured {
	return &unstructured.Unstructured{

		Object: map[string]interface{}{
			"kind":       "ConfigMap",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"name":      "example-service-dr-config",
				"namespace": "my-namespace",
				"annotations": map[string]interface{}{
					usecase.SwitchoverAnnotationKey: switchoverAnnotation,
				},
			},
			"data": map[string]interface{}{
				"mode":           mode,
				"noWait":         "false",
				"status_comment": "",
				"status_mode":    currentMode,
				"status_status":  currentStatus,
			},
		},
	}
}
