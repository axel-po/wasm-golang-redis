package config

import (
	"os"
	"time"
)

type Config struct {
	DataDir          string
	FlushInterval    time.Duration
	SnapshotInterval time.Duration
}

func Load() Config {
	return Config{
		DataDir:          getEnv("WASMREDIS_DATA_DIR", "data"),
		FlushInterval:    getEnvDuration("WASMREDIS_FLUSH_INTERVAL", time.Second),
		SnapshotInterval: getEnvDuration("WASMREDIS_SNAPSHOT_INTERVAL", 2*time.Minute),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
