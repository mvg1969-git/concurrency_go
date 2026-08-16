package config

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// Маршалинг всей конфигурации
func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := enc.AddObject("engine", &c.Engine); err != nil {
		return err
	}
	if err := enc.AddObject("network", &c.Network); err != nil {
		return err
	}
	if err := enc.AddObject("logging", &c.Logging); err != nil {
		return err
	}
	return nil
}

// Маршалинг подсекции Engine
func (e *EngineConfig) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", e.Type)
	return nil
}

// Маршалинг подсекции Network
func (n *NetworkConfig) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", n.Address)
	enc.AddInt("max_connections", n.MaxConnections)
	enc.AddInt64("max_message_size_bytes", int64(n.MaxMessageSize))
	enc.AddDuration("idle_timeout", time.Duration(n.IdleTimeout))
	return nil
}

// Маршалинг подсекции Logging
func (l *LoggingConfig) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("level", l.Level)
	enc.AddString("output", l.Output)
	return nil
}
