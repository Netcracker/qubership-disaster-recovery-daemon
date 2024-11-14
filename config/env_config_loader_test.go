package config

import "testing"

func NewTestEnvProvider(envs map[string]string) TestEnvProvider {
	return TestEnvProvider{envs: envs}
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

func TestAuthEnabledIsFalse(t *testing.T) {
	envs := make(map[string]string)
	cfgLoader := NewEnvConfigLoader(NewTestEnvProvider(envs))
	authConfig, err := cfgLoader.GetAuthConfig()
	if err != nil || authConfig.AuthEnabled {
		t.Fatalf("authentication must be enabled")
	}
}
