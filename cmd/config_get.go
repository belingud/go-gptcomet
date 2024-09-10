package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getConfigAsString(key string) (interface{}, error) {
	rawValue := viper.Get(key)

	valueType := reflect.TypeOf(rawValue)
	logger.Debugf("key: %s, type: %s", key, valueType.String())
	if valueType == nil {
		return nil, fmt.Errorf("key '%s' not found", key)
	}

	// convert raw value to string
	switch valueType.Kind() {
	case reflect.String:
		return rawValue.(string), nil
	case reflect.Map:
		mapValue := rawValue.(map[string]interface{})
		jsonData, err := json.MarshalIndent(mapValue, "", "  ")
		if err != nil {
			return nil, err
		}
		return string(jsonData), nil
	case reflect.Slice:
		sliceValue := rawValue.([]interface{})
		jsonData, err := json.MarshalIndent(sliceValue, "", "  ")
		if err != nil {
			return nil, err
		}
		return string(jsonData), nil
	case reflect.Int, reflect.Bool, reflect.Float64:
		return fmt.Sprintf("%v", rawValue), nil
	default:
		return nil, fmt.Errorf("unsupported type for key '%s'", key)
	}
}

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a config value",
	Long:  "Get a config value from gptcomet.yaml file, e.g. `gptcommit config get openai.model`",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		val, err := getConfigAsString(args[0])
		if err != nil {
			fmt.Println(Red(err.Error()))
			return
		}
		output := fmt.Sprintf("%s: %s", Green(args[0]), Green(val))
		fmt.Println(output)
	},
}
