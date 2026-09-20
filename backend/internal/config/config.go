package config

import (
	"os"
	"strings"
)

const (
	defaultPort    = "7530"
	defaultDataDir = "/data"
	defaultDBPath  = "/data/database/navi-dock.db"
	defaultWebDir  = "/app/web"
)

type Config struct {
	Port           string
	DataDir        string
	DBPath         string
	WebDir         string
	AuthSecret     string
	LANProbeDomain string
}

func Load() Config {
	return Config{
		Port:           env("SERVER_PORT", defaultPort),
		DataDir:        defaultDataDir,
		DBPath:         defaultDBPath,
		WebDir:         defaultWebDir,
		AuthSecret:     env("AUTH_SECRET", ""),
		LANProbeDomain: normalizeProbeDomain(env("LAN_PROBE_DOMAIN", "")),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// normalizeProbeDomain keeps host only (e.g. lan.zeven.site).
func normalizeProbeDomain(value string) string {
	input := strings.TrimSpace(value)
	if input == "" {
		return ""
	}
	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimPrefix(input, "http://")
	input = strings.TrimPrefix(input, "//")
	if i := strings.IndexByte(input, '/'); i >= 0 {
		input = input[:i]
	}
	return strings.Trim(input, ".")
}
