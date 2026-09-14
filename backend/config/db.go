package config

import (
	"fmt"
)

type PostgresConfig struct {
	User     string
	Password string
	DBName   string
	Host     string
	Port     string
}

type DB interface {
	ConnectString() string
}

func (p *PostgresConfig) ConnectString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", p.User, p.Password, p.Host, p.Port, p.DBName)
}
