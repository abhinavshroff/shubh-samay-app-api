package config

import "os"

type Config struct {
	DatabaseURL string
	AdminToken  string
	EphePath    string
	FCMKey      string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://postgres:fIbAoiNykONGhzWfEzgOihhSfhRTduQq@postgres.railway.internal:5432/railway"),
		AdminToken:  getEnv("ADMIN_TOKEN", "change-me-in-production"),
		EphePath:    getEnv("EPHE_PATH", "/usr/share/libswe/ephe"),
		FCMKey:      getEnv("FCM_SERVER_KEY", ""),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
