package email

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"multicarrier-email-api/internal/response"
)

type invalidEmailsServiceInterface interface {
	GetInvalidEmails(ctx context.Context) ([]Email, error)
}

type GetInvalidEmailsHandler struct {
	emailService invalidEmailsServiceInterface
}

func NewGetInvalidEmailsHandler(emailService invalidEmailsServiceInterface) *GetInvalidEmailsHandler {
	return &GetInvalidEmailsHandler{
		emailService: emailService,
	}
}

func (h *GetInvalidEmailsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	invalidEmails, err := h.emailService.GetInvalidEmails(context.TODO())
	if err != nil {
		slog.Error(fmt.Sprintf("error getting invalid emails: %v", err))
		response.WriteError(http.StatusInternalServerError, w, "error getting invalid emails")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(invalidEmails); err != nil {
		slog.Error(fmt.Sprintf("error encoding response: %v", err))
		response.WriteError(http.StatusInternalServerError, w, "error encoding response")
		return
	}
}

