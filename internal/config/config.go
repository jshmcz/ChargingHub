package config

import "os"

type Config struct {
	Port       string
	DevicePort string
	DBPath     string
	SecretKey  string
	StaticDir  string
	MobileDir  string
}

func Load() Config {
	return Config{
		Port:       env("PORT", "8080"),
		DevicePort: env("DEVICE_PORT", "8768"),
		DBPath:     env("DB_PATH", "data/charginghub.db"),
		SecretKey:  env("SECRET_KEY", "dev-secret-change-me"),
		StaticDir:  env("STATIC_DIR", "web/admin/dist"),
		MobileDir:  env("MOBILE_DIR", "web/mobile/dist"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
