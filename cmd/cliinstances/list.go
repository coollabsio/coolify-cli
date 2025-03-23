package cliinstances

import (
	"os"

	"github.com/coollabsio/coolify-cli/cmd/clitable"
	"github.com/coollabsio/coolify-cli/cmd/emoji"
	"github.com/spf13/cobra"
)

func (c *cliInstances) newListCommand() *cobra.Command {
	sensitive := false
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all instances",
		Long: `
List all instances from the CLI configuration file.
`,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			disableColor, err := cmd.Flags().GetBool("disableColor")
			if err != nil {
				return err
			}
			t := clitable.NewStyledTable(os.Stdout, disableColor)
			t.SetHeaders("Name", "URL", "Token", "Default")
			for _, instance := range c.instances {
				token := instance.Token
				if !sensitive && token != "" {
					token = "********" // !TODO implement either utils string or emoji for this
				}
				e := emoji.CrossMark
				if instance.Default {
					e = emoji.CheckMarkButton
				}
				t.AddRow(instance.Name, instance.Fqdn, token, e)
			}

			t.Render()
			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&sensitive, "sensitive", "s", false, "Show sensitive information such as tokens")

	return cmd
}
