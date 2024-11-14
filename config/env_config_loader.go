package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type EnvProvider interface {
	GetEnv(string, string) string
}

type OsEnvProvider struct{}

func (oec OsEnvProvider) GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func NewEnvConfigLoader(envProvider EnvProvider) *DefaultEnvConfigLoader {
	return &DefaultEnvConfigLoader{
		envProvider: envProvider,
	}
}

func GetDefaultEnvConfigLoader() *DefaultEnvConfigLoader {
	return &DefaultEnvConfigLoader{
		envProvider: OsEnvProvider{},
	}
}

type DefaultEnvConfigLoader struct {
	envProvider EnvProvider
}

func (decl DefaultEnvConfigLoader) GetCustomResourceConfig() (*CustomResourceConfig, error) {
	resourceEnv, err := decl.getRequiredEnv("RESOURCE_FOR_DR")
	if err != nil {
		return nil, err
	}
	resourceEnv = strings.ReplaceAll(resourceEnv, "'", "")
	resourceEnv = strings.ReplaceAll(resourceEnv, "\"", "")
	resource := strings.Split(resourceEnv, " ")
	if len(resource) != 4 {
		return nil, errors.New("RESOURCE_FOR_DR environment variable must contain exactly four variables which are separated by a single space")
	}
	namespace, err := decl.getRequiredEnv("NAMESPACE")
	if err != nil {
		return nil, err
	}
	return &CustomResourceConfig{
		Name:      resource[3],
		Namespace: namespace,
		Group:     resource[0],
		Version:   resource[1],
		Resource:  resource[2],
	}, nil
}

func (decl DefaultEnvConfigLoader) GetDisasterRecoveryPaths() (*DisasterRecoveryPath, error) {
	useDefaultPaths := decl.envProvider.GetEnv("USE_DEFAULT_PATHS", "")
	treatStatusAsField, err := strconv.ParseBool(decl.envProvider.GetEnv("TREAT_STATUS_AS_FIELD", "false"))
	if err != nil {
		return nil, err
	}
	if strings.ToLower(useDefaultPaths) == "true" {
		return &DisasterRecoveryPath{
			DisasterRecoveryStatusPath{
				[]string{"status", "disasterRecoveryStatus", "mode"},
				[]string{"status", "disasterRecoveryStatus", "status"},
				[]string{"status", "disasterRecoveryStatus", "comment"},
				treatStatusAsField,
			},
			[]string{"spec", "disasterRecovery", "mode"},
			[]string{"spec", "disasterRecovery", "noWait"},
			false,
		}, nil
	}
	drModePathEnv, err := decl.getRequiredEnv("DISASTER_RECOVERY_MODE_PATH")
	if err != nil {
		return nil, err
	}
	drModePath := strings.Split(drModePathEnv, ".")
	drNoWaitPathEnv, err := decl.getRequiredEnv("DISASTER_RECOVERY_NOWAIT_PATH")
	if err != nil {
		return nil, err
	}
	drNoWaitPath := strings.Split(drNoWaitPathEnv, ".")
	drNoWaitAsString, err := strconv.ParseBool(decl.envProvider.GetEnv("DISASTER_RECOVERY_NOWAIT_AS_STRING", "false"))
	if err != nil {
		return nil, err
	}
	drStatusModePathString, err := decl.getRequiredEnv("DISASTER_RECOVERY_STATUS_MODE_PATH")
	if err != nil {
		return nil, err
	}
	drStatusModePath := strings.Split(drStatusModePathString, ".")
	drStatusStatusPathString, err := decl.getRequiredEnv("DISASTER_RECOVERY_STATUS_STATUS_PATH")
	if err != nil {
		return nil, err
	}
	drStatusStatusPath := strings.Split(drStatusStatusPathString, ".")
	drStatusCommentPathString := decl.envProvider.GetEnv("DISASTER_RECOVERY_STATUS_COMMENT_PATH", "")
	var drStatusCommentPath []string
	if drStatusCommentPathString != "" {
		drStatusCommentPath = strings.Split(drStatusCommentPathString, ".")
	}
	drStatusPath := &DisasterRecoveryStatusPath{
		ModePath:           drStatusModePath,
		StatusPath:         drStatusStatusPath,
		CommentPath:        drStatusCommentPath,
		TreatStatusAsField: treatStatusAsField,
	}

	drp := &DisasterRecoveryPath{
		StatusPath:     *drStatusPath,
		ModePath:       drModePath,
		NoWaitPath:     drNoWaitPath,
		NoWaitAsString: drNoWaitAsString,
	}
	return drp, nil
}

func (decl DefaultEnvConfigLoader) GetHealthConfig() (*HealthConfig, error) {
	activeMainServices, err := decl.getServicesEnv("HEALTH_MAIN_SERVICES_ACTIVE", "deployment", "statefulset")
	if err != nil {
		return nil, err
	}
	if activeMainServices == nil {
		return nil, fmt.Errorf(RequiredEnvTemplatedError, "HEALTH_MAIN_SERVICES_ACTIVE")
	}

	activeAdditionalServices, err := decl.getServicesEnv("HEALTH_ADDITIONAL_SERVICES_ACTIVE", "deployment", "statefulset")
	if err != nil {
		return nil, err
	}

	standbyMainServices, err := decl.getServicesEnv("HEALTH_MAIN_SERVICES_STANDBY", "deployment", "statefulset")
	if err != nil {
		return nil, err
	}
	standbyAdditionalServices, err := decl.getServicesEnv("HEALTH_ADDITIONAL_SERVICES_STANDBY", "deployment", "statefulset")
	if err != nil {
		return nil, err
	}
	disableMainServices, err := decl.getServicesEnv("HEALTH_MAIN_SERVICES_DISABLED", "deployment", "statefulset")
	if err != nil {
		return nil, err
	}
	disableAdditionalServices, err := decl.getServicesEnv("HEALTH_ADDITIONAL_SERVICES_DISABLED", "deployment", "statefulset")
	if err != nil {
		return nil, err
	}
	additionalHealthStatusConfig, err := decl.GetAdditionalHealthStatusConfig()
	if err != nil {
		return nil, err
	}
	return &HealthConfig{
		ActiveMainServices:           activeMainServices,
		ActiveAdditionalServices:     activeAdditionalServices,
		StandbyMainServices:          standbyMainServices,
		StandbyAdditionalServices:    standbyAdditionalServices,
		DisableMainServices:          disableMainServices,
		DisableAdditionalServices:    disableAdditionalServices,
		AdditionalHealthStatusConfig: additionalHealthStatusConfig,
	}, nil
}

