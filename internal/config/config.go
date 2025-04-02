package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/go-playground/validator/v10"
)

type AwsConfig struct {
	BaseEndpoint string `yaml:"base-endpoint"`
	Key          string `yaml:"key" validate:"required"`
	Secret       string `yaml:"secret" validate:"required"`
	Region       string `yaml:"region" validate:"required"`
}

type AttachmentsConfig struct {
	BasePath string `yaml:"base-path" validate:"required"`
}

type EmlStorageConfig struct {
	Path string `yaml:"path" validate:"required"`
}

type ServerConfig struct {
	Port int `yaml:"port" validate:"required"`
}

type Config struct {
	Aws         AwsConfig         `yaml:"aws,flow" validate:"required"`
	Attachments AttachmentsConfig `yaml:"attachments,flow" validate:"required"`
	EmlStorage  EmlStorageConfig  `yaml:"eml-storage,flow" validate:"required"`
	Server      ServerConfig      `yaml:"server,flow" validate:"required"`
}

func NewFromYaml(filePath string) (*Config, error) {
	config := &Config{}

	yamlData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	yamlString := os.ExpandEnv(string(yamlData))
	reader := strings.NewReader(yamlString)

	if err := config.Load(reader); err != nil {
		return nil, err
	}

	return config, nil
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
	return err
}

func (c *Config) getAwsCredentialsProvider() credentials.StaticCredentialsProvider {
	return credentials.NewStaticCredentialsProvider(
		c.Aws.Key,
		c.Aws.Secret,
		"",
	)
}

func (c *Config) GetAwsConfig() aws.Config {
	cfg := aws.Config{
		Region:      c.Aws.Region,
		Credentials: c.getAwsCredentialsProvider(),
	}

	if c.Aws.BaseEndpoint != "" {
		cfg.BaseEndpoint = aws.String(c.Aws.BaseEndpoint)
	}

	return cfg
}

func (c *Config) GetAttachmentsBasePath() string {
	return c.Attachments.BasePath
}

func (c *Config) GetEmlStoragePath() string {
	return c.EmlStorage.Path
}

func (c *Config) GetServerPort() int {
	return c.Server.Port
}
