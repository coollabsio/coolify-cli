package cliinit

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/cli-coolify/cmd/coolTypes"
	"github.com/coollabsio/cli-coolify/cmd/runtime"
	"github.com/coollabsio/cli-coolify/cmd/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cliInit struct {
	coolify runtime.Getter
}

func New(c runtime.Getter) *cliInit {
	return &cliInit{
		coolify: c,
	}
}

var defaultInstances = []coolTypes.Instance{
	{
		Name:    "cloud",
		Default: true,
		Fqdn:    "https://app.coolify.io",
		Token:   "",
	}, {
		Name:  "localhost",
		Fqdn:  "http://localhost:8000",
		Token: "",
	},
}

func (c *cliInit) NewCommand() *cobra.Command {
	generateDefault := false
	force := false
	cmd := &cobra.Command{
		Use: "init",
		Example: utils.GetCommandExample(`
%[1]s init
%[1]s init --default
%[1]s init --force
`),
		Short: "Initialize a new Coolify CLI configuration file",
		Long: `
Initialize Coolify CLI by generating a configuration file in the default directory.
`,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if c.coolify().Config.JsonExists && !force {
				return errors.New("Configuration file already exists. Please use instances command to make further modifications or force flag to regenerate a new configuration file.")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if generateDefault {
				viper.Set("instances", defaultInstances)
				cmd.Println("Configuration file generated with default instances, use the instances command to make further modifications.")
				return c.coolify().Config.Save()
			}

			// Create a channel to receive the instances
			result := make(chan []coolTypes.Instance)
			p := tea.NewProgram(newInitModel(result))

			// Create a done channel to signal when the program is finished
			done := make(chan struct{})
			var programErr error

			// Run the program in a goroutine
			go func() {
				_, programErr = p.Run()
				close(done)
			}()

			// Wait for either the instances or context cancellation
			var instances []coolTypes.Instance
			select {
			case instances = <-result:
			case <-cmd.Context().Done():
				return fmt.Errorf("operation cancelled")
			case <-done:
				if programErr != nil {
					return fmt.Errorf("program error: %v", programErr)
				}
				return fmt.Errorf("program exited without saving instances")
			}

			viper.Set("instances", instances)
			return c.coolify().Config.Save()
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&generateDefault, "default", "d", false, "Generate a default configuration file (non-interactive)")
	flags.BoolVarP(&force, "force", "f", false, "Force the generation of a new configuration file")

	return cmd
}