func (decl DefaultEnvConfigLoader) GetAuthConfig() (*AuthConfig, error) {
	smsa := decl.envProvider.GetEnv("SITE_MANAGER_SERVICE_ACCOUNT_NAME", "")
	smNamespace := decl.envProvider.GetEnv("SITE_MANAGER_NAMESPACE", "")
	authEnabled := true
	if (smsa != "" && smNamespace == "") || (smsa == "" && smNamespace != "") {
		return nil, errors.New("both SITE_MANAGER_SERVICE_ACCOUNT_NAME and SITE_MANAGER_NAMESPACE must be set")
	} else if smsa == "" && smNamespace == "" {
		authEnabled = false
	}
	smCustomAudience := decl.envProvider.GetEnv("SITE_MANAGER_CUSTOM_AUDIENCE", "")
	return &AuthConfig{
		authEnabled,
		smsa,
		smNamespace,
		smCustomAudience,
	}, nil
}

func (decl DefaultEnvConfigLoader) GetServerConfig() (*ServerConfig, error) {
	suites, err := getCipherSuites(decl)
	if err != nil {
		return nil, err
	}
	tlsEnabled, err := strconv.ParseBool(decl.envProvider.GetEnv("TLS_ENABLED", "false"))
	if err != nil {
		return nil, err
	}
	defaultServerPort := "8080"
	if tlsEnabled {
		defaultServerPort = "8443"
	}
	portEnv := decl.envProvider.GetEnv("SERVER_PORT", defaultServerPort)
	port, err := strconv.Atoi(portEnv)
	if err != nil {
		return nil, err
	}
	certsPath := strings.TrimSuffix(decl.envProvider.GetEnv("CERTS_PATH", "/tls/"), "/")
	return &ServerConfig{
		Port:       port,
		Suites:     suites,
		TLSEnabled: tlsEnabled,
		CertsPath:  certsPath,
	}, nil
}

func getCipherSuites(decl DefaultEnvConfigLoader) ([]uint16, error) {
	allSuites := decl.envProvider.GetEnv("CIPHER_SUITES", "")
	var suites []uint16
	if allSuites != "" {
		suiteNames := strings.Split(allSuites, ",")
		for _, name := range suiteNames {
			if name != "" {
				id, err := checkIfCipherSuiteSupported(name)
				if err != nil {
					return nil, err
				}
				suites = append(suites, id)
			}
		}
	}
	return suites, nil
}

func checkIfCipherSuiteSupported(name string) (uint16, error) {
	for _, supportedSuite := range tls.CipherSuites() {
		if supportedSuite.Name == name {
			return supportedSuite.ID, nil
		}
	}
	return 0, fmt.Errorf("Unsupported cipher suite")
}

func (decl DefaultEnvConfigLoader) getRequiredEnv(key string) (string, error) {
	value := decl.envProvider.GetEnv(key, "")
	if value == "" {
		return "", fmt.Errorf(RequiredEnvTemplatedError, key)
	}
	return value, nil
}

func (decl DefaultEnvConfigLoader) getServicesEnv(key string, allowTypes ...string) (map[string][]string, error) {
	value := decl.envProvider.GetEnv(key, "")
	if value == "" {
		return nil, nil
	}
	result := make(map[string][]string)
	services := strings.Split(value, ",")
	for _, service := range services {
		parts := strings.Split(service, " ")
		if len(parts) != 2 {
			return nil,
				fmt.Errorf("%s environment variable must contain word pairs separated by commas and each pair contains exactly two words separated by a single space", key)
		}
		serviceType := parts[0]
		allowedType, allowed := isContained(serviceType, allowTypes)
		if !allowed {
			return nil, fmt.Errorf("environment variable %s must be in the list - [%s]", key, allowTypes)
		}
		serviceName := parts[1]
		result[allowedType] = append(result[allowedType], serviceName)
	}
	return result, nil
}

func (decl DefaultEnvConfigLoader) GetAdditionalHealthStatusConfig() (AdditionalHealthStatusConfig, error) {
	endpoint := decl.envProvider.GetEnv("ADDITIONAL_HEALTH_ENDPOINT", "")
	fullHealthEnabledString := decl.envProvider.GetEnv("EXTERNAL_FULL_HEALTH_ENABLED", "false")
	fullHealthEnabled := false
	if strings.ToLower(fullHealthEnabledString) == "true" {
		fullHealthEnabled = true
	}
	return AdditionalHealthStatusConfig{
		Endpoint:          endpoint,
		FullHealthEnabled: fullHealthEnabled,
	}, nil
}

func isContained(serviceType string, allowTypes []string) (string, bool) {
	for _, allowType := range allowTypes {
		if strings.ToLower(serviceType) == allowType {
			return allowType, true
		}
	}
	return "", false
}
