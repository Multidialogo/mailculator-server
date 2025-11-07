package email

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"multicarrier-email-api/internal/response"
)

type requeueEmailServiceInterface interface {
	RequeueEmail(ctx context.Context, id string) error
}

type RequeueEmailHandler struct {
	emailService requeueEmailServiceInterface
}

func NewRequeueEmailHandler(emailService requeueEmailServiceInterface) *RequeueEmailHandler {
	return &RequeueEmailHandler{
		emailService: emailService,
	}
}

func (h *RequeueEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.WriteError(http.StatusBadRequest, w, "id parameter is required")
		return
	}

	if err := h.emailService.RequeueEmail(context.TODO(), id); err != nil {
		slog.Error(fmt.Sprintf("error requeuing email: %v", err))
		response.WriteError(http.StatusInternalServerError, w, "error requeuing email")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
