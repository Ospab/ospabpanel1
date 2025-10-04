package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"ospab-panel/internal/core/server"
	"ospab-panel/internal/hypervisor"
)

type ServerHandlers struct {
	serverService *server.Service
	hvFactory     *hypervisor.HypervisorFactory
}

func NewServerHandlers(serverService *server.Service, hvFactory *hypervisor.HypervisorFactory) *ServerHandlers {
	return &ServerHandlers{serverService: serverService, hvFactory: hvFactory}
}

func (h *ServerHandlers) GetServers(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromHeader(r)
	list, err := h.serverService.GetServersByUserID(uid)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, http.StatusOK, list)
}

func (h *ServerHandlers) CreateServer(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromHeader(r)
	var req server.CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" || req.Host == "" || req.Port == 0 || req.Type == "" || req.Username == "" || req.Password == "" {
		sendErr(w, http.StatusBadRequest, "missing fields")
		return
	}
	srv, err := h.serverService.CreateServer(&req, uid)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, http.StatusCreated, srv)
}

func (h *ServerHandlers) GetServer(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromHeader(r)
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	srv, err := h.serverService.GetServerByID(id, uid)
	if err != nil {
		sendErr(w, http.StatusNotFound, "not found")
		return
	}
	sendJSON(w, http.StatusOK, srv)
}

func (h *ServerHandlers) UpdateServer(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromHeader(r)
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var req server.UpdateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	srv, err := h.serverService.UpdateServer(id, uid, &req)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSON(w, http.StatusOK, srv)
}

func (h *ServerHandlers) DeleteServer(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromHeader(r)
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.serverService.DeleteServer(id, uid); err != nil {
		sendErr(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Instances (объединённо VM/LXC) ---
func (h *ServerHandlers) ListInstances(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromHeader(r)
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	srv, err := h.serverService.GetServerByID(id, uid)
	if err != nil {
		sendErr(w, http.StatusNotFound, "server not found")
		return
	}
	client, err := h.hvFactory.CreateClient(srv.Type)
	if err != nil {
		sendErr(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	if err := client.Connect(ctx, &hypervisor.Server{ID: srv.ID, Name: srv.Name, Host: srv.Host, Port: srv.Port, Type: srv.Type, Username: srv.UsernameDecrypted, Password: srv.PasswordDecrypted, UserID: srv.UserID, IsActive: srv.IsActive}); err != nil {
		sendErr(w, http.StatusBadGateway, "connect failed")
		return
	}
	instances, err := client.GetInstances(ctx)
	if err != nil {
		sendErr(w, http.StatusBadGateway, err.Error())
		return
	}
	sendJSON(w, http.StatusOK, instances)
}

func (h *ServerHandlers) InstanceAction(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromHeader(r)
	vars := mux.Vars(r)
	sid, _ := strconv.Atoi(vars["id"])
	action := vars["action"]
	instID := vars["instanceId"]
	srv, err := h.serverService.GetServerByID(sid, uid)
	if err != nil {
		sendErr(w, http.StatusNotFound, "server not found")
		return
	}
	client, err := h.hvFactory.CreateClient(srv.Type)
	if err != nil {
		sendErr(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	if err := client.Connect(ctx, &hypervisor.Server{ID: srv.ID, Name: srv.Name, Host: srv.Host, Port: srv.Port, Type: srv.Type, Username: srv.UsernameDecrypted, Password: srv.PasswordDecrypted, UserID: srv.UserID, IsActive: srv.IsActive}); err != nil {
		sendErr(w, http.StatusBadGateway, "connect failed")
		return
	}
	instType := instanceTypeFromQuery(r)
	var actErr error
	switch action {
	case "start":
		actErr = client.StartInstance(ctx, instType, instID)
	case "stop":
		actErr = client.StopInstance(ctx, instType, instID)
	case "restart":
		actErr = client.RestartInstance(ctx, instType, instID)
	case "status":
		st, err := client.GetInstanceStatus(ctx, instType, instID)
		if err != nil {
			sendErr(w, http.StatusBadGateway, err.Error())
			return
		}
		sendJSON(w, http.StatusOK, map[string]string{"status": st})
		return
	case "config":
		cfg, err := client.GetInstanceConfig(ctx, instType, instID)
		if err != nil {
			sendErr(w, http.StatusBadGateway, err.Error())
			return
		}
		sendJSON(w, http.StatusOK, cfg)
		return
	default:
		sendErr(w, http.StatusBadRequest, "unsupported action")
		return
	}
	if actErr != nil {
		sendErr(w, http.StatusBadGateway, actErr.Error())
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"result": "ok"})
}

// --- helpers ---
func userIDFromHeader(r *http.Request) int {
	v := r.Header.Get("X-User-ID")
	id, _ := strconv.Atoi(v)
	return id
}
func sendJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
func sendErr(w http.ResponseWriter, code int, msg string) {
	sendJSON(w, code, map[string]string{"error": msg})
}
func instanceTypeFromQuery(r *http.Request) string {
	t := r.URL.Query().Get("type")
	if t == "" {
		return "vm"
	}
	t = strings.ToLower(t)
	if t != "vm" && t != "lxc" {
		return "vm"
	}
	return t
}
