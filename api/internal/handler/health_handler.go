package handler

import "net/http"

// HealthHandler returns 200 {"status":"ok"} — registered in main.go at /healthz.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
