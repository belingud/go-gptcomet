package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// BaseConfig defines configuration fields common to all providers
type BaseConfig struct {
	Verbose            bool     `mapstructure:"verbose"`
	FileIgnore         []string `mapstructure:"file_ignore"`
	OutputLang         string   `mapstructure:"output.lang"`
	BriefCommitMessage string   `mapstructure:"prompt.brief_commit_message"`
}

// ProviderConfig defines the interface for provider-specific configuration
type ProviderConfig interface {
	LoadConfig(v *viper.Viper)
}

// GroqConfig defines Groq provider-specific configuration fields
type GroqConfig struct {
	Base        BaseConfig `mapstructure:",squash"`
	APIBase     string    `mapstructure:"api_base"`
	APIKey      string    `mapstructure:"api_key"`
	Model       string    `mapstructure:"model"`
	MaxTokens   int       `mapstructure:"max_tokens"`
	Temperature float64   `mapstructure:"temperature"`
	TopP        float64   `mapstructure:"top_p"`
}

// OpenAIConfig defines OpenAI provider-specific configuration fields
type OpenAIConfig struct {
	Base         BaseConfig `mapstructure:",squash"`
	APIBase      string    `mapstructure:"api_base"`
	APIKey       string    `mapstructure:"api_key"`
	ExtraHeaders string    `mapstructure:"extra_headers"`
	Model        string    `mapstructure:"model"`
	Proxy        string    `mapstructure:"proxy"`
	Retries      int       `mapstructure:"retries"`
	Temperature  float64   `mapstructure:"temperature"`
	TopP         float64   `mapstructure:"top_p"`
}

// LoadConfig implements the ProviderConfig interface
func (g *GroqConfig) LoadConfig(v *viper.Viper) {
	v.UnmarshalKey("groq", &g)
}

// LoadConfig implements the ProviderConfig interface
func (o *OpenAIConfig) LoadConfig(v *viper.Viper) {
	v.UnmarshalKey("openai", &o)
}

// RootConfig defines the structure of the entire configuration file
type RootConfig struct {
	Provider       string         `mapstructure:"provider"`
	Base          BaseConfig     `mapstructure:",squash"`
	ProviderConf  ProviderConfig `mapstructure:"-"`
}

func LoadConfig() (*RootConfig, error) {
	// Initialize Viper
	v := viper.New()
	v.SetConfigName("config") // Configuration file name (without extension)
	v.SetConfigType("yaml")   // Configuration file type
	v.AddConfigPath(".")      // Configuration file search path

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Define structure
	config := &RootConfig{}

	// Parse provider field first
	if err := v.UnmarshalKey("provider", &config.Provider); err != nil {
		return nil, fmt.Errorf("error unmarshaling provider: %w", err)
	}

	// Choose appropriate configuration based on provider field value
	switch config.Provider {
	case "groq":
		groqConfig := &GroqConfig{}
		groqConfig.LoadConfig(v)
		config.Base = groqConfig.Base
		config.ProviderConf = groqConfig
	case "openai":
		openaiConfig := &OpenAIConfig{}
		openaiConfig.LoadConfig(v)
		config.Base = openaiConfig.Base
		config.ProviderConf = openaiConfig
	default:
		return nil, fmt.Errorf("unknown provider: %s", config.Provider)
	}

	return config, nil
}

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	// Use configuration
	fmt.Printf("Provider: %s\n", config.Provider)
	if groqConfig, ok := config.ProviderConf.(*GroqConfig); ok {
		fmt.Printf("Groq API Base: %s\n", groqConfig.APIBase)
	}
	if openaiConfig, ok := config.ProviderConf.(*OpenAIConfig); ok {
		fmt.Printf("OpenAI API Base: %s\n", openaiConfig.APIBase)
	}
}
