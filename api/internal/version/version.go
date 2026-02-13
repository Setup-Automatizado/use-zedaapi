// Package version provides version information for the WhatsApp API.
// Version is typically set at build time via ldflags:
//
//	go build -ldflags "-X go.mau.fi/whatsmeow/api/internal/version.version=$(cat ../VERSION)"
package version

import (
	"os"
	"path/filepath"
	"strings"
)

// These variables are set at build time via ldflags.
// If not set, they will be populated from the VERSION file or use defaults.
var (
	// version is the semantic version (e.g., "2.0.0-develop.1")
	version = ""
	// buildTime is the UTC build timestamp
	buildTime = ""
	// gitCommit is the short git commit hash
	gitCommit = ""
)

// Info contains version information.
type Info struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time,omitempty"`
	GitCommit string `json:"git_commit,omitempty"`
}

// Get returns the version information.
// It tries to read from the VERSION file if the version was not set at build time.
func Get() Info {
	v := version
	if v == "" {
		v = readVersionFile()
	}
	if v == "" {
		v = "unknown"
	}

	return Info{
		Version:   strings.TrimSpace(v),
		BuildTime: buildTime,
		GitCommit: gitCommit,
	}
}

// String returns the version string.
func String() string {
	return Get().Version
}

// readVersionFile attempts to read the VERSION file from known locations.
func readVersionFile() string {
	// Try relative paths from common execution locations
	paths := []string{
		"VERSION",
		"../VERSION",
		"../../VERSION",
	}

	// Also try from executable directory
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		paths = append(paths,
			filepath.Join(execDir, "VERSION"),
			filepath.Join(execDir, "..", "VERSION"),
			filepath.Join(execDir, "..", "..", "VERSION"),
		)
	}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	return ""
}
