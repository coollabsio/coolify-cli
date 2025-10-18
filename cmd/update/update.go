package update

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	selfupdate "github.com/creativeprojects/go-selfupdate"
	compareVersion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/version"
)

func NewUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update Coolify CLI",
		RunE: func(_ *cobra.Command, _ []string) error {
			latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("coollabsio/coolify-cli"))
			if err != nil {
				return fmt.Errorf("failed to detect latest version: %w", err)
			}
			if !found {
				return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
			}
			currentVersion, err := compareVersion.NewVersion(version.CliVersion)
			if err != nil {
				return fmt.Errorf("failed to parse current version: %w", err)
			}

			latestVersion, err := compareVersion.NewVersion(latest.Version())
			if err != nil {
				return fmt.Errorf("failed to parse latest version: %w", err)
			}
			if currentVersion.LessThan(latestVersion) {
				exe, err := os.Executable()
				if err != nil {
					return fmt.Errorf("could not locate executable path: %w", err)
				}
				if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
					return fmt.Errorf("error occurred while updating binary: %w", err)
				}
				log.Printf("Successfully updated to version %s", latest.Version())
			} else {
				log.Printf("No new update available. You are already running the latest version: %s", currentVersion.String())
			}
			return nil

		},
	}
}
