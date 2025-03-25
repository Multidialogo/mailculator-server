package API

type QueueCreationAPI struct {
	Data []struct {
		ID                    string            `json:"id" validate:"required,uuid"`
		Type                  string            `json:"type" validate:"required,eq=email"`
		From                  string            `json:"from" validate:"required,email"`
		ReplyTo               string            `json:"reply_to" validate:"required,email"`
		To                    string            `json:"to" validate:"required,email"`
		Subject               string            `json:"subject" validate:"required"`
		BodyHTML              string            `json:"body_html" validate:"required_without=BodyText"`
		BodyText              string            `json:"body_text" validate:"required_without=BodyHTML"`
		Attachments           []string          `json:"attachments" validate:"dive,uri"`
		CustomHeaders         map[string]string `json:"custom_headers"`
		CallbackCallOnSuccess string            `json:"callback_on_success"`
		CallbackCallOnFailure string            `json:"callback_on_failure"`
	} `json:"data" validate:"gt=0,dive,required"`
}
