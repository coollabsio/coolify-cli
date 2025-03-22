package cliinit

import (
	"errors"

	"github.com/coollabsio/coolify-cli/cmd/ask"
	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
	"github.com/coollabsio/coolify-cli/cmd/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type runtimeGetter func() *runtime.Coolify

type cliInit struct {
	coolify runtimeGetter
}

func New(c runtimeGetter) *cliInit {
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
		Example: `
coolify-cli init
coolify-cli init --default
coolify-cli init --force
`,
		Short: "Initialize a new Coolify CLI configuration file",
		Long: `
Initialize Coolify CLI by generating a configuration file in the default directory.
`,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if instances := viper.Get("instances"); instances != nil && !force {
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
			instances := make([]coolTypes.Instance, 0)
			cloudAnswer, err := ask.PromptYesOrNo("Do you use the Coolify Cloud service?", true)
			if err != nil {
				return err
			}
			addMore := true
			if cloudAnswer {
				token := ""
				tokenAnswer, err := ask.PromptYesOrNo("Have you generated a token already? (https://app.coolify.io/security/api-tokens)", true)
				if err != nil {
					return err
				}
				if tokenAnswer {
					token, err = ask.PromptString("Please enter your token now")
					if err != nil {
						return err
					}
				}
				instances = append(instances, coolTypes.Instance{
					Default: true,
					Fqdn:    "app.coolify.io",
					Token:   token,
					Name:    "cloud",
				})
				if token == "" {
					cmd.Println("Please generate a token via https://app.coolify.io/security/api-tokens and use the instance command to add it.")
				}
				addMore, err = ask.PromptYesOrNo("Do you want to add any self hosted instances?", false)
				if err != nil {
					return err
				}
			}

			if len(instances) == 0 || cloudAnswer && addMore {
				for addMore {
					fqdn, err := ask.PromptString("Please enter the full fqdn of your Coolify instance EG: https://my.coolify.tld")
					if err != nil {
						return err
					}
					token, err := ask.PromptString("Please enter the token for your Coolify instance")
					if err != nil {
						return err
					}
					name, err := ask.PromptString("Please enter the friendly name of your Coolify instance")
					if err != nil {
						return err
					}
					instances = append(instances, coolTypes.Instance{
						Default: false,
						Fqdn:    fqdn,
						Token:   token,
						Name:    name,
					})
					addMore, err = ask.PromptYesOrNo("Do you want to add more instances?", false)
					if err != nil {
						return err
					}
				}
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
