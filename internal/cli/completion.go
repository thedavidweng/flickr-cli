package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for flickr CLI.

To load completions:

Bash:

  $ source <(flickr completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ flickr completion bash > /etc/bash_completion.d/flickr
  # macOS:
  $ flickr completion bash > $(brew --prefix)/etc/bash_completion.d/flickr

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ flickr completion zsh > "${fpath[1]}/_flickr"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ flickr completion fish | source

  # To load completions for each session, execute once:
  $ flickr completion fish > ~/.config/fish/completions/flickr.fish

PowerShell:

  PS> flickr completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> flickr completion powershell > flickr.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(out, true)
		case "zsh":
			return rootCmd.GenZshCompletion(out)
		case "fish":
			return rootCmd.GenFishCompletion(out, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(out)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported shell: %s\n", args[0])
			return nil
		}
	},
}
