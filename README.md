

The `Qubership Disaster Recovery Daemon` (DRD) is a service that establishes communication between the Site Manager
and the current cluster operator or disaster recovery controller.

DRD provides the following features:
* [Disaster Recovery Server](#disaster-recovery-server) implements Site Manager contract and manage current mode in DR resource.
* [Disaster Recovery Controller](#disaster-recovery-controller) provides ability to implement DR controller for services without operator.

![DRD](./pic/drd.png)

Example of DRD chart template is presented [here](https://github.com/Netcracker/qubership-opensearch/blob/main/operator/charts/helm/opensearch-service/templates/operator/deployment.yaml#L71). 

# Disaster Recovery Server

## Common Information

DRD provides all REST endpoints to satisfy the `Site Manager` contract and
takes data from Kubernetes Custom Resource, Kubernetes API. By default, the cluster operator manages this Custom Resource 
and contains service switchover logic. DRD just triggers the cluster operator via Custom Resource changes. 
DRD is delivered as a docker image and has a list of environment variables to configure it. 
DRD can be deployed in the Kubernetes as a separated pod or as a side container for the operator pod.

## Environment variables

| Name                                  | Format                                                                                                                                                                                  | Description                                                                                                                                                                                                                                                                                                                                                 | Example                                             | Required                                                    |
|---------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|-------------------------------------------------------------|
| NAMESPACE                             | A string.                                                                                                                                                                               | The name of service namespace.                                                                                                                                                                                                                                                                                                                              | rabbitmq-service                                    | `true`                                                      |
| RESOURCE_FOR_DR                       | Four words in a single string separated by a single space.                                                                                                                              | This parameter specifies four values to find Kubernetes Custom Resource. These values are group, version, resource, and name of Custom Resource. First word, group, can be empty `""`.                                                                                                                                                                      | netcracker.com v2 rabbitmqservices rabbitmq-service | `true`                                                      |
| USE_DEFAULT_PATHS                     | A single boolean word.                                                                                                                                                                  | If this parameter is `true` the default values are used instead of `DISASTER_RECOVERY_*` environment variable values.                                                                                                                                                                                                                                       | true                                                | `false`                                                     |
| DISASTER_RECOVERY_MODE_PATH           | Several words separated by a dot.                                                                                                                                                       | This parameter specifies the path to disaster recovery `mode` field in Custom Resource.                                                                                                                                                                                                                                                                     | spec.disasterRecovery.mode                          | `true` if `USE_DEFAULT_PATHS` variable is not set to `true` |
| DISASTER_RECOVERY_NOWAIT_PATH         | Several words separated by a dot.                                                                                                                                                       | This parameter specifies the path to disaster recovery `no-wait` field in Custom Resource.                                                                                                                                                                                                                                                                  | spec.disasterRecovery.noWait                        | `true` if `USE_DEFAULT_PATHS` variable is not set to `true` |
| DISASTER_RECOVERY_STATUS_MODE_PATH    | Several words separated by a dot.                                                                                                                                                       | This parameter specifies the path to disaster recovery status `mode` field in Custom Resource.                                                                                                                                                                                                                                                              | status.disasterRecoveryStatus.mode                  | `true` if `USE_DEFAULT_PATHS` variable is not set to `true` |
| DISASTER_RECOVERY_STATUS_STATUS_PATH  | Several words separated by a dot.                                                                                                                                                       | This parameter specifies the path to disaster recovery status `status` field in Custom Resource.                                                                                                                                                                                                                                                            | status.disasterRecoveryStatus.status                | `true` if `USE_DEFAULT_PATHS` variable is not set to `true` |
| DISASTER_RECOVERY_STATUS_COMMENT_PATH | Several words separated by a dot.                                                                                                                                                       | This parameter specifies the path to disaster recovery status `comment` field in Custom Resource.                                                                                                                                                                                                                                                           | status.disasterRecoveryStatus.comment               | `false`                                                     |
| DISASTER_RECOVERY_NOWAIT_AS_STRING    | A single boolean word.                                                                                                                                                                  | If this parameter is `true` the disaster recovery daemon uses string values for `no-wait` parameter, otherwise boolean value is used.                                                                                                                                                                                                                       | false                                               | `false`                                                     |
| HEALTH_MAIN_SERVICES_ACTIVE           | Several word pairs separated by commas. Each pair contains two words separated by a single space. The first word is a Kubernetes workload type and the second one is the workload name. | This parameter specifies the main services for the health check on active side.                                                                                                                                                                                                                                                                             | deployment kafka-1,deployment kafka-2               | `true`                                                      |
| HEALTH_ADDITIONAL_SERVICES_ACTIVE     | Several word pairs separated by commas. Each pair contains two words separated by a single space. The first word is a Kubernetes workload type and the second one is the workload name. | This parameter specifies the additional services for the health check on active side.                                                                                                                                                                                                                                                                       | deployment rabbitmq-backup-daemon                   | `false`                                                     |
| HEALTH_MAIN_SERVICES_STANDBY          | Several word pairs separated by commas. Each pair contains two words separated by a single space. The first word is a Kubernetes workload type and the second one is the workload name. | This parameter specifies the main services for the health check on standby side. If the parameter is empty or is absent, the health status will be always `UP` on standby side.                                                                                                                                                                             | deployment kafka-1,deployment kafka-2               | `false`                                                     |
| HEALTH_ADDITIONAL_SERVICES_STANDBY    | Several word pairs separated by commas. Each pair contains two words separated by a single space. The first word is a Kubernetes workload type and the second one is the workload name. | This parameter specifies the additional services for the health check on standby side.                                                                                                                                                                                                                                                                      | deployment rabbitmq-backup-daemon                   | `false`                                                     |
| HEALTH_MAIN_SERVICES_DISABLED         | Several word pairs separated by commas. Each pair contains two words separated by a single space. The first word is a Kubernetes workload type and the second one is the workload name. | This parameter specifies the main services for the health check on `disable` side. If the parameter is empty or is absent, the health status will be always `UP` on `disable` side.                                                                                                                                                                         | deployment kafka-1,deployment kafka-2               | `false`                                                     |
| HEALTH_ADDITIONAL_SERVICES_DISABLED   | Several word pairs separated by commas. Each pair contains two words separated by a single space. The first word is a Kubernetes workload type and the second one is the workload name. | This parameter specifies the additional services for the health check on `disable` side.                                                                                                                                                                                                                                                                    | deployment rabbitmq-backup-daemon                   | `false`                                                     |
| SITE_MANAGER_SERVICE_ACCOUNT_NAME     | A single word.                                                                                                                                                                          | This parameter specifies the Site Manager service account name.                                                                                                                                                                                                                                                                                             | site-manager                                        | `false`                                                     |
| SITE_MANAGER_NAMESPACE                | A single word.                                                                                                                                                                          | This parameter specifies the Site Manager namespace.                                                                                                                                                                                                                                                                                                        | site-manager                                        | `false`                                                     |
| SITE_MANAGER_CUSTOM_AUDIENCE          | A single word.                                                                                                                                                                          | This parameter specifies the Site Manager custom audience applied for token during authntication.                                                                                                                                                                                                                                                           | sm-services                                         | `false`                                                     |
| SERVER_PORT                           | A number.                                                                                                                                                                               | This parameter specifies the DRD server port. The default value is `8068`.                                                                                                                                                                                                                                                                                  | 8069                                                | `false`                                                     |
| ADDITIONAL_HEALTH_ENDPOINT            | A string.                                                                                                                                                                               | This parameter specifies additional health endpoint. The endpoint response contains information about full cluster health state (if `EXTERNAL_FULL_HEALTH_ENABLED` is `true`) or additional cluster health state (if `EXTERNAL_FULL_HEALTH_ENABLED` is `false`). In the second case, the result will be calculate as `HEALTH_ADDITIONAL_SERVICES` variable. | http://(POD_IP):8069/healthz                        | `false`                                                     |
| EXTERNAL_FULL_HEALTH_ENABLED          | A boolean string.                                                                                                                                                                       | If this parameter is `true` the `ADDITIONAL_HEALTH_ENDPOINT` variable will be used as external full health endpoint. In this case all `HEALTH_*` environment variables are not necessary.                                                                                                                                                                   | true                                                | `false`                                                     |
| TLS_ENABLED                           | A boolean string.                                                                                                                                                                       | If this parameter is `true` TLS will be enabled for DRD container.                                                                                                                                                                                                                                                                                          | false                                               | `false`                                                     |
| CERTS_PATH                            | Path string.                                                                                                                                                                            | This parameter specifies path to folder with TLS certificates in DRD container.                                                                                                                                                                                                                                                                             | /tls/                                               | `false`                                                     |
| CIPHER_SUITES                         | Comma-separated list of strings. Each word is suite name supported by GO e.g. `TLS_RSA_WITH_3DES_EDE_CBC_SHA`                                                                           | This parameter specifies the list of cipher suites that are used to negotiate the security settings for a network connection using TLS or SSL network protocol                                                                                                                                                                                              | ""                                                  | `false`                                                     |
| TREAT_STATUS_AS_FIELD                 | A boolean.                                                                                                                                                                              | This parameter specifies whether resource status should be treated as field. It is necessary when initially `DISASTER_RECOVERY_STATUS_STATUS_PATH` does not have Status sub-resource. In that case status is set as a field to chosen resource. For example, it may be applicable for some of custom resources or ConfigMaps.                               | false                                               | `false`                                                     |

## REST API

DRD REST server provides three methods of interaction:

* `GET` `healthz` method allows finding out the state of the current cluster side.

  ```
  curl -XGET localhost:8068/healthz
  ```

  Where `8068` is the default server port.

  The response to such a request is as follows:

  ```json
  {"status":"up"}
  ```

  Where:
    * `status` is the current state of the cluster side. The four possible status values are as follows:
        * `up` - All service's workloads are ready.
        * `degraded` - Some of the service's workloads (the main health service or additional health service) are not ready.
        * `down` - The main health service is down.
        * `disabled` - The service is switched off.

* `GET` `sitemanager` method allows finding out the mode of the current cluster side and the actual state of the switchover procedure.

  ```
  curl -XGET localhost:8068/sitemanager
  ```

  Where `8068` is the default server port.

  The response to such a request is as follows:

  ```json
  {"mode":"standby","status":"done"}
  ```

  Where:
    * `mode` is the mode in which the cluster side is deployed. The possible mode values are as follows:
        * `active` - The service accepts external requests from clients.
        * `standby` - The service does not accept external requests from clients.
        * `disabled` - The service does not accept external requests from clients.
    * `status` is the current state of switchover for the service cluster side. The three possible status values are as follows:
        * `running` - The switchover is in progress.
        * `done` - The switchover is successful.
        * `failed` - Something went wrong during the switchover.
    * `comment` is the message which contains a detailed description of the problem.

* `POST` `sitemanager` method allows switching mode for the current side of the service cluster.

  ```
  curl -XPOST -H "Content-Type: application/json" localhost:8068/sitemanager -d '{"mode":"<MODE>"}'
  ```

  Where:
    * Where `8068` is the default server port.
    * `<MODE>` is the mode to be applied to the cluster side. The possible mode values are as follows:
        * `active` - The service accepts external requests from clients.
        * `standby` - The service does not accept external requests from clients.
        * `disabled` - The service does not accept external requests from clients.

  The response to such a request is as follows:

  ```json
  {"mode":"standby"}
  ```

  Where:
    * `mode` is the mode that is applied to the cluster side. The possible values are `active`, `standby`, and `disabled`.
    * `status` is the state of the request on the REST server. The only possible value is `failed`, when something goes wrong while processing the request.
    * `comment` is the message which contains a detailed description of the problem and is only filled out if the `status` value is `failed`.

# Authentication

All the DRD SM endpoints can be secured via Kubernetes JWT Service Account Tokens. A Site Manager Kubernetes token should be specified in the Request Header.
Examples for DRD REST endpoints:

  ```
  curl -XGET -H "Authorization: Bearer <TOKEN>" localhost:8068/healthz
  ```

  ```
  curl -XGET -H "Authorization: Bearer <TOKEN>" localhost:8068/sitemanager
  ```

  ```
  curl -XPOST -H "Content-Type: application/json, Authorization: Bearer <TOKEN>" localhost:8068/sitemanager -d '{"mode":"<MODE>"}'
  ```

Where `TOKEN` is a Site Manager Kubernetes token.

Authentication will be enabled only if both `SITE_MANAGER_SERVICE_ACCOUNT_NAME` and `SITE_MANAGER_NAMESPACE` environment variables are specified.
If these environment variables are not specified, the authentication will be disabled.

If authentication is enabled and the `SITE_MANAGER_CUSTOM_AUDIENCE` environment variable is specified, then custom audience
is applied to TokenReview request.

# Example of Configurations

## Custom Resource

Custom Resource with default paths:
```yaml
apiVersion: qubership.org/v1
kind: MyService
metadata:
  name: example-service
  namespace: my-namespace
spec:
  disasterRecovery:
    mode: 'standby'
    noWait: false
status:
  disasterRecoveryStatus:
    comment: 'replication has finished successfully'
    mode: 'standby'
    status: 'done'
```

Environment Variables:
```yaml
- name: NAMESPACE
  value: 'my-namespace'
- name: RESOURCE_FOR_DR
  value: 'qubership.org v1 myservices example-service'
- name: USE_DEFAULT_PATHS
  value: 'true'
- name: HEALTH_MAIN_SERVICES_ACTIVE
  value: 'StatefulSet example-service'
```

## Config Map

Config Map:
```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: example-service-dr-config
  namespace: my-namespace
data:
  mode: 'standby'
  noWait: 'false'
  status_comment: 'replication has finished successfully'
  status_mode: 'standby'
  status_status: 'done'
```

Environment Variables:
```yaml
- name: NAMESPACE
  value: 'my-namespace'
- name: RESOURCE_FOR_DR
  value: '"" v1 configmaps example-service-dr-config'
- name: USE_DEFAULT_PATHS
  value: 'false'
- name: DISASTER_RECOVERY_MODE_PATH
  value: 'data.mode'
- name: DISASTER_RECOVERY_NOWAIT_PATH
  value: 'data.noWait'
- name: DISASTER_RECOVERY_STATUS_MODE_PATH
  value: 'data.status_mode'
- name: DISASTER_RECOVERY_STATUS_STATUS_PATH
  value: 'data.status_status'
- name: DISASTER_RECOVERY_STATUS_COMMENT_PATH
  value: 'data.status_comment'
- name: DISASTER_RECOVERY_NOWAIT_AS_STRING
  value: 'true'
- name: HEALTH_MAIN_SERVICES_ACTIVE
  value: 'StatefulSet example-service'
```

# Disaster Recovery Extension

Disaster Recovery Daemon provides an ability to implement controller for watching changes of Disaster Recovery resource for cases when service does not have its own operator.

DRD extension is a golang application which starts server and controller with function which implements custom DR logic and/or custom health check logic.

## Extension Repository

1. Make a repository or folder for your golang application.
2. In `go.mod` add import `github.com/Netcracker/qubership-disaster-recovery-daemon` with actual [version](https://github.com/Netcracker/qubership-disaster-recovery-daemon/tags).
3. Implement `Main.go` which starts server and controller with function which implements custom DR logic and custom health check logic.
4. Build a Docker image with your golang application.
5. Add a new deployment or container for DRD application to your Helm chart, with corresponding [environment variables](#environment-variables).

## Configuration

To start custom server or controller you need to provide configuration object with necessary parameters.

Configuration can be loaded from some kind of sources with implementing interface configuration loader `config.ConfigLoader`, 
by default DRD provides only environment variables configuration loader `config.DefaultEnvConfigLoader` which uses corresponding [environment variables](#environment-variables).

```go
cfgLoader := config.GetDefaultEnvConfigLoader()
cfg, err := config.NewConfig(cfgLoader)
```

## DR Server and Health

To create and start DR server you need created configuration:

```go
server.NewServer(cfg).Run()
```

You can also specify custom health check function (by default DRD uses pods readiness probes to calculate health):

```go
server.NewServer(cfg).WithHealthFunc(healthFunc, false).Run()
```

Contract for health check function is:

```go
WithHealthFunc(healthFunc func(request entity.HealthRequest) (entity.HealthResponse, error), fullHealth bool)
```

`entity.HealthRequest` contains fields:
* `mode` is a current DR mode for cluster side. Type: `string`. Values: `active`, `standby` or `disabled`). This is required field.

`entity.HealthResponse` contains fields:
* `status` is a result of health check operation. Type: `string`. Values: `up`, `down` or `degraded`). This is required field.
* `comment` is a comment of performing health check operation. Type: `string`.

`fullHealth` function argument means whether health overrides pod readiness health check (if `fullHealth: true`) or should be used as additional health status (if `fullHealth: false`).
If `fullHealth: false` then the following rules are applied:
* All pods ready: UP, additional health: UP -> UP
* Some pods are not ready: DEGRADED, additional health: UP -> DEGRADED
* All pods are down: DOWN, additional health: UP -> DOWN
* All pods ready: UP, additional health: DOWN or DEGRADED -> DEGRADED

**NOTE:** The health check function is an optional feature, if no function is specified the default approach with `HEALTH_MAIN_SERVICES_ACTIVE...` is used.

## DR Controller

To create and start controller you need created configuration and controller func:

```go
controller.NewController(cfg).
        WithFunc(func).
        WithRetry(3, time.Second * 5).
        Run()
```

`WithFunc` takes DR controller function. DR function must be set for DR controller.

Contract for DR controller function is:

```go
func(controllerRequest entity.ControllerRequest) (entity.ControllerResponse, error)
```

`entity.ControllerRequest` contains fields:
* `mode` is a disaster recovery mode from resource. Type: `string`. Values: `active`, `standby` or `disabled`).
* `noWait` is a flag meaning this is failover operation. Type: `bool`.
* `eventType` is a type of resource event. Type: `string`. Values: `ADDED`, `MODIFIED` or `DELETED`).
* `object` is an original DR resource object.

`entity.ControllerResponse` contains fields:
* `mode` is a disaster recovery mode after performing DR operation. Type: `string`. Values: `active`, `standby` or `disabled`). This is required field.
* `status` is a result of performing DR operation. Type: `string`. Values: `done`, `running` or `failed`).  This is required field.
* `comment` is a comment of performing DR operation. Type: `string`.

