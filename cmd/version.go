package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gptcommit",
	Long:  `All software has versions. This is gptcommit's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gptcommit Static Site Generator v0.0.1 -- HEAD")
	},
}
