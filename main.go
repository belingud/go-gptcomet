package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"gptcomet/internal/client"
	"gptcomet/internal/config"
	"gptcomet/internal/git"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var version = "dev"

// Style definitions
var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	boxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("2")).
		Padding(0, 1)
)

func formatCommitMessage(msg string) string {
	return boxStyle.Render(successStyle.Render(msg))
}

func main() {
	var debug bool

	var rootCmd = &cobra.Command{
		Use:     "gptcomet",
		Short:   "GPT Comet - AI-powered Git commit message generator",
		Version: version,
	}

	// Add debug flag to root command
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	rootCmd.AddCommand(newProviderCmd())
	rootCmd.AddCommand(newCommitCmd(&debug))
	rootCmd.AddCommand(newConfigCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readMaskedInput(prompt string) (string, error) {
	var result []byte
	fmt.Print(prompt)

	// Get the terminal file descriptor
	fd := int(syscall.Stdin)
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	for {
		var b [1]byte
		n, err := syscall.Read(fd, b[:])
		if err != nil {
			return "", err
		}
		if n == 0 {
			continue
		}

		switch b[0] {
		case 3: // ctrl+c
			return "", fmt.Errorf("interrupted")
		case 13: // enter
			fmt.Print("\n")
			return string(result), nil
		case 127: // backspace
			if len(result) > 0 {
				result = result[:len(result)-1]
				fmt.Print("\b \b") // erase last * character
			}
		default:
			if b[0] >= 32 { // printable characters
				result = append(result, b[0])
				fmt.Print("*")
			}
		}
	}
}

func newProviderCmd() *cobra.Command {
	const (
		defaultProvider  = "openai"
		defaultAPIBase   = "https://api.openai.com/v1"
		defaultMaxTokens = 1024
		defaultModel     = "gpt-4"
	)

	cmd := &cobra.Command{
		Use:   "newprovider",
		Short: "Add a new API provider interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a reader for user input
			reader := bufio.NewReader(os.Stdin)

			// Get provider name
			fmt.Printf("Enter provider name [%s]: ", defaultProvider)
			provider, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read provider name: %w", err)
			}
			provider = strings.TrimSpace(provider)
			if provider == "" {
				provider = defaultProvider
			}

			// Get API base
			fmt.Printf("Enter API base URL [%s]: ", defaultAPIBase)
			apiBase, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read API base: %w", err)
			}
			apiBase = strings.TrimSpace(apiBase)
			if apiBase == "" {
				apiBase = defaultAPIBase
			}

			// Get API key (with masked input)
			apiKey, err := readMaskedInput("Enter API key: ")
			if err != nil {
				return fmt.Errorf("failed to read API key: %w", err)
			}
			if apiKey == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			// Get model
			fmt.Printf("Enter model name [%s]: ", defaultModel)
			model, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read model: %w", err)
			}
			model = strings.TrimSpace(model)
			if model == "" {
				model = defaultModel
			}

			// Get model max tokens
			fmt.Printf("Enter model max tokens [%d]: ", defaultMaxTokens)
			maxTokensStr, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read max tokens: %w", err)
			}
			maxTokensStr = strings.TrimSpace(maxTokensStr)
			maxTokens := defaultMaxTokens
			if maxTokensStr != "" {
				maxTokens, err = strconv.Atoi(maxTokensStr)
				if err != nil {
					return fmt.Errorf("invalid max tokens value: %w", err)
				}
			}

			// Create config manager
			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			// Check if provider already exists
			if _, exists := cfgManager.Get(provider); exists {
				fmt.Printf("Provider '%s' already exists. Do you want to overwrite it? [y/N]: ", provider)
				answer, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read answer: %w", err)
				}
				answer = strings.ToLower(strings.TrimSpace(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("Operation cancelled")
					return nil
				}
			}

			// Set the provider configuration
			providerConfig := map[string]interface{}{
				"api_key":          apiKey,
				"api_base":         apiBase,
				"model":            model,
				"model_max_tokens": maxTokens,
			}
			if err := cfgManager.Set(provider, providerConfig); err != nil {
				return fmt.Errorf("failed to set provider config: %w", err)
			}

			fmt.Printf("\nProvider configuration saved:\n")
			fmt.Printf("  Provider: %s\n", provider)
			fmt.Printf("  API Base: %s\n", apiBase)
			fmt.Printf("  Model: %s\n", model)
			fmt.Printf("  Max Tokens: %d\n", maxTokens)
			return nil
		},
	}

	return cmd
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	// get command
	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			value, exists := cfgManager.Get(args[0])
			if !exists {
				fmt.Printf("Key '%s' not found in configuration\n", args[0])
				return nil
			}

			data, err := json.MarshalIndent(value, "", "  ")
			if err != nil {
				return err
			}

			fmt.Printf("Value for key '%s':\n%s\n", args[0], string(data))
			return nil
		},
	}

	// list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List config content",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			configStr, err := cfgManager.List()
			if err != nil {
				return fmt.Errorf("failed to list config: %w", err)
			}
			fmt.Print(configStr)
			return nil
		},
	}

	// reset command
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			promptOnly, _ := cmd.Flags().GetBool("prompt")
			if err := cfgManager.Reset(promptOnly); err != nil {
				return err
			}

			if promptOnly {
				fmt.Println("Prompt configuration has been reset to default values")
			} else {
				fmt.Println("Configuration has been reset to default values")
			}
			return nil
		},
	}
	resetCmd.Flags().Bool("prompt", false, "Reset only prompt configuration")

	// set command
	setCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			var value interface{}
			if err := json.Unmarshal([]byte(args[1]), &value); err != nil {
				value = args[1]
			}

			if err := cfgManager.Set(args[0], value); err != nil {
				return err
			}

			fmt.Printf("Successfully set '%s' to: %v\n", args[0], args[1])
			return nil
		},
	}

	// path command
	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Get config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			fmt.Printf("Configuration file path: %s\n", cfgManager.GetPath())
			return nil
		},
	}

	// remove command
	removeCmd := &cobra.Command{
		Use:   "remove [key] [value]",
		Short: "Remove config value or a value from a list",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			// Check if key exists before removing
			_, exists := cfgManager.Get(args[0])
			if !exists {
				fmt.Printf("Key '%s' not found in configuration\n", args[0])
				return nil
			}

			value := ""
			if len(args) > 1 {
				value = args[1]
			}

			if err := cfgManager.Remove(args[0], value); err != nil {
				return err
			}

			if value == "" {
				fmt.Printf("Successfully removed key '%s'\n", args[0])
			} else {
				fmt.Printf("Successfully removed '%s' from '%s'\n", value, args[0])
			}
			return nil
		},
	}

	// append command
	appendCmd := &cobra.Command{
		Use:   "append [key] [value]",
		Short: "Append value to a list config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			var value interface{}
			if err := json.Unmarshal([]byte(args[1]), &value); err != nil {
				value = args[1]
			}

			// Check if key exists and is a list
			current, exists := cfgManager.Get(args[0])
			if exists {
				if _, ok := current.([]interface{}); !ok {
					fmt.Printf("Warning: Key '%s' exists but is not a list. It will be converted to a list.\n", args[0])
				}
			}

			if err := cfgManager.Append(args[0], value); err != nil {
				return err
			}

			fmt.Printf("Successfully appended '%v' to '%s'\n", args[1], args[0])
			return nil
		},
	}

	// keys command
	keysCmd := &cobra.Command{
		Use:   "keys",
		Short: "List all supported configuration keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get config manager
			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			// Get supported keys
			keys := cfgManager.GetSupportedKeys()

			// Print keys
			fmt.Println("Supported configuration keys:")
			for _, key := range keys {
				fmt.Printf("  %s\n", key)
			}
			return nil
		},
	}

	cmd.AddCommand(getCmd, listCmd, resetCmd, setCmd, pathCmd, removeCmd, appendCmd, keysCmd)
	return cmd
}

