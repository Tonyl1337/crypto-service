package config

import "time"

type Config struct {
	HTTP      HTTPConfig      `yaml:"http"`
	Database  DatabaseConfig  `yaml:"database"`
	Telegram  TelegramConfig  `yaml:"telegram"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
}

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type TelegramConfig struct {
	Token string `yaml:"token"`
}

type SchedulerConfig struct {
	Interval time.Duration `yaml:"interval"`
}
