package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// BaseConfig 定义了所有提供者共有的配置字段
type BaseConfig struct {
	Verbose            bool     `mapstructure:"verbose"`
	FileIgnore         []string `mapstructure:"file_ignore"`
	OutputLang         string   `mapstructure:"output.lang"`
	BriefCommitMessage string   `mapstructure:"prompt.brief_commit_message"`
}

// ProviderConfig 定义了特定提供者的配置接口
type ProviderConfig interface {
	LoadConfig(v *viper.Viper)
}

// GroqConfig 定义了 Groq 提供者特定的配置字段
type GroqConfig struct {
	BaseConfig
	APIBase     string  `mapstructure:"api_base"`
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
	TopP        float64 `mapstructure:"top_p"`
}

// OpenAIConfig 定义了 OpenAI 提供者特定的配置字段
type OpenAIConfig struct {
	BaseConfig
	APIBase      string  `mapstructure:"api_base"`
	APIKey       string  `mapstructure:"api_key"`
	ExtraHeaders string  `mapstructure:"extra_headers"`
	Model        string  `mapstructure:"model"`
	Proxy        string  `mapstructure:"proxy"`
	Retries      int     `mapstructure:"retries"`
	Temperature  float64 `mapstructure:"temperature"`
	TopP         float64 `mapstructure:"top_p"`
}

// LoadConfig 实现了 ProviderConfig 接口
func (g *GroqConfig) LoadConfig(v *viper.Viper) {
	v.UnmarshalKey("groq", &g)
}

// LoadConfig 实现了 ProviderConfig 接口
func (o *OpenAIConfig) LoadConfig(v *viper.Viper) {
	v.UnmarshalKey("openai", &o)
}

// Config 定义了整个配置文件的结构
type Config struct {
	Provider string `mapstructure:"provider"`
	BaseConfig
	ProviderConfig
}

func main() {
	// 初始化 Viper
	v := viper.New()
	v.SetConfigName("config") // 配置文件名称（不包括文件扩展名）
	v.SetConfigType("yaml")   // 配置文件类型
	v.AddConfigPath(".")      // 配置文件搜索路径

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// 定义结构体
	var config Config

	// 根据 provider 字段的值，选择相应的配置
	switch config.Provider {
	case "groq":
		groqConfig := &GroqConfig{}
		groqConfig.LoadConfig(v)
		config.BaseConfig = groqConfig.BaseConfig
		config.ProviderConfig = groqConfig
	case "openai":
		openaiConfig := &OpenAIConfig{}
		openaiConfig.LoadConfig(v)
		config.BaseConfig = openaiConfig.BaseConfig
		config.ProviderConfig = openaiConfig
	default:
		log.Fatalf("Unknown provider: %s", config.Provider)
	}

	// 使用配置
	fmt.Printf("Provider: %s\n", config.Provider)
	if groqConfig, ok := config.ProviderConfig.(*GroqConfig); ok {
		fmt.Printf("Groq API Base: %s\n", groqConfig.APIBase)
	}
	if openaiConfig, ok := config.ProviderConfig.(*OpenAIConfig); ok {
		fmt.Printf("OpenAI API Base: %s\n", openaiConfig.APIBase)
	}
}
