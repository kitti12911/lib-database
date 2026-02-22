package database

import (
	"fmt"
	"time"
)

type Config struct {
	Host     string     `mapstructure:"host"     env:"DB_HOST"     validate:"required,hostname|ip"`
	Port     string     `mapstructure:"port"     env:"DB_PORT"     validate:"required,numeric,gte=1,lte=65535"`
	User     string     `mapstructure:"user"     env:"DB_USER"     validate:"required"`
	Password string     `mapstructure:"password" env:"DB_PASSWORD" validate:"required"`
	Database string     `mapstructure:"database" env:"DB_DATABASE" validate:"required"`
	Pool     PoolConfig `mapstructure:"pool"`
}

type PoolConfig struct {
	MaxConns        int32         `mapstructure:"maxConns"        validate:"omitempty,gte=1"`
	MinConns        int32         `mapstructure:"minConns"        validate:"omitempty,gte=0,ltefield=MaxConns"`
	MaxConnLifetime time.Duration `mapstructure:"maxConnLifeTime" validate:"omitempty,gt=0"`
	MaxConnIdleTime time.Duration `mapstructure:"maxConnIdleTime" validate:"omitempty,gt=0"`
}

func (c Config) connString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		c.User, c.Password, c.Host, c.Port, c.Database,
	)
}
