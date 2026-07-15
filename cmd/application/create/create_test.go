package create

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCreateCommands_ExposeNewSettingsAndTagsFlags(t *testing.T) {
	commands := []*cobra.Command{
		NewPublicCommand(), NewGitHubCommand(), NewDeployKeyCommand(),
		NewDockerfileCommand(), NewDockerImageCommand(),
	}
	for _, command := range commands {
		for _, flag := range []string{
			"disable-build-cache", "docker-images-to-keep", "include-source-commit-in-build",
			"inject-build-args-to-dockerfile", "is-env-sorting-enabled", "is-git-lfs-enabled",
			"is-git-shallow-clone-enabled", "is-git-submodules-enabled", "is-gzip-enabled",
			"is-pr-deployments-public-enabled", "is-preview-deployments-enabled",
			"is-raw-compose-deployment-enabled", "is-stripprefix-enabled", "stop-grace-period",
			"use-build-secrets", "tag", "tags",
		} {
			assert.NotNil(t, command.Flags().Lookup(flag), "%s missing --%s", command.Name(), flag)
		}
	}
}
