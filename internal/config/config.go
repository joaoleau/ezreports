package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Global    Global     `mapstructure:"global"`
	Receivers []Receiver `mapstructure:"receivers"`
	Reports   []Report   `mapstructure:"reports"`
}

type Global struct {
	SMTPConfig SMTPConfig `mapstructure:"smtp_config"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"smtp_host"`
	Port     int    `mapstructure:"smtp_port"`
	From     string `mapstructure:"smtp_from"`
	Username string `mapstructure:"smtp_username"`
	Password string `mapstructure:"smtp_password"`
}

type Receiver struct {
	Name  string 			  `mapstructure:"name"`
	EmailConfigs EmailConfigs	  `mapstructure:"email_configs"`
}

type EmailConfigs struct {
	Subject string `mapstructure:"subject"`
	To []string `mapstructure:"to"`
	HTMLBody string `mapstructure:"html_body"`
}

type Report struct {
	Title    string  `mapstructure:"title"`
	Receiver string  `mapstructure:"receiver"`
	Panels   []Panel `mapstructure:"panels"`
	Format   string  `mapstructure:"format"`
	Template string  `mapstructure:"template"`
}

type Panel struct {
	Title       string   `mapstructure:"title"`
	DashboardID string   `mapstructure:"dashboard_id"`
	Description string   `mapstructure:"description"`
	Params      []string `mapstructure:"params"`
}


func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	v := viper.New()
	v.SetConfigName("ezreports_config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
