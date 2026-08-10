package config

import "os"

const (
	defaultPort    = "7530"
	defaultDataDir = "/data"
	defaultDBPath  = "/data/database/navi-dock.db"
	defaultWebDir  = "/app/web"
)

type Config struct {
	Port    string
	DataDir string
	DBPath  string
	WebDir  string
}

func Load() Config {
	return Config{
		Port:    env("SERVER_PORT", defaultPort),
		DataDir: defaultDataDir,
		DBPath:  defaultDBPath,
		WebDir:  defaultWebDir,
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
