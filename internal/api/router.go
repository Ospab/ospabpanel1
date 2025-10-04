package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (h *Handler) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	// Публичные
	api.HandleFunc("/auth/login", h.Login).Methods(http.MethodPost)
	api.HandleFunc("/auth/register", h.Register).Methods(http.MethodPost)

	// Защищённые
	api.HandleFunc("/status", h.AuthMiddleware(h.Status)).Methods(http.MethodGet)
	api.HandleFunc("/version", h.AuthMiddleware(h.Version)).Methods(http.MethodGet)

	// Серверы (CRUD)
	sh := NewServerHandlers(h.serverService, h.hvFactory)
	api.HandleFunc("/servers", h.AuthMiddleware(sh.GetServers)).Methods(http.MethodGet)
	api.HandleFunc("/servers", h.AuthMiddleware(sh.CreateServer)).Methods(http.MethodPost)
	api.HandleFunc("/servers/{id}", h.AuthMiddleware(sh.GetServer)).Methods(http.MethodGet)
	api.HandleFunc("/servers/{id}", h.AuthMiddleware(sh.UpdateServer)).Methods(http.MethodPut, http.MethodPatch)
	api.HandleFunc("/servers/{id}", h.AuthMiddleware(sh.DeleteServer)).Methods(http.MethodDelete)

	// Инстансы
	api.HandleFunc("/servers/{id}/instances", h.AuthMiddleware(sh.ListInstances)).Methods(http.MethodGet)
	api.HandleFunc("/servers/{id}/instances/{instanceId}/{action}", h.AuthMiddleware(sh.InstanceAction)).Methods(http.MethodPost)

	// Hypervisor endpoints
	api.HandleFunc("/hypervisors", h.AuthMiddleware(h.ListHypervisors)).Methods(http.MethodGet)
	api.HandleFunc("/hypervisors/check", h.AuthMiddleware(h.CheckHypervisorConnection)).Methods(http.MethodPost)
	api.HandleFunc("/servers/{id}/connection", h.AuthMiddleware(h.GetServerConnection)).Methods(http.MethodGet)
	api.HandleFunc("/servers/{id}/connection", h.AuthMiddleware(h.UpdateServerConnection)).Methods(http.MethodPatch)

	// CORS
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	return r
}
