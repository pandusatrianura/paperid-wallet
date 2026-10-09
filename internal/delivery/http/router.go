package http

import "net/http"

// NewRouter configures and returns an http.Handler with all registered routes.
func NewRouter(disburseHandler *DisburseHandler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/disburse", disburseHandler)
	return mux
}
