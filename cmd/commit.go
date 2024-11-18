package cmd

import (
	"fmt"

	"github.com/belingud/gptcommit/internal/config"
	"github.com/belingud/gptcommit/internal/generator"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate and create a commit with an AI-generated message",
	Long: `Generate a commit message using AI and create a commit with it.
The message is generated based on the staged changes in your git repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config
		cfg, err := config.NewManager("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create generator
		gen, err := generator.NewGenerator(cfg.Get(), "")
		if err != nil {
			return fmt.Errorf("failed to create generator: %w", err)
		}

		// Generate message and commit
		if err := gen.CommitChanges(); err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
