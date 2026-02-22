package mssql

import (
	"net"
	"net/url"

	"github.com/kitti12911/lib-database/dbsql"
)

type Config struct {
	Host                   string           `mapstructure:"host"                     env:"DB_HOST"                       validate:"required,hostname|ip"`
	Port                   string           `mapstructure:"port"                     env:"DB_PORT"                       validate:"omitempty,numeric,gte=1,lte=65535"`
	Instance               string           `mapstructure:"instance"                 env:"DB_INSTANCE"`
	User                   string           `mapstructure:"user"                     env:"DB_USER"                       validate:"required"`
	Password               string           `mapstructure:"password"                 env:"DB_PASSWORD"                   validate:"required"`
	Database               string           `mapstructure:"database"                 env:"DB_DATABASE"                   validate:"required"`
	Encrypt                string           `mapstructure:"encrypt"                  env:"DB_ENCRYPT"`
	TrustServerCertificate bool             `mapstructure:"trust_server_certificate" env:"DB_TRUST_SERVER_CERTIFICATE"`
	Pool                   dbsql.PoolConfig `mapstructure:"pool"`
}

func (c Config) connString() string {
	query := url.Values{}
	query.Set("database", c.Database)

	if c.Encrypt != "" {
		query.Set("encrypt", c.Encrypt)
	}
	if c.TrustServerCertificate {
		query.Set("TrustServerCertificate", "true")
	}

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(c.User, c.Password),
		RawQuery: query.Encode(),
	}

	if c.Instance != "" {
		u.Host = c.Host
		u.Path = c.Instance
	} else {
		port := c.Port
		if port == "" {
			port = "1433"
		}
		u.Host = net.JoinHostPort(c.Host, port)
	}

	return u.String()
}