type textEditor struct {
	textarea textarea.Model
	err      error
}

func (m textEditor) Init() tea.Cmd {
	return textarea.Blink
}

func (m textEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if msg.Alt {
				return m, tea.Quit
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m textEditor) View() string {
	return fmt.Sprintf(
		"Edit commit message (Ctrl+C or Alt+Esc to save and exit):\n\n%s",
		m.textarea.View(),
	)
}

func editText(initialText string) (string, error) {
	// Get terminal width
	width, _, err := term.GetSize(int(syscall.Stdout))
	if err != nil {
		width = 100 // Default width if unable to get terminal size
	}

	ta := textarea.New()
	ta.SetValue(initialText)
	ta.Focus()
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.CharLimit = 4096
	ta.SetWidth(width - 4) // Leave some margin for borders
	ta.SetHeight(10)       // Set a reasonable height

	m := textEditor{
		textarea: ta,
		err:      nil,
	}

	p := tea.NewProgram(m)
	model, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run editor: %w", err)
	}

	finalModel := model.(textEditor)
	if finalModel.err != nil {
		return "", finalModel.err
	}

	return strings.TrimSpace(finalModel.textarea.Value()), nil
}

func newCommitCmd(debug *bool) *cobra.Command {
	var repoPath string
	var rich bool

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Generate and create a commit with staged changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if repoPath == "" {
				var err error
				repoPath, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
			}

			// Check for staged changes
			hasStagedChanges, err := git.HasStagedChanges(repoPath)
			if err != nil {
				return fmt.Errorf("failed to check staged changes: %w", err)
			}
			if !hasStagedChanges {
				return fmt.Errorf("no staged changes found")
			}

			// Get diff
			diff, err := git.GetDiff(repoPath)
			if err != nil {
				return fmt.Errorf("failed to get diff: %w", err)
			}

			// Create config manager
			cfgManager, err := config.New()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			// Get client config
			clientConfig, err := cfgManager.GetClientConfig()
			if err != nil {
				return fmt.Errorf("failed to get client config: %w", err)
			}

			// Set debug mode from global flag
			clientConfig.Debug = *debug

			// Create client
			client := client.New(clientConfig)

			reader := bufio.NewReader(os.Stdin)
			var commitMsg string
			for {
				if commitMsg != "" {
					fmt.Printf("\nCurrent commit message:\n%s\n", formatCommitMessage(commitMsg))
				}

				// Get prompt based on rich flag
				prompt := cfgManager.GetPrompt(rich)

				if commitMsg == "" {
					// Generate commit message
					var err error
					commitMsg, err = client.GenerateCommitMessage(diff, prompt)
					if err != nil {
						return fmt.Errorf("failed to generate commit message: %w", err)
					}
					fmt.Printf("\nGenerated commit message:\n%s\n", formatCommitMessage(commitMsg))
				}

				fmt.Print("\nWhat would you like to do? ([Y]es/[n]o/[r]etry/[e]dit): ")
				answer, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read answer: %w", err)
				}
				answer = strings.ToLower(strings.TrimSpace(answer))

				// If empty answer, use default (yes)
				if answer == "" {
					answer = "y"
				}

				switch answer {
				case "y", "yes":
					// Create commit
					if err := git.Commit(repoPath, commitMsg); err != nil {
						return fmt.Errorf("failed to create commit: %w", err)
					}
					fmt.Printf("\nSuccessfully created commit with message:\n%s\n", formatCommitMessage(commitMsg))
					return nil
				case "n", "no":
					fmt.Println("Operation cancelled")
					return nil
				case "r", "retry":
					commitMsg = ""
					continue
				case "e", "edit":
					edited, err := editText(commitMsg)
					if err != nil {
						fmt.Printf("Error editing message: %v\n", err)
						continue
					}
					commitMsg = edited
					continue
				default:
					fmt.Println("Invalid option, please try again")
					continue
				}
			}
		},
	}

	cmd.Flags().StringVarP(&repoPath, "path", "p", "", "Repository path")
	cmd.Flags().BoolVarP(&rich, "rich", "r", false, "Generate rich commit message with details")
	return cmd
}
