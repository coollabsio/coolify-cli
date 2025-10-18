package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
)

func NewCompletionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <shell>",
		Short: "Output shell completion code for the specified shell",
		Long: `To load completions:

### Bash

To load completions into the current shell execute:

    source <(coolify completion bash)

In order to make the completions permanent, append the line above to
your .bashrc.

### Zsh

If shell completions are not already enabled for your environment need
to enable them. Add the following line to your ~/.zshrc file:

    autoload -Uz compinit; compinit

To load completions for each session execute the following commands:

    mkdir -p ~/.config/coolify/completion/zsh
    coolify completion zsh > ~/.config/coolify/completion/zsh/_coolify

Finally add the following line to your ~/.zshrc file, *before* you
call the compinit function:

    fpath+=(~/.config/coolify/completion/zsh)

In the end your ~/.zshrc file should contain the following two lines
in the order given here.

    fpath+=(~/.config/coolify/completion/zsh)
    #  ... anything else that needs to be done before compinit
    autoload -Uz compinit; compinit
    # ...

You will need to start a new shell for this setup to take effect.

### Fish

To load completions into the current shell execute:

    coolify completion fish | source

In order to make the completions permanent execute once:

     coolify completion fish > ~/.config/fish/completions/coolify.fish

### PowerShell:

To load completions into the current shell execute:

  PS> coolify completion powershell | Out-String | Invoke-Expression

To load completions for every new session, run 
and source this file from your PowerShell profile.

  PS> coolify completion powershell > coolify.ps1
`,
		Args:                  cli.ExactArgs(1, "<shell>"),
		ValidArgs:             []string{"bash", "fish", "zsh", "powershell"},
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			switch args[0] {
			case "bash":
				err = cmd.Root().GenBashCompletion(os.Stdout)
			case "fish":
				err = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "zsh":
				err = cmd.Root().GenZshCompletion(os.Stdout)
			case "powershell":
				err = cmd.Root().GenPowerShellCompletion(os.Stdout)
			default:
				err = fmt.Errorf("Unsupported shell: %s", args[0])
			}
			return err
		},
	}
	return cmd
}
