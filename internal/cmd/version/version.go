package cliversion

import (
	"github.com/coollabsio/cli-coolify/internal/config"
	"github.com/spf13/cobra"
)

type cliVersion struct {
	coolify config.Getter
}

func New(c config.Getter) *cliVersion {
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
			cmd.Println(c.coolify().GetFormattedVersion())
		},
	}

	return cmd
}