The result of operation execution will be saved to DR Resource. 

`WithRetry` takes number of attempts and delay for retry policy. 
Controller runs retry only if error happens during function execution, if function returned `failed` status, no retry is called.
If no retry parameters are specified controller calls function only one time.

## Example     

The below is an example of `Main.go` for custom resource [Config Map](#config-map) presented above:

```go
package main

import (
	"context"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/api/entity"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/client"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/config"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/controller"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/server"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
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

	// Start DRD server with custom health function inside. This func calculates only additional health status (fullHealth: false)
	go server.NewServer(cfg).
		WithHealthFunc(func(request entity.HealthRequest) (entity.HealthResponse, error) {
			// Do some health check logic, e,g, using kubernetes client kubeClient.CoreV1()...
			return entity.HealthResponse{Status: entity.UP}, nil
		}, false).
		Run()

	// Start DRD controller with external DR function
	controller.NewController(cfg).
		WithFunc(drFunction).
		Run()
}

// DR function implementation
func drFunction(controllerRequest entity.ControllerRequest) (entity.ControllerResponse, error) {
	var configMap v1.ConfigMap
	// Convert unstructured object to expected type (ConfigMap in our case)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(controllerRequest.Object, &configMap)
	if err != nil {
		return entity.ControllerResponse{}, err
	}
	// Do some DR logic
	return entity.ControllerResponse{
		SwitchoverState: entity.SwitchoverState{
			Mode:    controllerRequest.Mode,
			Status:  entity.DONE,
			Comment: "switchover successfully done",
		},
	}, nil
}

```