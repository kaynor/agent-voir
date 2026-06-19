package gateway

import (
	"os"
	"strconv"
)

// Config holds gateway runtime settings from environment variables.
type Config struct {
	Addr            string
	APIKey          string
	CacheMode       string
	CacheTTLSeconds int64
	RedisAddr       string
	OpenAIAPIKey    string
	OpenAIBaseURL   string
	RegistryAPIURL     string
	RegistryAPIKey     string
	TokenAccountingURL string
	OPAURL             string
}

func LoadConfig() Config {
	return Config{
		Addr:            env("GATEWAY_ADDR", ":8080"),
		APIKey:          env("GATEWAY_API_KEY", "agentvoir-local-dev-key"),
		CacheMode:       env("CACHE_MODE", "exact_only"),
		CacheTTLSeconds: envInt64("CACHE_DEFAULT_TTL_SECONDS", 86400),
		RedisAddr:       os.Getenv("REDIS_ADDR"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:   env("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		RegistryAPIURL:     env("REGISTRY_API_URL", "http://localhost:8081"),
		RegistryAPIKey:     os.Getenv("REGISTRY_API_KEY"),
		TokenAccountingURL: env("TOKEN_ACCOUNTING_URL", "http://localhost:8082"),
		OPAURL:             os.Getenv("OPA_URL"),
	}
}

func (c Config) CacheReadEnabled() bool {
	switch c.CacheMode {
	case "off", "write_only":
		return false
	default:
		return true
	}
}

func (c Config) CacheWriteEnabled() bool {
	return c.CacheMode != "off"
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
