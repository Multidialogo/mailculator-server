package config

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/go-playground/validator/v10"
)

type AwsConfig struct {
	BaseEndpoint string `yaml:"base-endpoint"`
	sdkConfig    aws.Config
}

type AttachmentsConfig struct {
	BasePath string `yaml:"base-path" validate:"required"`
}

type EmlStorageConfig struct {
	Path string `yaml:"path" validate:"required"`
}

type OutboxConfig struct {
	TableName string `yaml:"table-name" validate:"required"`
}

type ServerConfig struct {
	Port int `yaml:"port" validate:"required"`
}

type Config struct {
	Aws         AwsConfig         `yaml:"aws,flow"`
	Attachments AttachmentsConfig `yaml:"attachments,flow" validate:"required"`
	EmlStorage  EmlStorageConfig  `yaml:"eml-storage,flow" validate:"required"`
	Outbox      OutboxConfig      `yaml:"outbox,flow" validate:"required"`
	Server      ServerConfig      `yaml:"server,flow" validate:"required"`
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

	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	if c.Aws.BaseEndpoint != "" {
		awsConfig.BaseEndpoint = aws.String(c.Aws.BaseEndpoint)
	}

	c.Aws.sdkConfig = awsConfig
	return nil
}

func (c *Config) GetAwsConfig() aws.Config {
	return c.Aws.sdkConfig
}

func (c *Config) GetAttachmentsBasePath() string {
	return c.Attachments.BasePath
}

func (c *Config) GetEmlStoragePath() string {
	return c.EmlStorage.Path
}

func (c *Config) GetOutboxTableName() string {
	return c.Outbox.TableName
}

func (c *Config) GetServerPort() int {
	return c.Server.Port
}
