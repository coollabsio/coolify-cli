package cliversion

import (
	"github.com/coollabsio/coolify-cli/cmd/runtime"
	"github.com/spf13/cobra"
)

type cliVersion struct {
	coolify runtime.Getter
}

func New(c runtime.Getter) *cliVersion {
	return &cliVersion{
		coolify: c,
	}
}

func (c *cliVersion) NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version ",
		Short: "CLI version",
		Long: `
Print the version of the CLI.
`,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(runtime.Version)
		},
	}

	return cmd
}
