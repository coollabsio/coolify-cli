package update

import (
	"fmt"
	"os"
	"runtime"

	coolifyRuntime "github.com/coollabsio/cli-coolify/cmd/runtime"
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

				cmd.Printf("Updating to version v%s\n", latest.Version())

				// The selfupdate library needs to know the name of the binary inside the tar.gz
				// This is set to "coolify" in our goreleaser configuration
				if err := selfupdate.UpdateTo(
					cmd.Context(),
					latest.AssetURL,
					"coolify", // The actual binary name inside the archive
					exe,
				); err != nil {
					return fmt.Errorf("error updating binary: %v", err)
				}

				cmd.Printf("Successfully updated to version v%s\n", latest.Version())
			} else {
				cmd.Printf("You are already on the latest version: %s\n", c.coolify().GetFormattedVersion())
			}

			return nil
		},
	}

	return cmd
}
