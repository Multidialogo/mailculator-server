package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/go-playground/validator/v10"
)

type MySQLConfig struct {
	Host     string `yaml:"host" validate:"required"`
	Port     int    `yaml:"port" validate:"required"`
	User     string `yaml:"user" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Database string `yaml:"database" validate:"required"`
}

type PayloadStorageConfig struct {
	Path string `yaml:"path" validate:"required"`
}

type OutboxConfig struct {
	StaleEmailsThresholdMinutes int `yaml:"stale-emails-threshold-minutes" validate:"required"`
}

type ServerConfig struct {
	Port int `yaml:"port" validate:"required"`
}

type Config struct {
	MySQL          MySQLConfig          `yaml:"mysql,flow" validate:"required"`
	PayloadStorage PayloadStorageConfig `yaml:"payload-storage,flow" validate:"required"`
	Outbox         OutboxConfig         `yaml:"outbox,flow" validate:"required"`
	Server         ServerConfig         `yaml:"server,flow" validate:"required"`
}

func NewFromYamlContent(yamlContent []byte) (*Config, error) {
	cfg := &Config{}
	yamlString := os.ExpandEnv(string(yamlContent))
	reader := strings.NewReader(yamlString)

	if err := cfg.Load(reader); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Load(r io.Reader) error {
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)

	decodeErr := decoder.Decode(c)
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(c)

	if decodeErr != nil && err != nil {
		return fmt.Errorf("%w\n%w", err, decodeErr)
	}
	if decodeErr != nil {
		return decodeErr
	}
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.MySQL.User,
		c.MySQL.Password,
		c.MySQL.Host,
		c.MySQL.Port,
		c.MySQL.Database,
	)
}

func (c *Config) GetPayloadStoragePath() string {
	return c.PayloadStorage.Path
}

func (c *Config) GetStaleEmailsThresholdMinutes() int {
	return c.Outbox.StaleEmailsThresholdMinutes
}

func (c *Config) GetServerPort() int {
	return c.Server.Port
}
