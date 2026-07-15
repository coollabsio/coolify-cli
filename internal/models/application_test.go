package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplication_UnmarshalSettings(t *testing.T) {
	var app Application
	err := json.Unmarshal([]byte(`{
		"uuid":"app-uuid",
		"settings":{
			"is_log_drain_enabled":true,
			"is_gpu_enabled":true,
			"gpu_driver":"nvidia",
			"is_include_timestamps":true,
			"is_consistent_container_name_enabled":true,
			"custom_internal_name":"backend",
			"disable_build_cache":true,
			"is_git_shallow_clone_enabled":true,
			"use_build_secrets":true,
			"docker_images_to_keep":7,
			"stop_grace_period":30
		}
	}`), &app)

	require.NoError(t, err)
	require.NotNil(t, app.Settings)
	assert.True(t, *app.Settings.IsLogDrainEnabled)
	assert.True(t, *app.Settings.IsGPUEnabled)
	assert.Equal(t, "nvidia", *app.Settings.GPUDriver)
	assert.True(t, *app.Settings.IsIncludeTimestamps)
	assert.True(t, *app.Settings.IsConsistentContainerNameEnabled)
	assert.Equal(t, "backend", *app.Settings.CustomInternalName)
	assert.True(t, *app.Settings.DisableBuildCache)
	assert.True(t, *app.Settings.IsGitShallowCloneEnabled)
	assert.True(t, *app.Settings.UseBuildSecrets)
	assert.Equal(t, 7, *app.Settings.DockerImagesToKeep)
	assert.Equal(t, 30, *app.Settings.StopGracePeriod)
}

func TestApplicationCreateSettings_MarshalNewAPIFields(t *testing.T) {
	enabled := true
	imagesToKeep := 5
	stopGracePeriod := 60
	req := ApplicationCreatePublicRequest{
		ApplicationSettingsRequest: ApplicationSettingsRequest{
			DisableBuildCache:             &enabled,
			DockerImagesToKeep:            &imagesToKeep,
			IncludeSourceCommitInBuild:    &enabled,
			InjectBuildArgsToDockerfile:   &enabled,
			IsEnvSortingEnabled:           &enabled,
			IsGitLFSEnabled:               &enabled,
			IsGitShallowCloneEnabled:      &enabled,
			IsGitSubmodulesEnabled:        &enabled,
			IsGzipEnabled:                 &enabled,
			IsPRDeploymentsPublicEnabled:  &enabled,
			IsPreviewDeploymentsEnabled:   &enabled,
			IsRawComposeDeploymentEnabled: &enabled,
			IsStripPrefixEnabled:          &enabled,
			StopGracePeriod:               &stopGracePeriod,
			UseBuildSecrets:               &enabled,
		},
		Tags: []string{"production", "web"},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)
	var body map[string]any
	require.NoError(t, json.Unmarshal(data, &body))

	for _, field := range []string{
		"disable_build_cache", "docker_images_to_keep", "include_source_commit_in_build",
		"inject_build_args_to_dockerfile", "is_env_sorting_enabled", "is_git_lfs_enabled",
		"is_git_shallow_clone_enabled", "is_git_submodules_enabled", "is_gzip_enabled",
		"is_pr_deployments_public_enabled", "is_preview_deployments_enabled",
		"is_raw_compose_deployment_enabled", "is_stripprefix_enabled", "stop_grace_period",
		"use_build_secrets", "tags",
	} {
		assert.Contains(t, body, field)
	}
}
