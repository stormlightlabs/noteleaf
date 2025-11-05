package version

import "fmt"

var (
	// Version is the semantic version (e.g., "1.0.0", "1.0.0-rc1")
	Version string = "dev"
	// Commit is the git commit hash
	Commit string = "none"
	// BuildDate is the build timestamp
	BuildDate string = "unknown"
)

// String returns a formatted version string
func String() string {
	if Version == "dev" && Commit != "none" {
		return fmt.Sprintf("%s+%s", Version, Commit)
	}
	return Version
}

// UserAgent returns a formatted user agent string for HTTP requests
func UserAgent(appName, contactEmail string) string {
	return fmt.Sprintf("%s/%s (%s)", appName, Version, contactEmail)
}
