package config

import "git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/api/entity"

const (
	RequiredEnvTemplatedError = "the environment variable '%s' must not be empty"
)

type (
	Config struct {
		CustomResourceConfig
		HealthConfig
		DisasterRecoveryPath
		AuthConfig
		ServerConfig
	}

	CustomResourceConfig struct {
		Name      string
		Namespace string
		Group     string
		Version   string
		Resource  string
	}

	HealthConfig struct {
		ActiveMainServices           map[string][]string
		ActiveAdditionalServices     map[string][]string
		StandbyMainServices          map[string][]string
		StandbyAdditionalServices    map[string][]string
		DisableMainServices          map[string][]string
		DisableAdditionalServices    map[string][]string
		AdditionalHealthStatusConfig AdditionalHealthStatusConfig
		DisasterRecoveryStatusPath   DisasterRecoveryStatusPath
	}

	DisasterRecoveryPath struct {
		StatusPath     DisasterRecoveryStatusPath
		ModePath       []string
		NoWaitPath     []string
		NoWaitAsString bool
	}

	DisasterRecoveryStatusPath struct {
		ModePath           []string
		StatusPath         []string
		CommentPath        []string
		TreatStatusAsField bool
	}

	AuthConfig struct {
		AuthEnabled                   bool
		SiteManagerServiceAccountName string
		SiteManagerNamespace          string
		SiteManagerCustomAudience     string
	}

	AdditionalHealthStatusConfig struct {
		Endpoint          string
		HealthFunc        func(request entity.HealthRequest) (entity.HealthResponse, error)
		FullHealthEnabled bool
	}

	ServerConfig struct {
		Port       int
		Suites     []uint16
		TLSEnabled bool
		CertsPath  string
	}
)

type ConfigLoader interface {
	GetCustomResourceConfig() (*CustomResourceConfig, error)
	GetDisasterRecoveryPaths() (*DisasterRecoveryPath, error)
	GetHealthConfig() (*HealthConfig, error)
	GetAuthConfig() (*AuthConfig, error)
	GetServerConfig() (*ServerConfig, error)
}

type EnvConfigLoader interface {
	ConfigLoader
	getRequiredEnv(string) (string, error)
	getServicesEnv(string, ...string) (map[string][]string, error)
}
