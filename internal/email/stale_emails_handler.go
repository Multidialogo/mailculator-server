package email

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"multicarrier-email-api/internal/response"
)

type staleEmailsServiceInterface interface {
	GetStaleEmails(ctx context.Context) ([]StaleEmail, error)
}

type GetStaleEmailsHandler struct {
	emailService staleEmailsServiceInterface
}

func NewGetStaleEmailsHandler(emailService staleEmailsServiceInterface) *GetStaleEmailsHandler {
	return &GetStaleEmailsHandler{
		emailService: emailService,
	}
}

func (h *GetStaleEmailsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	staleEmails, err := h.emailService.GetStaleEmails(context.TODO())
	if err != nil {
		slog.Error(fmt.Sprintf("error getting stale emails: %v", err))
		response.WriteError(http.StatusInternalServerError, w, "error getting stale emails")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(staleEmails); err != nil {
		slog.Error(fmt.Sprintf("error encoding response: %v", err))
		response.WriteError(http.StatusInternalServerError, w, "error encoding response")
		return
	}
}
