package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var SupportedProviders = []string{"openai", "groq", "ollama", "azure", "anthropic", "watsonx", "mistral", "cohere"}

var (
	DefaultMaxTokens = 4096
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up gptcommit",
	Long:  "Set up gptcommit",
	Run: func(cmd *cobra.Command, args []string) {
		// light blue
		labelStyle := lipgloss.NewStyle().
			Padding(0, 1, 0, 1).
			Background(lipgloss.Color("#24b6ff")).
			Foreground(lipgloss.Color("0"))
		providerPrompt := promptui.Select{
			Label: labelStyle.Render("Select provider"),
			Items: SupportedProviders,
		}
		_, provider, err := providerPrompt.Run()
		if err != nil {
			logger.Errorf("Failed to select provider: %v", err)
		}

		apiKeyPrompt := promptui.Prompt{
			Label: labelStyle.Render("Enter API key"),
			// HideEntered: true,
			Mask: '*',
		}

		apiKey, err := apiKeyPrompt.Run()
		if err != nil {
			logger.Errorf("Failed to enter API key: %v", err)
		}
		if apiKey == "" {
			logger.Error("API key cannot be empty")
			os.Exit(1)
		}

		modelPrompt := promptui.Prompt{
			Label:     "Enter model",
			AllowEdit: true,
			Validate: func(input string) error {
				if input == "" {
					return fmt.Errorf("model cannot be empty")
				}
				return nil
			},
		}

		model, err := modelPrompt.Run()
		if err != nil {
			logger.Errorf("Failed to enter model: %v", err)
		}

		// input max tokens
		maxTokensPrompt := promptui.Prompt{
			Label:     "max_tokens",
			Default:   "4096",
			AllowEdit: true,
			Validate: func(input string) error {
				n, err := strconv.Atoi(input)
				if n < 1 || err != nil {
					return fmt.Errorf("max tokens must be a positive integer greater than 0")
				}
				return err
			},
		}
		maxTokens, err := maxTokensPrompt.Run()
		if err != nil {
			logger.Errorf("Failed to enter max tokens: %v", err)
		}
		maxTokensInt, err := strconv.Atoi(maxTokens)
		if err != nil {
			logger.Errorf("Failed to convert max tokens to integer: %v", err)
		}

		fmt.Println(maxTokensInt)
		fmt.Println(apiKey)
		fmt.Println(provider)
		fmt.Println(model)
	},
}
