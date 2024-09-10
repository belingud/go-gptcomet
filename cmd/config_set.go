package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a config value",
	Long:  "Set a config value in gptcomet.yaml file, e.g. `gptcommit config set openai.model gpt-3.5-turbo`",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		fmt.Println(key, value)
		viper.Set(key, value)
		err := viper.WriteConfig()
		if err != nil {
			logger.Fatal(Red(err.Error()))
		}
		fmt.Println("Config: set", Green(key), "to", Green(value))
	},
}
