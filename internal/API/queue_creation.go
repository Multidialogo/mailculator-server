package API

type QueueCreationAPI struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			From          string            `json:"from"`
			ReplyTo       string            `json:"replyTo"`
			To            string            `json:"to"`
			Subject       string            `json:"subject"`
			BodyHTML      string            `json:"bodyHTML"`
			BodyText      string            `json:"bodyText"`
			Attachments   []string          `json:"attachments"`
			CustomHeaders map[string]string `json:"customHeaders"`
		} `json:"attributes"`
	} `json:"data"`
}
