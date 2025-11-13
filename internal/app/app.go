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
	emailService *email.Service
}

type configProvider interface {
	GetAwsConfig() aws.Config
	GetPayloadStoragePath() string
	GetOutboxTableName() string
	GetStaleEmailsThresholdMinutes() int
}

func NewApp(cp configProvider) *App {
	payloadStorage := email.NewPayloadStorage(cp.GetPayloadStoragePath())
	dynamo := dynamodb.NewFromConfig(cp.GetAwsConfig())
	db := email.NewDatabase(dynamo, cp.GetOutboxTableName(), cp.GetStaleEmailsThresholdMinutes())

	emailService := email.NewService(payloadStorage, db)

	return &App{
		emailService: emailService,
	}
}

func (a *App) NewServer(port int) *http.Server {
	mux := http.NewServeMux()

	createEmail := email.NewCreateEmailHandler(a.emailService)
	mux.Handle("POST /emails", createEmail)

	getStaleEmails := email.NewGetStaleEmailsHandler(a.emailService)
	mux.Handle("GET /stale-emails", getStaleEmails)

	getInvalidEmails := email.NewGetInvalidEmailsHandler(a.emailService)
	mux.Handle("GET /invalid-emails", getInvalidEmails)

	requeueEmail := email.NewRequeueEmailHandler(a.emailService)
	mux.Handle("POST /emails/{id}/requeue", requeueEmail)

	health := new(healthcheck.Handler)
	mux.Handle("GET /health-check", health)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
}
