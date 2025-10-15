package version

var (
	version = "0.1.0"

	// For final releases, we set this to an empty string.
	versionPrerelease = "dev"

	// Version of the Coolify CLI.
	Version = func() string {
		if versionPrerelease != "" {
			return version + "-" + versionPrerelease
		}
		return version
	}()
)
