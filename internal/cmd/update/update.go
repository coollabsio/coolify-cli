package update

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/coollabsio/cli-coolify/internal/config"
	"github.com/coollabsio/cli-coolify/internal/updater"
	"github.com/spf13/cobra"
)

type cliUpdate struct {
	coolify config.Getter
}

func New(c config.Getter) *cliUpdate {
	return &cliUpdate{
		coolify: c,
	}
}

func (c *cliUpdate) NewCommand() *cobra.Command {
	var preRelease bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update Coolify CLI",
		Long: `
Update the Coolify CLI to the latest version from GitHub releases.

By default, the command will update to the latest stable version.
Use the --pre-release flag to update to the latest pre-release version.
`,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// we should check if the current version is a pre-release
			currentVersion := c.coolify().Version
			isPreRelease := strings.Contains(currentVersion, "-")
			// Create our custom updater
			update := updater.New("coollabsio", "cli-coolify", c.coolify().Version)

			// Check for updates
			c.coolify().Logger.Infof("Checking for updates...")

			// Check if an update is available without performing the update
			release, hasUpdate, err := update.Check(cmd.Context(), preRelease)
			if err != nil {
				return fmt.Errorf("error checking for updates: %v", err)
			}

			if isPreRelease && !preRelease && !hasUpdate {
				c.coolify().Logger.Warnf("You are on a pre-release version of the CLI. Use the --pre-release flag to update to the latest pre-release version.")
				return nil
			}

			if !hasUpdate {
				c.coolify().Logger.Infof("You are already on the latest version: %s\n", c.coolify().GetFormattedVersion())
				return nil
			}

			c.coolify().Logger.Infof("Found new version: v%s (current: %s)\n", release.Version, c.coolify().GetFormattedVersion())

			// Format OS/Arch for display
			platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
			c.coolify().Logger.Infof("Downloading update for %s...", platform)

			// Perform the update
			newVersion, err := update.To(cmd.Context(), release)
			if err != nil {
				return fmt.Errorf("update failed: %v", err)
			}

			c.coolify().Logger.Infof("Successfully updated to version v%s\n", newVersion)

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&preRelease, "pre-release", false, "Update to pre-release version")

	return cmd
}
