package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewCompletionCmd(root *cobra.Command) *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for cka-lab-runner.

To load completions:

Bash:

  $ source <(cka-lab-runner completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ cka-lab-runner completion bash > /etc/bash_completion.d/cka-lab-runner
  # macOS:
  $ cka-lab-runner completion bash > $(brew --prefix)/etc/bash_completion.d/cka-lab-runner

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ cka-lab-runner completion zsh > "${fpath[1]}/_cka-lab-runner"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ cka-lab-runner completion fish | source

  # To load completions for each session, execute once:
  $ cka-lab-runner completion fish > ~/.config/fish/completions/cka-lab-runner.fish

PowerShell:

  PS> cka-lab-runner completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> cka-lab-runner completion powershell > cka-lab-runner.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}

	return completionCmd
}
