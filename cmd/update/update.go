package update

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/coollabsio/coolify-cli/internal/version"
	selfupdate "github.com/creativeprojects/go-selfupdate"
	compareVersion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update Coolify CLI",
		Run: func(cmd *cobra.Command, args []string) {
			latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("coollabsio/coolify-cli"))
			if err != nil {
				log.Printf("Error occurred while detecting version: %v", err)
				return
			}
			if !found {
				log.Printf("Latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
				return
			}
			currentVersion, err := compareVersion.NewVersion(version.CliVersion)
			if err != nil {
				log.Printf("Could not parse current version: %v", err)
				return
			}

			latestVersion, err := compareVersion.NewVersion(latest.Version())
			if err != nil {
				log.Printf("Could not parse latest version: %v", err)
				return
			}
			if currentVersion.LessThan(latestVersion) {
				exe, err := os.Executable()
				if err != nil {
					log.Printf("Could not locate executable path: %v", err)
					return
				}
				if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
					fmt.Printf("Error occurred while updating binary: %v", err)
					return
				}
				log.Printf("Successfully updated to version %s", latest.Version())
			} else {
				log.Printf("No new update available. You are already running the latest version: %s", currentVersion.String())
			}

		},
	}
}
