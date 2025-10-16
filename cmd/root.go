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
var CliVersion = "1.0.1"

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
	ContextName string
	Debug   bool
	ShowSensitive bool
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
	Long:  fmt.Sprintf("A CLI tool to interact with Coolify API.\nVersion: %s", CliVersion),
	SilenceUsage: true, // Don't show usage on errors
	SilenceErrors: false, // Still print errors
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// getAPIClient creates an API client from command flags or config
func getAPIClient(cmd *cobra.Command) (*api.Client, error) {
	// Get flags
	token, _ := cmd.Flags().GetString("token")
	contextName, _ := cmd.Flags().GetString("context")
	debug, _ := cmd.Flags().GetBool("debug")

	// Load config to get instance details
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var instance *config.Instance
	// Use context if specified, otherwise use default
	if contextName != "" {
		instance, err = cfg.GetInstance(contextName)
		if err != nil {
			return nil, fmt.Errorf("context '%s' not found: %w", contextName, err)
		}
	} else {
		instance, err = cfg.GetDefault()
		if err != nil {
			return nil, fmt.Errorf("no default instance configured: %w", err)
		}
	}

	// Get FQDN from instance
	fqdn := instance.FQDN

	// Use token from flag if provided, otherwise use instance token
	if token == "" {
		token = instance.Token
	}

	// Create client
	client := api.NewClient(fqdn, token, api.WithDebug(debug))

	// Set legacy global variables for backward compatibility
	Fqdn = fqdn
	Token = token
	Debug = debug

	return client, nil
}

// exactArgs returns a validator that ensures exactly n arguments are provided with a helpful error message
func exactArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			if n == 1 {
				return fmt.Errorf("missing required argument: %s\n\nUsage: %s", usage, cmd.UseLine())
			}
			return fmt.Errorf("expected %d argument(s), got %d\n\nUsage: %s", n, len(args), cmd.UseLine())
		}
		return nil
	}
}

// minArgs returns a validator that ensures at least n arguments are provided with a helpful error message
func minArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return fmt.Errorf("missing required arguments: %s\n\nUsage: %s", usage, cmd.UseLine())
		}
		return nil
	}
}

// parseInt parses a string to int with better error message
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0, fmt.Errorf("'%s' is not a valid integer", s)
	}
	return result, nil
}

// splitOwnerRepo splits owner/repo string into parts
func splitOwnerRepo(s string) []string {
	parts := make([]string, 0, 2)
	var current string
	for _, char := range s {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
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
	latestVersion := versions[len(versions)-1]

	// Compare versions properly using semantic versioning
	currentVersion, err := compareVersion.NewVersion(CliVersion)
	if err != nil {
		return latestVersion.String(), err
	}

	if latestVersion.GreaterThan(currentVersion) {
		fmt.Printf("There is a new version of Coolify CLI available.\nPlease update with 'coolify update'.\n\n")
	}
	return latestVersion.String(), nil
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

	rootCmd.PersistentFlags().StringVarP(&Token, "token", "", "", "Token for authentication (override context token)")
	rootCmd.PersistentFlags().StringVarP(&ContextName, "context", "", "", "Use specific context by name")

	rootCmd.PersistentFlags().StringVarP(&Format, "format", "", "table", "Format output (table|json|pretty)")
	rootCmd.PersistentFlags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")
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
	latestVersionStr, err := CheckLatestVersionOfCli()
	if err != nil {
		if Debug {
			log.Println("Failed to check for updates:", err)
		}
	}

	// Compare versions properly using semantic versioning
	if latestVersionStr != "" {
		latestVersion, err := compareVersion.NewVersion(latestVersionStr)
		if err == nil {
			currentVersion, err := compareVersion.NewVersion(CliVersion)
			if err == nil && latestVersion.GreaterThan(currentVersion) {
				if Debug {
					log.Printf("New version of Coolify CLI is available: %s\n", latestVersionStr)
				}
			}
		}
	}
}
