package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v3"
)

// Кастомный тип для парсинга размера данных (например, "4KB", "1MB")
type BytesSize int64

func (b *BytesSize) UnmarshalYAML(value *yaml.Node) error {
	str := strings.ToUpper(value.Value)
	var multiplier int64 = 1

	switch {
	case strings.HasSuffix(str, "KB"):
		multiplier = 1024
		str = strings.TrimSuffix(str, "KB")
	case strings.HasSuffix(str, "MB"):
		multiplier = 1024 * 1024
		str = strings.TrimSuffix(str, "MB")
	case strings.HasSuffix(str, "B"):
		str = strings.TrimSuffix(str, "B")
	}

	parsed, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid byte size format '%s': %w", value.Value, err)
	}

	*b = BytesSize(parsed * multiplier)
	return nil
}

// Кастомный тип для парсинга Duration (gopkg.in/yaml.v3 не умеет из коробки парсить time.Duration)
type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	dur, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("invalid duration format '%s': %w", value.Value, err)
	}
	*d = Duration(dur)
	return nil
}

// Основная структура конфигурации
type Config struct {
	Engine  EngineConfig  `yaml:"engine"`
	Network NetworkConfig `yaml:"network"`
	Logging LoggingConfig `yaml:"logging"`
}

type EngineConfig struct {
	Type string `yaml:"type"`
}

type NetworkConfig struct {
	Address        string    `yaml:"address"`
	MaxConnections int       `yaml:"max_connections"`
	MaxMessageSize BytesSize `yaml:"max_message_size"`
	IdleTimeout    Duration  `yaml:"idle_timeout"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Engine: EngineConfig{
			Type: "in_memory",
		},
		Network: NetworkConfig{
			Address:        "0.0.0.0:8080",
			MaxConnections: 1000,
			MaxMessageSize: 1024 * 1024, // 1MB по умолчанию
			IdleTimeout:    Duration(60 * time.Second),
		},
		Logging: LoggingConfig{
			Level:  "info",
			Output: "stdout",
		},
	}
}

// LoadConfig читает файл и накладывает его поверх дефолтных настроек
func LoadConfig(path string) (*Config, error) {
	cfg := NewDefaultConfig()

	file, err := os.Open(path)
	if err != nil {
		// Если файла нет, возвращаем ошибку или можно вернуть дефолтный конфиг
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode yaml: %w", err)
	}

	return cfg, nil
}
