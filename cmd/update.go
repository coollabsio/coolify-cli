package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	selfupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Coolify CLI",
	Run: func(cmd *cobra.Command, args []string) {
		latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("coollabsio/coolify-cli"))
		if err != nil {
			log.Printf("error occurred while detecting version: %v", err)
			return
		}
		log.Printf("found latest version: %t", found)
		if !found {
			log.Printf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
			return
		}

		if latest.LessOrEqual(Version) {
			log.Printf("Current version (%s) is the latest", Version)
			return
		}

		exe, err := os.Executable()
		if err != nil {
			log.Printf("could not locate executable path: %v", err)
			return
		}
		log.Printf("Current version: %s, Latest version: %s", Version, latest.Version())
		log.Printf("asset url: %s, asset name: %s", latest.AssetURL, latest.AssetName)
		if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
			fmt.Printf("error occurred while updating binary: %v", err)
			return
		}
		log.Printf("Successfully updated to version %s", latest.Version())
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
