package instances

import (
	"errors"
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/cli-coolify/internal/config"
	"github.com/coollabsio/cli-coolify/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *cliInstances) newAddCommand() *cobra.Command {
	force := false
	isNewDefault := false
	cmd := &cobra.Command{
		Use: "add [name] [fqdn] [token]",
		Example: utils.GetCommandExample(`
%[1]s instances add MyInstance https://my.instance.tld 1234
%[1]s instances add AnotherInstance https://another.instance.tld 5678 --default
%[1]s instances add MyInstance https://my.instance.tld 91011 --force
%[1]s instances add  # Interactive mode
`),
		Short: "Add a new instance",
		Long: `
Add a new instance to the CLI configuration file.
If no arguments are provided, an interactive form will be shown.
`,
		Aliases:      []string{"create"},
		SilenceUsage: true,
		Args:         cobra.RangeArgs(0, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return c.runInteractiveMode(cmd, force, isNewDefault)
			} else if len(args) != 3 {
				return errors.New("command requires either 0 arguments (interactive mode) or exactly 3 arguments (name, fqdn, token)")
			}
			return c.runNonInteractiveMode(args, force, isNewDefault)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&force, "force", "f", false, "Force overwrite existing instance with the same name")
	flags.BoolVarP(&isNewDefault, "default", "d", false, "Set this instance as the default instance")

	return cmd
}

func (c *cliInstances) runInteractiveMode(cmd *cobra.Command, force, isDefault bool) error {
	result := make(chan config.Instance)
	p := tea.NewProgram(newAddModel(result, force, isDefault))

	// Create a done channel to signal when the program is finished
	done := make(chan struct{})
	var programErr error

	// Run the program in a goroutine
	go func() {
		_, programErr = p.Run()
		close(done)
	}()

	// Wait for either the instance or context cancellation
	var instance config.Instance
	select {
	case instance = <-result:
	case <-cmd.Context().Done():
		return fmt.Errorf("operation cancelled")
	case <-done:
		if programErr != nil {
			return fmt.Errorf("program error: %v", programErr)
		}
		return fmt.Errorf("program exited without saving instance")
	}

	// Check for existing instance with same name
	for i, existing := range c.instances {
		if existing.Name == instance.Name {
			if !force {
				return errors.New("instance with the same name already exists. Use the force flag to overwrite or instances set to modify individual attributes")
			}
			c.instances = slices.Delete(c.instances, i, i+1)
			break
		}
	}

	if isDefault {
		for i := range c.instances {
			c.instances[i].Default = false
		}
	}

	c.instances = append(c.instances, instance)
	viper.Set("instances", c.instances)
	return c.coolify().Save()
}

func (c *cliInstances) runNonInteractiveMode(args []string, force, isNewDefault bool) error {
	// Check for existing instance with same name
	for i, instance := range c.instances {
		if instance.Name == args[0] {
			if !force {
				return errors.New("instance with the same name already exists. Use the force flag to overwrite or instances set to modify individual attributes")
			}
			c.instances = slices.Delete(c.instances, i, i+1)
			break
		}
	}

	newInstance := config.Instance{
		Name:    args[0],
		Fqdn:    args[1],
		Token:   args[2],
		Default: isNewDefault,
	}

	if isNewDefault {
		for i := range c.instances {
			c.instances[i].Default = false
		}
	}

	c.instances = append(c.instances, newInstance)
	viper.Set("instances", c.instances)
	return c.coolify().Save()
}
