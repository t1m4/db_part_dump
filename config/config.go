package config

import (
	"fmt"
	"time"

	"github.com/t1m4/db_part_dump/internal/constants"

	"github.com/spf13/viper"
)

var AllowedDbTypes map[string]bool = map[string]bool{
	"postgres": true,
}

type Database struct {
	DBType          string        `mapstructure:"db_type"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	SSLMode         string        `mapstructure:"ssl_mode"`
}
type Filter struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"value"`
}
type Table struct {
	Name    string   `mapstructure:"name"`
	Filters []Filter `mapstructure:"filters"`
}

type Settings struct {
	Output                string   `mapstructure:"output"`
	Format                string   `mapstructure:"format"` // json, sql, or both
	SchemaName            string   `mapstructure:"schema_name"`
	Tables                []Table  `mapstructure:"tables"`
	Direction             string   `mapstructure:"direction"`               // outgoing, incoming
	IncludeIncomingTables []string `mapstructure:"include_incoming_tables"` // Slice of table name for which do search to incoming fks
}

type Config struct {
	Database Database `mapstructure:"database"`
	Settings Settings `mapstructure:"settings"`
}

func (c *Config) Validate() error {
	if _, ok := AllowedDbTypes[c.Database.DBType]; !ok {
		return fmt.Errorf("no supported db type %s", c.Database.DBType)
	}
	return nil
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Settings.Direction == "" {
		config.Settings.Direction = constants.OUTGOING
	}

	return &config, nil
}

func GetDSN(config *Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode,
	)
}
