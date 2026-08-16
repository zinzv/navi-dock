package version

import "os"

// Version is injected at build time via -ldflags, e.g.
// -X github.com/navi-dock/navi-dock/internal/version.Version=0.1.4
// Runtime override: APP_VERSION.
var Version = "dev"

func Current() string {
	if v := os.Getenv("APP_VERSION"); v != "" {
		return v
	}
	if Version != "" {
		return Version
	}
	return "dev"
}
