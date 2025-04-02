package response

import (
	"encoding/json"
	"net/http"
)

type errorMessage struct {
	Error string `json:"error"`
}

func WriteError(status int, w http.ResponseWriter, msg string) {
	body, _ := json.Marshal(errorMessage{Error: msg})
	http.Error(w, string(body), status)
}
