package cmd

import (
	"log"
	"os"
	"time"

	"github.com/coollabsio/cli-coolify/cmd/cliinit"
	"github.com/coollabsio/cli-coolify/cmd/cliinstances"
	"github.com/coollabsio/cli-coolify/cmd/cliprivatekeys"
	"github.com/coollabsio/cli-coolify/cmd/cliservers"
	"github.com/coollabsio/cli-coolify/cmd/cliupdate"
	"github.com/coollabsio/cli-coolify/cmd/cliversion"
	"github.com/coollabsio/cli-coolify/cmd/runtime"
	"github.com/spf13/cobra"
)

var (
	coolify *runtime.Coolify
)

type cliRoot struct {
	outputFormat string
	fqdn         string
	token        string
	name         string
	timeout      time.Duration
	insecure     bool
	logLevel     string
}

func NewCliRoot() *cliRoot {
	return &cliRoot{}
}

func (cli *cliRoot) runtime() *runtime.Coolify {
	return coolify
}

func (cli *cliRoot) initialize() error {
	coolify = runtime.NewCoolify(cli.fqdn, cli.token, cli.logLevel)

	// Log initialization message
	coolify.LogTrace("Initializing Coolify CLI with log level: %s", cli.logLevel)

	// Use the new load method on the Coolify struct
	return coolify.Load(cli.name)
}

func (cli *cliRoot) NewCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "coolify",
		Short: "Coolify CLI",
		Long:  `A CLI tool to interact with Coolify API.`,
	}

	pFlags := cmd.PersistentFlags()

	pFlags.StringVar(&cli.token, "token", "", "Token for authentication (https://app.coolify.io/security/api-tokens)")
	pFlags.StringVar(&cli.fqdn, "host", "", "Coolify instance hostname EG: https://app.coolify.io")
	pFlags.StringVarP(&cli.name, "name", "n", "", "Name of the instance to use from configuration file")
	pFlags.StringVar(&cli.outputFormat, "format", "table", "Format output (table|json|pretty)")
	pFlags.Bool("disableColor", false, "Disable color output for table format")
	pFlags.DurationVar(&cli.timeout, "timeout", 30*time.Second, "HTTP client timeout")
	pFlags.BoolVar(&cli.insecure, "insecure", false, "Skip TLS verification")
	pFlags.StringVar(&cli.logLevel, "log-level", "info", "Set log level (trace|debug|info|warn|error|fatal|panic)")

	cmd.AddCommand(cliinit.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliinstances.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliversion.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliupdate.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliprivatekeys.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliservers.New(cli.runtime).NewCommand())

	if len(os.Args) > 1 {
		cobra.OnInitialize(
			func() {
				if err := cli.initialize(); err != nil {
					// handle it in future
					log.Println(err)
					os.Exit(1)
				}
			},
		)
	}
	return cmd, nil
}
