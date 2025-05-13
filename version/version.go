package version

// Version information set at build time
var (
	// Version is the current version of codacy-cli
	Version = "development"

	// GitCommit is the git commit hash of the build
	GitCommit = "unknown"

	// BuildTime is the time the binary was built
	BuildTime = "unknown"
)

// GetVersion returns the current version of codacy-cli
func GetVersion() string {
	if GitCommit != "unknown" {
		return Version + " (" + GitCommit + ") built at " + BuildTime
	}
	return Version
}
