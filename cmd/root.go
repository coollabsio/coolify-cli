package cmd

import (
	"log"
	"os"

	"github.com/coollabsio/coolify-cli/cmd/cliinit"
	"github.com/coollabsio/coolify-cli/cmd/cliinstances"
	"github.com/coollabsio/coolify-cli/cmd/cliversion"
	"github.com/coollabsio/coolify-cli/cmd/runtime"
	"github.com/spf13/cobra"
)

var (
	coolify *runtime.Coolify
)

type cliRoot struct {
	logTrace     bool
	logDebug     bool
	logInfo      bool
	logWarn      bool
	logErr       bool
	outputColor  string
	outputFormat string
	fqdn         string
	token        string
	name         string
}

func NewCliRoot() *cliRoot {
	return &cliRoot{}
}

func (cli *cliRoot) runtime() *runtime.Coolify {
	return coolify
}

func (cli *cliRoot) initialize() error {
	coolify = runtime.NewCoolify(cli.fqdn, cli.token)
	return coolify.Config.Load(cli.name)
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

	cmd.AddCommand(cliinit.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliinstances.New(cli.runtime).NewCommand())
	cmd.AddCommand(cliversion.New(cli.runtime).NewCommand())

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
