package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	ClusterName              string
	FluxNamespace            string
	CollectControllerLogs    bool
	NotificationSlackWebhook string
	NotificationWebhookURL   string
	GitHubOwner              string
	GitHubRepo               string
	GitHubToken              string
	RequeueInterval          time.Duration
	LogLevel                 string
}

// Load reads environment variables into a Config instance.
func Load() (Config, error) {
	cfg := Config{
		ClusterName:              getenv("FLUXBRAIN_CLUSTER", ""),
		FluxNamespace:            getenv("FLUXBRAIN_FLUX_NAMESPACE", "flux-system"),
		CollectControllerLogs:    getenvBool("FLUXBRAIN_COLLECT_LOGS", false),
		NotificationSlackWebhook: getenv("FLUXBRAIN_SLACK_WEBHOOK", ""),
		NotificationWebhookURL:   getenv("FLUXBRAIN_WEBHOOK_URL", ""),
		GitHubOwner:              getenv("FLUXBRAIN_GITHUB_OWNER", ""),
		GitHubRepo:               getenv("FLUXBRAIN_GITHUB_REPO", ""),
		GitHubToken:              getenv("FLUXBRAIN_GITHUB_TOKEN", ""),
		RequeueInterval:          getenvDuration("FLUXBRAIN_REQUEUE_INTERVAL", 5*time.Minute),
		LogLevel:                 getenv("FLUXBRAIN_LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	if c.ClusterName == "" {
		return errors.New("FLUXBRAIN_CLUSTER is required")
	}
	// Note: Errorbrain-Integration ist optional bis Library verfügbar ist
	return nil
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getenvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.ParseBool(v)
		if err == nil {
			return parsed
		}
	}
	return def
}

func getenvDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		parsed, err := time.ParseDuration(v)
		if err == nil {
			return parsed
		}
		fmt.Fprintf(os.Stderr, "invalid duration for %s: %v\n", key, err)
	}
	return def
}
