package app

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"multicarrier-email-api/internal/email"
	"multicarrier-email-api/internal/healthcheck"
)

type App struct {
	attachmentsBasePath string
	emailService        *email.Service
}

type configProvider interface {
	GetAwsConfig() aws.Config
	GetAttachmentsBasePath() string
	GetEmlStoragePath() string
}

func NewApp(cp configProvider) *App {
	emlStorage := email.NewEMLStorage(cp.GetEmlStoragePath())
	dynamo := dynamodb.NewFromConfig(cp.GetAwsConfig())
	db := email.New(dynamo)

	emailService := email.NewService(emlStorage, db)

	return &App{
		attachmentsBasePath: cp.GetAttachmentsBasePath(),
		emailService:        emailService,
	}
}

func (a *App) NewServer(port int) *http.Server {
	mux := http.NewServeMux()

	createEmail := email.NewCreateEmailHandler(a.attachmentsBasePath, a.emailService)
	mux.Handle("POST /emails", createEmail)

	health := new(healthcheck.Handler)
	mux.Handle("GET /health-check", health)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
}
