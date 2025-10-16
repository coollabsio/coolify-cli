package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/coollabsio/coolify-cli/cmd/application"
	"github.com/coollabsio/coolify-cli/cmd/completion"
	"github.com/coollabsio/coolify-cli/cmd/context"
	"github.com/coollabsio/coolify-cli/cmd/database"
	"github.com/coollabsio/coolify-cli/cmd/deployment"
	"github.com/coollabsio/coolify-cli/cmd/github"
	"github.com/coollabsio/coolify-cli/cmd/privatekeys"
	"github.com/coollabsio/coolify-cli/cmd/project"
	"github.com/coollabsio/coolify-cli/cmd/resources"
	"github.com/coollabsio/coolify-cli/cmd/server"
	"github.com/coollabsio/coolify-cli/cmd/service"
	"github.com/coollabsio/coolify-cli/cmd/teams"
	"github.com/coollabsio/coolify-cli/cmd/update"
	cliversion "github.com/coollabsio/coolify-cli/cmd/version"
	"github.com/coollabsio/coolify-cli/internal/config"
	"github.com/coollabsio/coolify-cli/internal/version"
	compareVersion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Legacy global variables - kept for backward compatibility during migration
// TODO: Remove these once all commands are refactored
var (
	Version            string
	Name               string
	Fqdn               string
	Token              string
	ContextName        string
	Debug              bool
	ShowSensitive      bool
	Format             string
	JsonMode           bool
	PrettyMode         bool
	SetDefaultInstance bool
	w                  = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
)

var rootCmd = &cobra.Command{
	Use:           "coolify",
	Short:         "Coolify CLI",
	Long:          `A CLI tool to interact with Coolify API.`,
	SilenceUsage:  true,  // Don't show usage on errors
	SilenceErrors: false, // Still print errors
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(0)
	}
}

func init() {
	rootCmd = &cobra.Command{
		Use:           "coolify",
		Short:         "Coolify CLI",
		Long:          fmt.Sprintf("A CLI tool to interact with Coolify API.\nVersion: %s", version.CliVersion),
		SilenceUsage:  true,  // Don't show usage on errors
		SilenceErrors: false, // Still print errors
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&Token, "token", "", "", "Token for authentication (override context token)")
	rootCmd.PersistentFlags().StringVarP(&ContextName, "context", "", "", "Use specific context by name")

	rootCmd.PersistentFlags().StringVarP(&Format, "format", "", "table", "Format output (table|json|pretty)")
	rootCmd.PersistentFlags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "", false, "Debug mode")

	// Register all subcommands
	rootCmd.AddCommand(application.NewAppCommand())
	rootCmd.AddCommand(context.NewContextCommand())
	rootCmd.AddCommand(completion.NewCompletionsCommand())
	rootCmd.AddCommand(database.NewDatabaseCommand())
	rootCmd.AddCommand(deployment.NewDeploymentCommand())
	rootCmd.AddCommand(github.NewGitHubCommand())
	rootCmd.AddCommand(privatekeys.NewPrivateKeysCommand())
	rootCmd.AddCommand(project.NewProjectCommand())
	rootCmd.AddCommand(resources.NewResourceCommand())
	rootCmd.AddCommand(server.NewServerCommand())
	rootCmd.AddCommand(service.NewServiceCommand())
	rootCmd.AddCommand(teams.NewTeamsCommand())
	rootCmd.AddCommand(update.NewUpdateCommand())
	rootCmd.AddCommand(cliversion.NewVersionCommand())
	rootCmd.AddCommand(NewDocsCommand())
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(config.Path()[:len(config.Path())-len("/config.json")])

	// Ensure config directory exists
	configDir := config.Path()[:len(config.Path())-len("/config.json")]
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found. Creating a new one at", config.Path())
			if err := config.CreateDefault(); err != nil {
				log.Printf("Failed to create default config: %v\n", err)
				return
			}
			// Reload config after creating default
			if err := viper.ReadInConfig(); err != nil {
				log.Printf("Failed to read newly created config: %v\n", err)
				return
			}
		} else {
			fmt.Println("Error reading config file:", err)
			return
		}
	}

	if Debug {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Note: We don't pre-populate Fqdn/Token here anymore
	// They are loaded on-demand by getAPIClient() based on --instance or default instance
	// This allows --instance flag to work correctly

	// Check for updates
	latestVersionStr, err := version.CheckLatestVersionOfCli(Debug)
	if err != nil {
		if Debug {
			log.Println("Failed to check for updates:", err)
		}
	}

	// Compare versions properly using semantic versioning
	if latestVersionStr != "" {
		latestVersion, err := compareVersion.NewVersion(latestVersionStr)
		if err == nil {
			currentVersion, err := compareVersion.NewVersion(version.CliVersion)
			if err == nil && latestVersion.GreaterThan(currentVersion) {
				if Debug {
					log.Printf("New version of Coolify CLI is available: %s\n", latestVersionStr)
				}
			}
		}
	}
}
