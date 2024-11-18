package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/belingud/gptcommit/internal/config"
	"github.com/belingud/gptcommit/internal/logger"
	"path/filepath"
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
		var providerConfig config.ProviderConfig
		// provider := viper.GetString("provider")
		viper.GetStringSlice("file_ignore")

		fmt.Println(providerConfig)
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
			Label: "API key",
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

		// Create config directory if it doesn't exist
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Errorf("Failed to get home directory: %v", err)
			os.Exit(1)
		}
		configDir := filepath.Join(home, ".config", "gptcommit")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			logger.Errorf("Failed to create config directory: %v", err)
			os.Exit(1)
		}

		// Create and save configuration
		configPath := filepath.Join(configDir, "gptcommit.yaml")
		cfg := map[string]interface{}{
			"provider": provider,
			"api_key": apiKey,
			"model": model,
			"max_tokens": maxTokensInt,
		}

		// Write config to file
		data, err := yaml.Marshal(cfg)
		if err != nil {
			logger.Errorf("Failed to marshal config: %v", err)
			os.Exit(1)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			logger.Errorf("Failed to write config file: %v", err)
			os.Exit(1)
		}

		logger.Info("Configuration saved successfully!")
	},
}
