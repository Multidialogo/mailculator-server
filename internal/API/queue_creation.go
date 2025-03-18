package API

type QueueCreationAPI struct {
	Data []struct {
		ID                    string            `json:"id"`
		Type                  string            `json:"type"`
		From                  string            `json:"from"`
		ReplyTo               string            `json:"reply_to"`
		To                    string            `json:"to"`
		Subject               string            `json:"subject"`
		BodyHTML              string            `json:"body_html"`
		BodyText              string            `json:"body_text"`
		Attachments           []string          `json:"attachments"`
		CustomHeaders         map[string]string `json:"custom_headers"`
		CallbackCallOnSuccess string            `json:"callback_on_success"`
		CallbackCallOnFailure string            `json:"callback_on_failure"`
	} `json:"data"`
}
