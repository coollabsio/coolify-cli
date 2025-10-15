package root

import (
	"fmt"
	"strings"
	"time"

	cliinit "github.com/coollabsio/cli-coolify/internal/cmd/init"
	cliinstances "github.com/coollabsio/cli-coolify/internal/cmd/instances"
	cliprivatekeys "github.com/coollabsio/cli-coolify/internal/cmd/privatekeys"
	cliservers "github.com/coollabsio/cli-coolify/internal/cmd/servers"
	cliupdate "github.com/coollabsio/cli-coolify/internal/cmd/update"
	cliversion "github.com/coollabsio/cli-coolify/internal/cmd/version"
	"github.com/coollabsio/cli-coolify/internal/config"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

// Execute runs the root command
func Execute() error {
	cli := NewCliRoot()
	cmd, err := cli.NewCommand()
	if err != nil {
		return err
	}
	return cmd.Execute()
}

type cliRoot struct {
	outputFormat string
	fqdn         string
	token        string
	instance         string
	timeout      time.Duration
	insecure     bool
	logLevel     string
	coolify      *config.Coolify
}

func NewCliRoot() *cliRoot {
	return &cliRoot{}
}

func (cli *cliRoot) runtime() *config.Coolify {
	return cli.coolify
}

func (cli *cliRoot) initialize() error {
	cli.coolify = config.NewCoolify(cli.fqdn, cli.token, cli.logLevel)

	// Log initialization message
	cli.coolify.LogTrace("Initializing Coolify CLI with log level: %s", cli.logLevel)

	// Use the new load method on the Coolify struct
	return cli.coolify.Load(cli.instance)
}

func (cli *cliRoot) NewCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "coolify",
		Short: "Coolify CLI",
		Long:  `A CLI tool to interact with Coolify API.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize coolify if not already done
			if cli.coolify == nil {
				if err := cli.initialize(); err != nil {
					return fmt.Errorf("failed to initialize: %w", err)
				}
			}

			minCoolifyVersion := "4.0.0-beta.433"
			// List of commands that do not require a version check
			skipVersionCheck := map[string]bool{
				"init":      true,
				"version":   true,
				"update":    true,
				"help":      true,
				"instances": true,
			}

			if skipVersionCheck[cmd.Name()] {
				return nil
			}

			// Check if client is configured
			if cli.coolify.Client == nil {
				return fmt.Errorf("coolify client not configured. Please run 'coolify init' first or provide --host and --token flags")
			}

			resp, err := cli.coolify.Client.VersionWithResponse(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get version from Coolify instance: %w", err)
			}

			versionResp := "0.0.0"
			if resp.StatusCode() == 200 {
				versionResp = strings.TrimSpace(string(resp.Body))
			}

			v, err := version.NewVersion(versionResp)
			if err != nil {
				return fmt.Errorf("failed to parse Coolify instance version: %w", err)
			}

			minV, err := version.NewVersion(minCoolifyVersion)
			if err != nil {
				return fmt.Errorf("failed to parse minimum required Coolify version: %w", err)
			}

			if v.LessThan(minV) {
				return fmt.Errorf("your Coolify instance version '%s' is not supported. Please update to at least version '%s'", v.String(), minV.String())
			}

			return nil
		},
	}

	pFlags := cmd.PersistentFlags()

	pFlags.StringVar(&cli.token, "token", "", "Token for authentication (https://app.coolify.io/security/api-tokens)")
	pFlags.StringVar(&cli.fqdn, "host", "", "Coolify instance hostname EG: https://app.coolify.io")
	pFlags.StringVarP(&cli.instance, "instance", "i", "", "Name of the instance to use from configuration file")
	pFlags.StringVar(&cli.outputFormat, "format", "table", "Format output (table|json|pretty)")
	pFlags.Bool("disableColor", false, "Disable color output for table format")
	pFlags.DurationVar(&cli.timeout, "timeout", 30*time.Second, "HTTP client timeout in seconds")
	pFlags.BoolVar(&cli.insecure, "insecure", false, "Skip TLS verification")
	pFlags.StringVar(&cli.logLevel, "log-level", "info", "Set log level (trace|debug|info|warn|error|fatal|panic)")

	cmd.AddCommand(cliinit.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliinstances.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliversion.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliupdate.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliprivatekeys.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliservers.New(cli.runtime).NewCommand())

	return cmd, nil
}
