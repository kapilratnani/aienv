package cmd

import (
	"fmt"

	"github.com/kapilratnani/aienv/internal/shell"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Install shell function for activation",
	Long: `Installs the 'aienv' shell function to your .bashrc or .zshrc.
This enables 'aienv <name>' to activate environments without typing 'eval'.

After running, restart your shell or run 'source ~/.zshrc'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sh := shell.Detect()
		rcPath := shell.RCFile(sh)

		installed, err := shell.IsInstalled(rcPath)
		if err != nil {
			return fmt.Errorf("checking installation: %w", err)
		}
		if installed {
			fmt.Printf("aienv function already installed in %s\n", rcPath)
			return nil
		}

		if err := shell.Install(rcPath); err != nil {
			return fmt.Errorf("installing shell function: %w", err)
		}

		fmt.Printf("Installed aienv function to %s\n", rcPath)
		fmt.Printf("Run 'source %s' or restart your shell to use it.\n", rcPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
