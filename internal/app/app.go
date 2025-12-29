package app

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"

	"multicarrier-email-api/internal/email"
	"multicarrier-email-api/internal/healthcheck"
)

type App struct {
	emailService *email.Service
	db           *sql.DB
}

type configProvider interface {
	GetMySQLDSN() string
	GetPayloadStoragePath() string
	GetStaleEmailsThresholdMinutes() int
}

func NewApp(cp configProvider) (*App, error) {
	db, err := sql.Open("mysql", cp.GetMySQLDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	payloadStorage := email.NewPayloadStorage(cp.GetPayloadStoragePath())
	emailDB := email.NewDatabase(db, cp.GetStaleEmailsThresholdMinutes())

	emailService := email.NewService(payloadStorage, emailDB)

	return &App{
		emailService: emailService,
		db:           db,
	}, nil
}

func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
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
