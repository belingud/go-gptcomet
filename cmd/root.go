/*
Copyright © 2024 belingud <im.victor@qq.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Red                = color.New(color.FgRed).SprintFunc()
	Green              = color.New(color.FgGreen).SprintFunc()
	logger *log.Logger = log.NewWithOptions(
		os.Stderr, log.Options{
			ReportTimestamp: false,
			ReportCaller:    false,
			Prefix:          "[GPTCommit]",
			Level:           log.InfoLevel,
		})
	styles = log.DefaultStyles()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "gptcommit",
	Aliases: []string{"gmsg"},
	Short:   "AI powered commit message generator",
	Long: `GPTCommit is an AI powered commit message generator.
Support for multiple providers, utilizing 'git diff —staged' information to craft commit messages,
with capabilities for editing and submission.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	LocalInit := func() {
		initConfig(rootCmd)
	}
	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR!!").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("204")).
		Foreground(lipgloss.Color("0"))

	cobra.OnInitialize(LocalInit)

	var cfgFile string
	var localMode bool
	var Verbose bool

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/gptcommit.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&localMode, "local", "l", false, "local mode, use .git/gptcommit.yaml")

	// config manage subcommand
	rootCmd.AddCommand(configCmd)
	// setup subcommand
	rootCmd.AddCommand(setupCmd)
	// version subcommand
	rootCmd.AddCommand(versionCmd)
}

func initConfig(rootCmd *cobra.Command) {
	// localMode := viper.GetBool("local")
	localMode, err := rootCmd.Flags().GetBool("local")
	if err != nil {
		logger.Fatal(err)
	}
	cfgFile, err := rootCmd.Flags().GetString("config")
	if err != nil {
		logger.Fatal(err)
	}
	if cfgFile == "" {
		if localMode {
			// Use .git/gptcommit.yaml if local mode is enabled
			cfgFile = ".git/gptcommit.yaml"
		} else {
			// Use $HOME/.config/gptcommit/gptcommit.yaml
			// Find home directory.
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			// Search config in home directory with name "~/.config/gptcommit/gptcommit.yaml".
			cfgFile = filepath.Join(home, ".config", "gptcommit", "gptcommit.yaml")
		}
	}
	viper.SetConfigFile(cfgFile)
	if _, err := os.Stat(viper.ConfigFileUsed()); os.IsNotExist(err) {
		logger.Errorf("Config file %s not found, please use `gptcommit setup` to set up", viper.ConfigFileUsed())
		os.Exit(1)
	}

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		logger.Fatal(fmt.Errorf("%s", err))
	}
	verbose, err := rootCmd.Flags().GetBool("verbose")
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	if verbose {
		// Set log level to debug, print verbose log
		logger.SetLevel(log.DebugLevel)
		logger.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}
}
