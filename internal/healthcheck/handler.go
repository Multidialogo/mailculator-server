package healthcheck

import "net/http"

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
