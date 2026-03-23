package oracle

import (
	"net"
	"net/url"

	"github.com/kitti12911/lib-database/dbsql"
)

type Config struct {
	Host        string           `mapstructure:"host"         env:"DB_HOST"         validate:"required,hostname|ip"`
	Port        string           `mapstructure:"port"         env:"DB_PORT"         validate:"omitempty,numeric,gte=1,lte=65535"`
	ServiceName string           `mapstructure:"service_name" env:"DB_SERVICE_NAME" validate:"required_without=SID"`
	SID         string           `mapstructure:"sid"          env:"DB_SID"          validate:"required_without=ServiceName"`
	User        string           `mapstructure:"user"         env:"DB_USER"         validate:"required"`
	Password    string           `mapstructure:"password"     env:"DB_PASSWORD"     validate:"required"`
	Wallet      string           `mapstructure:"wallet"       env:"DB_WALLET"`
	SSL         bool             `mapstructure:"ssl"          env:"DB_SSL"`
	SSLVerify   bool             `mapstructure:"ssl_verify"   env:"DB_SSL_VERIFY"`
	Pool        dbsql.PoolConfig `mapstructure:"pool"`
}

func (c Config) connString() string {
	port := c.Port
	if port == "" {
		port = "1521"
	}

	u := &url.URL{
		Scheme: "oracle",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, port),
	}

	if c.ServiceName != "" {
		u.Path = c.ServiceName
	}

	query := url.Values{}

	if c.SID != "" {
		query.Set("SID", c.SID)
	}

	if c.SSL {
		query.Set("ssl", "true")
	}

	if c.SSL && !c.SSLVerify {
		query.Set("ssl verify", "false")
	}

	if c.Wallet != "" {
		query.Set("wallet", c.Wallet)
	}

	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	return u.String()
}
