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

func TestDockerComposeDomains_MarshalRequestShape(t *testing.T) {
	domains := []DockerComposeDomain{
		{Name: "litellm", Domain: "https://litellm.example.com"},
		{Name: "admin", Domain: "https://admin.example.com,https://admin2.example.com"},
	}
	tests := []struct {
		name string
		req  any
	}{
		{name: "public create", req: ApplicationCreatePublicRequest{DockerComposeDomains: domains}},
		{name: "github create", req: ApplicationCreateGitHubAppRequest{DockerComposeDomains: domains}},
		{name: "deploy key create", req: ApplicationCreateDeployKeyRequest{DockerComposeDomains: domains}},
		{name: "update", req: ApplicationUpdateRequest{DockerComposeDomains: domains}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			require.NoError(t, err)

			var body map[string]any
			require.NoError(t, json.Unmarshal(data, &body))
			items, ok := body["docker_compose_domains"].([]any)
			require.True(t, ok)
			require.Len(t, items, 2)
			assert.Equal(t, map[string]any{
				"name":   "litellm",
				"domain": "https://litellm.example.com",
			}, items[0])
		})
	}
}

func TestDockerComposeDomains_OmittedWhenUnset(t *testing.T) {
	data, err := json.Marshal(ApplicationUpdateRequest{})
	require.NoError(t, err)
	assert.NotContains(t, string(data), "docker_compose_domains")
}

func TestEnvironmentVariableRequests_MarshalBuildtimeField(t *testing.T) {
	buildTime := false
	runtime := true
	literal := true
	key := "EXAMPLE_FLAG"
	value := "disabled"

	tests := []struct {
		name string
		req  any
	}{
		{
			name: "create request",
			req: EnvironmentVariableCreateRequest{
				Key:         key,
				Value:       value,
				IsBuildTime: &buildTime,
				IsRuntime:   &runtime,
				IsLiteral:   &literal,
			},
		},
		{
			name: "update request",
			req: EnvironmentVariableUpdateRequest{
				Key:         &key,
				Value:       &value,
				IsBuildTime: &buildTime,
				IsRuntime:   &runtime,
				IsLiteral:   &literal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			require.NoError(t, err)

			var body map[string]any
			require.NoError(t, json.Unmarshal(data, &body))

			// The application environment API expects is_buildtime as the
			// build-time field name; is_build_time is rejected as unknown.
			assert.Equal(t, false, body["is_buildtime"])
			assert.NotContains(t, body, "is_build_time")

			// Explicitly supplied boolean flags must be preserved.
			assert.Equal(t, true, body["is_runtime"])
			assert.Equal(t, true, body["is_literal"])
		})
	}
}

func TestEnvironmentVariableCreateRequest_OmitsUnsetBuildtimeField(t *testing.T) {
	// An unset pointer means --build-time was not provided. The field should be
	// omitted rather than sent as either spelling of the API field.
	req := EnvironmentVariableCreateRequest{
		Key:   "EXAMPLE_FLAG",
		Value: "disabled",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	assert.NotContains(t, string(data), `"is_buildtime"`)
	assert.NotContains(t, string(data), `"is_build_time"`)
}
