package appflags

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/models"
)

// BindSettingsFlags registers application settings supported by both create and update endpoints.
func BindSettingsFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.Bool("disable-build-cache", false, "Disable the build cache")
	flags.Int("docker-images-to-keep", 0, "Number of Docker images to retain")
	flags.Bool("include-source-commit-in-build", false, "Include the source commit in the build")
	flags.Bool("inject-build-args-to-dockerfile", false, "Inject build arguments into the Dockerfile")
	flags.Bool("is-env-sorting-enabled", false, "Sort environment variables")
	flags.Bool("is-git-lfs-enabled", false, "Enable Git LFS")
	flags.Bool("is-git-shallow-clone-enabled", false, "Use a shallow Git clone")
	flags.Bool("is-git-submodules-enabled", false, "Clone Git submodules")
	flags.Bool("is-gzip-enabled", false, "Enable gzip compression")
	flags.Bool("is-pr-deployments-public-enabled", false, "Make pull request deployments public")
	flags.Bool("is-preview-deployments-enabled", false, "Enable preview deployments")
	flags.Bool("is-raw-compose-deployment-enabled", false, "Deploy the raw Docker Compose definition")
	flags.Bool("is-stripprefix-enabled", false, "Enable path prefix stripping")
	flags.Int("stop-grace-period", 0, "Container stop grace period in seconds")
	flags.Bool("use-build-secrets", false, "Use Docker Build Secrets for build-time variables")
	// Advanced settings exposed via API parity
	flags.Bool("is-log-drain-enabled", false, "Enable application log drain")
	flags.Bool("is-gpu-enabled", false, "Enable GPU")
	flags.String("gpu-driver", "", "GPU driver")
	flags.String("gpu-count", "", "GPU count")
	flags.String("gpu-device-ids", "", "GPU device IDs")
	flags.String("gpu-options", "", "GPU options")
	flags.Bool("is-consistent-container-name-enabled", false, "Use consistent container names")
	flags.String("custom-internal-name", "", "Custom internal hostname")
	flags.String("preview-url-template", "", "Preview URL template")
	flags.Int("max-restart-count", 0, "Maximum container restart count")
}

// ApplySettingsFlags copies explicitly changed flags to an API request.
func ApplySettingsFlags(cmd *cobra.Command, settings *models.ApplicationSettingsRequest) bool {
	changed := false
	setBool := func(flag string, target **bool) {
		if cmd.Flags().Changed(flag) {
			value, _ := cmd.Flags().GetBool(flag)
			*target = &value
			changed = true
		}
	}
	setInt := func(flag string, target **int) {
		if cmd.Flags().Changed(flag) {
			value, _ := cmd.Flags().GetInt(flag)
			*target = &value
			changed = true
		}
	}

	setBool("disable-build-cache", &settings.DisableBuildCache)
	setInt("docker-images-to-keep", &settings.DockerImagesToKeep)
	setBool("include-source-commit-in-build", &settings.IncludeSourceCommitInBuild)
	setBool("inject-build-args-to-dockerfile", &settings.InjectBuildArgsToDockerfile)
	setBool("is-env-sorting-enabled", &settings.IsEnvSortingEnabled)
	setBool("is-git-lfs-enabled", &settings.IsGitLFSEnabled)
	setBool("is-git-shallow-clone-enabled", &settings.IsGitShallowCloneEnabled)
	setBool("is-git-submodules-enabled", &settings.IsGitSubmodulesEnabled)
	setBool("is-gzip-enabled", &settings.IsGzipEnabled)
	setBool("is-pr-deployments-public-enabled", &settings.IsPRDeploymentsPublicEnabled)
	setBool("is-preview-deployments-enabled", &settings.IsPreviewDeploymentsEnabled)
	setBool("is-raw-compose-deployment-enabled", &settings.IsRawComposeDeploymentEnabled)
	setBool("is-stripprefix-enabled", &settings.IsStripPrefixEnabled)
	setInt("stop-grace-period", &settings.StopGracePeriod)
	setBool("use-build-secrets", &settings.UseBuildSecrets)
	setBool("is-log-drain-enabled", &settings.IsLogDrainEnabled)
	setBool("is-gpu-enabled", &settings.IsGPUEnabled)
	setBool("is-consistent-container-name-enabled", &settings.IsConsistentContainerNameEnabled)

	setString := func(flag string, target **string) {
		if cmd.Flags().Changed(flag) {
			value, _ := cmd.Flags().GetString(flag)
			*target = &value
			changed = true
		}
	}
	setString("gpu-driver", &settings.GPUDriver)
	setString("gpu-count", &settings.GPUCount)
	setString("gpu-device-ids", &settings.GPUDeviceIDs)
	setString("gpu-options", &settings.GPUOptions)
	setString("custom-internal-name", &settings.CustomInternalName)

	return changed
}

// ApplyAdvancedAppFlags applies top-level application fields not nested under settings.
func ApplyAdvancedAppFlags(cmd *cobra.Command, req *models.ApplicationUpdateRequest) bool {
	changed := false
	if cmd.Flags().Changed("preview-url-template") {
		v, _ := cmd.Flags().GetString("preview-url-template")
		req.PreviewURLTemplate = &v
		changed = true
	}
	if cmd.Flags().Changed("max-restart-count") {
		v, _ := cmd.Flags().GetInt("max-restart-count")
		req.MaxRestartCount = &v
		changed = true
	}
	return changed
}

// BindTagsFlag registers the tags accepted by application creation endpoints.
func BindTagsFlag(cmd *cobra.Command) {
	cmd.Flags().StringArray("tag", nil, "Tag to assign to the application (repeatable)")
	cmd.Flags().StringSlice("tags", nil, "Tags to assign to the application")
}

// ApplyTagsFlag copies explicitly supplied tags to a create request.
func ApplyTagsFlag(cmd *cobra.Command, tags *[]string) bool {
	if !cmd.Flags().Changed("tag") && !cmd.Flags().Changed("tags") {
		return false
	}
	repeated, _ := cmd.Flags().GetStringArray("tag")
	commaSeparated, _ := cmd.Flags().GetStringSlice("tags")
	*tags = append(*tags, repeated...)
	*tags = append(*tags, commaSeparated...)
	return true
}
