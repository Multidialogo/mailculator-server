package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"multicarrier-email-api/internal/response"

	"github.com/go-playground/validator/v10"
)

type scanAndSetTTLRequest struct {
	TTLTimestamp int64 `json:"ttl_timestamp" validate:"required,min=1"`
	MaxRecords   int   `json:"max_records" validate:"required,min=1,max=10000"`
}

type scanAndSetTTLServiceInterface interface {
	ScanAndSetTTL(ctx context.Context, ttlTimestamp int64, maxRecords int) (*ScanAndSetTTLResult, error)
}

type ScanAndSetTTLHandler struct {
	emailService scanAndSetTTLServiceInterface
}

func NewScanAndSetTTLHandler(emailService scanAndSetTTLServiceInterface) *ScanAndSetTTLHandler {
	return &ScanAndSetTTLHandler{
		emailService: emailService,
	}
}

func (h *ScanAndSetTTLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteError(http.StatusBadRequest, w, fmt.Sprintf("error reading request body: %v", err))
		return
	}

	slog.Info(fmt.Sprintf("received scan-and-set-ttl request body: %v", string(body)))

	buf := bytes.NewBuffer(body)
	decoder := json.NewDecoder(buf)

	var request scanAndSetTTLRequest
	if err := decoder.Decode(&request); err != nil {
		slog.Error(fmt.Sprintf("bad scan-and-set-ttl payload: %v", string(body)))
		response.WriteError(http.StatusBadRequest, w, fmt.Sprintf("error unmarshalling request body: %v", err))
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(request); err != nil {
		response.WriteError(http.StatusBadRequest, w, fmt.Sprintf("error validating request body: %v", err))
		return
	}

	result, err := h.emailService.ScanAndSetTTL(context.TODO(), request.TTLTimestamp, request.MaxRecords)
	if err != nil {
		slog.Error(fmt.Sprintf("error scanning and setting TTL: %v", err))
		response.WriteError(http.StatusInternalServerError, w, "error scanning and setting TTL")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		slog.Error(fmt.Sprintf("error encoding response: %v", err))
		response.WriteError(http.StatusInternalServerError, w, "error encoding response")
		return
	}
}
