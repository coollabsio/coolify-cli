package update

import (
	"fmt"
	"os"
	"runtime"

	coolifyRuntime "github.com/coollabsio/coolify-cli/cmd/runtime"
	selfupdate "github.com/creativeprojects/go-selfupdate"
	compareVersion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

type cliUpdate struct {
	coolify coolifyRuntime.Getter
}

func New(c coolifyRuntime.Getter) *cliUpdate {
	return &cliUpdate{
		coolify: c,
	}
}

func (c *cliUpdate) NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update Coolify CLI",
		Long: `
Update the Coolify CLI to the latest version from GitHub releases.
`,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the latest release from GitHub
			latest, found, err := selfupdate.DetectLatest(
				cmd.Context(),
				selfupdate.ParseSlug("coollabsio/cli-coolify"),
			)
			if err != nil {
				return fmt.Errorf("error detecting version: %v", err)
			}
			if !found {
				return fmt.Errorf("latest version for %s/%s not found", runtime.GOOS, runtime.GOARCH)
			}

			// Compare versions
			currentVersion, err := compareVersion.NewVersion(c.coolify().Version)
			if err != nil {
				return fmt.Errorf("could not parse current version: %v", err)
			}

			latestVersion, err := compareVersion.NewVersion(latest.Version())
			if err != nil {
				return fmt.Errorf("could not parse latest version: %v", err)
			}

			// Update if needed
			if currentVersion.LessThan(latestVersion) {
				exe, err := os.Executable()
				if err != nil {
					return fmt.Errorf("could not locate executable path: %v", err)
				}

				cmd.Println("Updating to version", latest.Version())
				if err := selfupdate.UpdateTo(
					cmd.Context(),
					latest.AssetURL,
					latest.AssetName,
					exe,
				); err != nil {
					return fmt.Errorf("error updating binary: %v", err)
				}

				cmd.Println("Successfully updated to version", latest.Version())
			} else {
				cmd.Println("You are already on the latest version:", c.coolify().Version)
			}

			return nil
		},
	}

	return cmd
}
