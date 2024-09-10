package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "GPTCommit config management",
	Long: `GPTCommit config commands provide access to GPTCommit configuration settings.
Default config file is $HOME/.config/gptcommit/gptcommit.yaml`,
	Example: "gmsg config get openai.model",
	Aliases: []string{"conf"},
	Args:    cobra.MinimumNArgs(1),
}

func init() {
	// gmsg conf get <key>
	configCmd.AddCommand(getCmd)
	// gmsg conf set <key> <value>
	configCmd.AddCommand(setCmd)
}
