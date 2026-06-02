package piwigo

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

// DBConfig holds MySQL connection configuration.
type DBConfig struct {
	Host        string
	Port        int
	DB          string
	User        string
	Password    string
	TablePrefix string
}

var validPrefix = regexp.MustCompile(`^[A-Za-z0-9_]*$`)

// Validate checks the DBConfig for validity.
func (c DBConfig) Validate() error {
	if c.DB == "" {
		return fmt.Errorf("database name required")
	}
	if c.User == "" {
		return fmt.Errorf("user required")
	}
	if !validPrefix.MatchString(c.TablePrefix) {
		return fmt.Errorf("invalid table prefix: must match ^[A-Za-z0-9_]*$")
	}
	return nil
}

// DSN returns the MySQL DSN string.
func DSN(cfg DBConfig) string {
	if cfg.Port == 0 {
		cfg.Port = 3306
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
}

// Open opens a MySQL connection.
func Open(ctx context.Context, cfg DBConfig) (*sql.DB, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	dsn := DSN(cfg)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}
