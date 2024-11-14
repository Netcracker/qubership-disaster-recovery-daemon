package config

func NewConfig(configLoader ConfigLoader) (*Config, error) {
	crCfg, err := configLoader.GetCustomResourceConfig()
	if err != nil {
		return nil, err
	}

	healthConfig, err := configLoader.GetHealthConfig()
	if err != nil {
		return nil, err
	}

	drp, err := configLoader.GetDisasterRecoveryPaths()
	if err != nil {
		return nil, err
	}

	healthConfig.DisasterRecoveryStatusPath = drp.StatusPath

	auth, err := configLoader.GetAuthConfig()
	if err != nil {
		return nil, err
	}

	serverConfig, err := configLoader.GetServerConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		*crCfg,
		*healthConfig,
		*drp,
		*auth,
		*serverConfig,
	}
	return cfg, nil
}
