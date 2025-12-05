package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/coollabsio/coolify-cli/cmd/application"
	"github.com/coollabsio/coolify-cli/cmd/completion"
	configcmd "github.com/coollabsio/coolify-cli/cmd/config"
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
	JSONMode           bool
	PrettyMode         bool
	SetDefaultInstance bool
)

var rootCmd = &cobra.Command{
	Use:           "coolify",
	Short:         "Coolify CLI",
	Long:          `A CLI tool to interact with Coolify API.`,
	SilenceUsage:  true,  // Don't show usage on errors
	SilenceErrors: false, // Still print errors
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		return nil
	},
}

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd = &cobra.Command{
		Use:           "coolify",
		Short:         "Coolify CLI",
		Long:          fmt.Sprintf("A CLI tool to interact with Coolify API.\nVersion: %s", version.GetVersion()),
		SilenceUsage:  true,  // Don't show usage on errors
		SilenceErrors: false, // Still print errors
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
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
	rootCmd.AddCommand(completion.NewCompletionsCommand())
	rootCmd.AddCommand(configcmd.NewConfigCommand())
	rootCmd.AddCommand(context.NewContextCommand())
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
		if err := os.MkdirAll(configDir, 0750); err != nil {
			log.Printf("Failed to create config directory: %v\n", err)
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &notFoundErr) {
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

	// Check for updates (errors are handled silently inside the function)
	_, _ = version.CheckLatestVersionOfCli(Debug)
}
