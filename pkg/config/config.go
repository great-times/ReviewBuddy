package config

import (
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Web      WebConfig      `mapstructure:"web"`
	Database DatabaseConfig `mapstructure:"database"`
	Agent    AgentConfig    `mapstructure:"agent"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type WebConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type AgentConfig struct {
	Provider       string `mapstructure:"provider"`
	BaseURL        string `mapstructure:"base_url"`
	APIKey         string `mapstructure:"api_key"`
	Model          string `mapstructure:"model"`
	EmbeddingModel string `mapstructure:"embedding_model"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
	SystemPrompt   string `mapstructure:"system_prompt"`
}

// Load 按优先级加载配置：命令行 path → 环境变量 REVIEWBUDDY_CONFIG/CHANGEBUDDY_CONFIG → configs
func Load(path string) (*Config, error) {
	v := viper.New()
	setDefaults(v)

	candidates := []string{}
	if path != "" {
		candidates = append(candidates, path)
	}
	if env := os.Getenv("REVIEWBUDDY_CONFIG"); env != "" {
		candidates = append(candidates, env)
	}
	if env := os.Getenv("CHANGEBUDDY_CONFIG"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates, "configs/config.yaml", "configs/config.yaml.example")

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			v.SetConfigFile(c)
			if err := v.ReadInConfig(); err != nil {
				return nil, err
			}
			break
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 26405)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("web.port", 26406)
	v.SetDefault("database.path", "./data/reviewbuddy.db")
	v.SetDefault("agent.provider", "mock")
	v.SetDefault("agent.model", "hermes-3")
	v.SetDefault("agent.timeout_seconds", 120)
}
