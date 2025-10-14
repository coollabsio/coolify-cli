package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/config"
	compareVersion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CliVersion is the CLI version
var CliVersion = "1.0.0"

// CheckInterval for version checking
var CheckInterval = 10 * time.Minute

// SensitiveInformationOverlay is the string used to hide sensitive data
var SensitiveInformationOverlay = "********"

// Legacy global variables - kept for backward compatibility during migration
// TODO: Remove these once all commands are refactored
var (
	Version string
	Name    string
	Fqdn    string
	Token   string
	InstanceName string
	Debug   bool
	ShowSensitive bool
	Force   bool
	Format  string
	JsonMode bool
	PrettyMode bool
	SetDefaultInstance bool
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	Instance http.Client
)

// Tag represents a git tag for version checking
type Tag struct {
	Ref string `json:"ref"`
}

var rootCmd = &cobra.Command{
	Use:   "coolify",
	Short: "Coolify CLI",
	Long:  `A CLI tool to interact with Coolify API.`,
	SilenceUsage: true, // Don't show usage on errors
	SilenceErrors: false, // Still print errors
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// getAPIClient creates an API client from command flags or config
func getAPIClient(cmd *cobra.Command) (*api.Client, error) {
	// Try to get from flags first (check both local and persistent flags)
	fqdn, _ := cmd.Flags().GetString("host")
	token, _ := cmd.Flags().GetString("token")
	instanceName, _ := cmd.Flags().GetString("instance")
	debug, _ := cmd.Flags().GetBool("debug")

	// If not from flags, load from config
	if fqdn == "" || token == "" {
		cfg, err := config.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}

		var instance *config.Instance
		// Use instance if specified, otherwise use default
		if instanceName != "" {
			instance, err = cfg.GetInstance(instanceName)
			if err != nil {
				return nil, fmt.Errorf("instance '%s' not found: %w", instanceName, err)
			}
		} else {
			instance, err = cfg.GetDefault()
			if err != nil {
				return nil, fmt.Errorf("no default instance configured: %w", err)
			}
		}

		if fqdn == "" {
			fqdn = instance.FQDN
		}
		if token == "" {
			token = instance.Token
		}
	}

	// Create client
	client := api.NewClient(fqdn, token, api.WithDebug(debug))

	// Set legacy global variables for backward compatibility
	Fqdn = fqdn
	Token = token
	Debug = debug

	return client, nil
}

// CheckLatestVersionOfCli checks for CLI updates
func CheckLatestVersionOfCli() (string, error) {
	lastCheck := viper.GetString("lastupdatechecktime")
	if lastCheck != "" {
		lastCheckTime, err := time.Parse(time.RFC3339, lastCheck)
		if err == nil && lastCheckTime.Add(CheckInterval).After(time.Now()) {
			if Debug {
				log.Println("Skipping update check. Last check was less than 10 minutes ago.")
			}
			return CliVersion, nil
		}
	}

	// Update check time
	viper.Set("lastupdatechecktime", time.Now().Format(time.RFC3339))
	viper.WriteConfig()

	url := "https://api.github.com/repos/coollabsio/coolify-cli/git/refs/tags"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%d - Failed to fetch data from %s. Error: %s", resp.StatusCode, url, string(body))
	}

	var tags []Tag
	if err := json.Unmarshal(body, &tags); err != nil {
		return "", err
	}

	versionsRaw := make([]string, 0, len(tags))
	for _, tag := range tags {
		versionStr := tag.Ref[10:]
		versionsRaw = append(versionsRaw, versionStr)
	}

	versions := make([]*compareVersion.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		v, err := compareVersion.NewVersion(raw)
		if err != nil {
			return "", err
		}
		versions[i] = v
	}

	sort.Sort(compareVersion.Collection(versions))
	latestVersion := versions[len(versions)-1].String()
	if latestVersion != CliVersion {
		fmt.Printf("There is a new version of Coolify CLI available.\nPlease update with 'coolify update'.\n\n")
	}
	return latestVersion, nil
}

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(0)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&Token, "token", "", "", "Token for authentication (https://app.coolify.io/security/api-tokens)")
	rootCmd.PersistentFlags().StringVarP(&Fqdn, "host", "", "", "Coolify instance hostname")
	rootCmd.PersistentFlags().StringVarP(&InstanceName, "instance", "", "", "Use specific instance by name")

	rootCmd.PersistentFlags().StringVarP(&Format, "format", "", "table", "Format output (table|json|pretty)")
	rootCmd.PersistentFlags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")
	rootCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "Force")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "", false, "Debug mode")
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
	latestVersion, err := CheckLatestVersionOfCli()
	if err != nil {
		if Debug {
			log.Println("Failed to check for updates:", err)
		}
	}
	if latestVersion != CliVersion {
		if Debug {
			log.Printf("New version of Coolify CLI is available: %s\n", latestVersion)
		}
	}
}
