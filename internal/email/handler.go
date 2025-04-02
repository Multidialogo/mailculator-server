package email

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"multicarrier-email-api/internal/eml"
	"multicarrier-email-api/internal/response"
)

type createEmailRequestBody struct {
	Data []struct {
		Id            string            `json:"id" validate:"required,uuid"`
		From          string            `json:"from" validate:"required,email"`
		ReplyTo       string            `json:"reply_to" validate:"required,email"`
		To            string            `json:"to" validate:"required,email"`
		Subject       string            `json:"subject" validate:"required"`
		BodyHTML      string            `json:"body_html" validate:"required_without=BodyText"`
		BodyText      string            `json:"body_text" validate:"required_without=BodyHTML"`
		Attachments   []string          `json:"attachments" validate:"dive,uri"`
		CustomHeaders map[string]string `json:"custom_headers"`
	} `json:"data" validate:"gt=0,dive,required"`
}

type serviceInterface interface {
	Save(ctx context.Context, in []eml.EML) error
}

type CreateEmailHandler struct {
	attachmentsBasePath string
	emailService        serviceInterface
}

func NewCreateEmailHandler(attachmentsBasePath string, emailService serviceInterface) *CreateEmailHandler {
	return &CreateEmailHandler{
		attachmentsBasePath: attachmentsBasePath,
		emailService:        emailService,
	}
}

func (h *CreateEmailHandler) emlDataSliceFromBody(rb createEmailRequestBody) []eml.EML {
	emails := make([]eml.EML, len(rb.Data))
	for i, e := range rb.Data {
		for i, attachment := range e.Attachments {
			e.Attachments[i] = filepath.Join(h.attachmentsBasePath, attachment)
		}

		emails[i] = eml.EML{
			MessageId:     e.Id,
			From:          e.From,
			ReplyTo:       e.ReplyTo,
			To:            e.To,
			Subject:       e.Subject,
			BodyHTML:      e.BodyHTML,
			BodyText:      e.BodyText,
			Date:          time.Now(),
			Attachments:   e.Attachments,
			CustomHeaders: e.CustomHeaders,
		}
	}

	return emails
}

func (h *CreateEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var requestBody createEmailRequestBody
	if err := decoder.Decode(&requestBody); err != nil {
		response.WriteError(http.StatusBadRequest, w, fmt.Sprintf("error unmarshalling request body: %v", err))
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(requestBody); err != nil {
		response.WriteError(http.StatusBadRequest, w, fmt.Sprintf("error validating request body: %v", err))
		return
	}

	emlDataSlice := h.emlDataSliceFromBody(requestBody)

	if err := h.emailService.Save(context.TODO(), emlDataSlice); err != nil {
		slog.Error(fmt.Sprintf("error saving emails: %v", err))
		response.WriteError(http.StatusConflict, w, "error saving emails")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("{}"))
}
