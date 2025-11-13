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

type emailDataInput struct {
	Id            string            `json:"id" validate:"required,uuid"`
	From          string            `json:"from" validate:"required,email"`
	ReplyTo       string            `json:"reply_to" validate:"required,email"`
	To            string            `json:"to" validate:"required,email"`
	Subject       string            `json:"subject" validate:"required"`
	BodyHTML      string            `json:"body_html" validate:"required_without=BodyText"`
	BodyText      string            `json:"body_text" validate:"required_without=BodyHTML"`
	Attachments   []string          `json:"attachments" validate:"dive,uri"`
	CustomHeaders map[string]string `json:"custom_headers"`
}

type createEmailRequestBody struct {
	Data []emailDataInput `json:"data" validate:"gt=0,dive,required"`
}

type CreateEmailResult struct {
	ID     string       `json:"id"`
	Status string       `json:"status"`
	Error  *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BatchEmailResponse struct {
	Summary struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"summary"`
	Results []CreateEmailResult `json:"results"`
}

type serviceInterface interface {
	Save(ctx context.Context, emailRequests []EmailRequest) error
}

type CreateEmailHandler struct {
	emailService serviceInterface
}

func NewCreateEmailHandler(emailService serviceInterface) *CreateEmailHandler {
	return &CreateEmailHandler{
		emailService: emailService,
	}
}

func (h *CreateEmailHandler) emailRequestsFromBody(rb createEmailRequestBody) ([]EmailRequest, error) {
	emailRequests := make([]EmailRequest, len(rb.Data))

	for i, e := range rb.Data {
		payloadBytes, err := json.Marshal(e)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal single email payload: %w", err)
		}

		emailRequests[i] = EmailRequest{
			MessageId:    e.Id,
			PayloadBytes: payloadBytes,
		}
	}

	return emailRequests, nil
}

func (h *CreateEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteError(http.StatusBadRequest, w, fmt.Sprintf("error reading request body: %v", err))
		return
	}

	slog.Info(fmt.Sprintf("received body: %v", string(body)))

	buf := bytes.NewBuffer(body)
	decoder := json.NewDecoder(buf)

	var requestBody createEmailRequestBody
	if err := decoder.Decode(&requestBody); err != nil {
		slog.Error(fmt.Sprintf("bad payload: %v", string(body)))
		response.WriteError(http.StatusBadRequest, w, fmt.Sprintf("error unmarshalling request body: %v", err))
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(requestBody); err != nil {
		response.WriteError(http.StatusBadRequest, w, fmt.Sprintf("error validating request body: %v", err))
		return
	}

	emailRequests, err := h.emailRequestsFromBody(requestBody)
	if err != nil {
		slog.Error(fmt.Sprintf("error creating email requests: %v", err))
		response.WriteError(http.StatusInternalServerError, w, "error creating email requests")
		return
	}

	if err := h.emailService.Save(context.TODO(), emailRequests); err != nil {
		slog.Error(fmt.Sprintf("error saving emails: %v", err))
		response.WriteError(http.StatusConflict, w, "error saving emails")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("{}"))
}
