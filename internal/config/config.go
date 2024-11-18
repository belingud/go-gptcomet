package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for gptcommit
type Config struct {
	Provider     string   `yaml:"provider"`
	APIKey       string   `yaml:"api_key"`
	Model        string   `yaml:"model"`
	APIBase      string   `yaml:"api_base"`
	Language     string   `yaml:"language"`
	FileIgnore   []string `yaml:"file_ignore"`
	Retries      int      `yaml:"retries"`
	Proxy        string   `yaml:"proxy"`
	OutputFormat string   `yaml:"output_format"`
}

const (
	DefaultProvider = "openai"
	DefaultModel    = "gpt-3.5-turbo"
	DefaultAPIBase  = "https://api.openai.com/v1"
	DefaultRetries  = 3
	DefaultLanguage = "en"
)

// Manager handles configuration loading and saving
type Manager struct {
	config     *Config
	configPath string
}

// NewManager creates a new configuration manager
func NewManager(configPath string) (*Manager, error) {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configPath = filepath.Join(home, ".config", "gptcommit", "config.yaml")
	}

	manager := &Manager{
		configPath: configPath,
		config:    &Config{},
	}

	// Create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return nil, err
		}
		manager.config = &Config{
			Provider: DefaultProvider,
			Model:    DefaultModel,
			APIBase:  DefaultAPIBase,
			Retries:  DefaultRetries,
			Language: DefaultLanguage,
		}
		if err := manager.Save(); err != nil {
			return nil, err
		}
	} else {
		if err := manager.Load(); err != nil {
			return nil, err
		}
	}

	return manager, nil
}

// Load reads the configuration from file
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, m.config)
}

// Save writes the configuration to file
func (m *Manager) Save() error {
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// Get returns the current configuration
func (m *Manager) Get() *Config {
	return m.config
}

// Set updates the configuration
func (m *Manager) Set(key string, value interface{}) error {
	// Update config based on key
	switch key {
	case "provider":
		if v, ok := value.(string); ok {
			m.config.Provider = v
		}
	case "api_key":
		if v, ok := value.(string); ok {
			m.config.APIKey = v
		}
	case "model":
		if v, ok := value.(string); ok {
			m.config.Model = v
		}
	case "api_base":
		if v, ok := value.(string); ok {
			m.config.APIBase = v
		}
	case "language":
		if v, ok := value.(string); ok {
			m.config.Language = v
		}
	case "retries":
		if v, ok := value.(int); ok {
			m.config.Retries = v
		}
	case "proxy":
		if v, ok := value.(string); ok {
			m.config.Proxy = v
		}
	case "output_format":
		if v, ok := value.(string); ok {
			m.config.OutputFormat = v
		}
	}

	return m.Save()
}
