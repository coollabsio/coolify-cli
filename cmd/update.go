package cmd

import (
	"context"
	"fmt"
	"log"
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
			fmt.Printf("error occurred while detecting version: %w", err)
			return
		}
		if !found {
			fmt.Printf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
			return
		}

		if latest.LessOrEqual(Version) {
			log.Printf("Current version (%s) is the latest", Version)
			return
		}

		// exe, err := os.Executable()
		// if err != nil {
		// 	errors.New("could not locate executable path")
		// 	return
		// }
		// if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		// 	fmt.Printf("error occurred while updating binary: %w", err)
		// 	return
		// }
		log.Printf("Successfully updated to version %s", latest.Version())
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
