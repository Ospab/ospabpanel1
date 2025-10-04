package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	coreServer "ospab-panel/internal/core/server"
	"ospab-panel/internal/core/user"
	"ospab-panel/internal/hypervisor"
	"ospab-panel/pkg/auth"
)

type Handler struct {
	userService   *user.Service
	serverService *coreServer.Service
	hvFactory     *hypervisor.HypervisorFactory
	jwtManager    *auth.JWTManager
}

func NewHandler(userService *user.Service, serverService *coreServer.Service, hvFactory *hypervisor.HypervisorFactory, jwtManager *auth.JWTManager) *Handler {
	return &Handler{
		userService:   userService,
		serverService: serverService,
		hvFactory:     hvFactory,
		jwtManager:    jwtManager,
	}
}

type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Version string `json:"version"`
}

type VersionResponse struct {
	Version     string `json:"version"`
	BuildDate   string `json:"build_date"`
	Environment string `json:"environment"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Middleware для JWT авторизации
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			h.sendError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			h.sendError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		claims, err := h.jwtManager.ValidateToken(bearerToken[1])
		if err != nil {
			h.sendError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Добавляем информацию о пользователе в заголовки
		r.Header.Set("X-User-ID", strconv.Itoa(claims.UserID))
		r.Header.Set("X-Username", claims.Username)

		next(w, r)
	}
}

// POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var loginReq user.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}
	if loginReq.Username == "" || loginReq.Password == "" {
		h.sendError(w, http.StatusBadRequest, "Требуются логин и пароль")
		return
	}
	u, err := h.userService.GetUserByUsername(loginReq.Username)
	if err != nil || !h.userService.ValidatePassword(u, loginReq.Password) {
		h.sendError(w, http.StatusUnauthorized, "Неверный логин или пароль")
		return
	}
	token, err := h.jwtManager.GenerateToken(u.ID, u.Username)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}
	h.sendJSON(w, http.StatusOK, user.LoginResponse{Token: token, User: *u})
}

// POST /api/auth/register
type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		h.sendError(w, http.StatusBadRequest, "All fields required")
		return
	}
	u, err := h.userService.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		msg := "Ошибка регистрации"
		switch {
		case errors.Is(err, user.ErrPasswordTooShort):
			msg = "Пароль слишком короткий (минимум 8 символов)"
		case errors.Is(err, user.ErrUsernameTaken):
			msg = "Логин уже используется"
		case errors.Is(err, user.ErrEmailTaken):
			msg = "Email уже используется"
		case errors.Is(err, user.ErrDuplicateValue):
			msg = "Значение уже существует"
		default:
			msg = err.Error()
		}
		h.sendError(w, http.StatusBadRequest, msg)
		return
	}
	token, err := h.jwtManager.GenerateToken(u.ID, u.Username)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed token")
		return
	}
	h.sendJSON(w, http.StatusCreated, user.LoginResponse{Token: token, User: *u})
}

// GET /api/status - требует авторизации
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	response := StatusResponse{
		Status:  "ok",
		Message: "ospab panel 0.1-alpha working",
		Version: "0.1-alpha",
	}
	h.sendJSON(w, http.StatusOK, response)
}

// GET /api/version - требует авторизации
func (h *Handler) Version(w http.ResponseWriter, r *http.Request) {
	response := VersionResponse{
		Version:     "0.1-alpha",
		BuildDate:   "2025-09-30",
		Environment: "development",
	}
	h.sendJSON(w, http.StatusOK, response)
}

// Вспомогательные функции
func (h *Handler) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}

// --- Hypervisor endpoints ---
type HypervisorType struct {
	Code   string   `json:"code"`
	Name   string   `json:"name"`
	Params []string `json:"params"`
}

type HypervisorCheckRequest struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateConnectionRequest struct {
	Host     *string `json:"host,omitempty"`
	Port     *int    `json:"port,omitempty"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

func (h *Handler) ListHypervisors(w http.ResponseWriter, r *http.Request) {
	types := []HypervisorType{
		{Code: "prx", Name: "Proxmox", Params: []string{"host", "port", "username", "password"}},
		{Code: "vmv", Name: "VMware ESXi", Params: []string{"host", "port", "username", "password"}},
		{Code: "hyv", Name: "Hyper-V", Params: []string{"host", "port", "username", "password"}},
		{Code: "kvm", Name: "KVM/QEMU", Params: []string{"host", "port", "username", "password"}},
		{Code: "xen", Name: "XenServer", Params: []string{"host", "port", "username", "password"}},
	}
	h.sendJSON(w, http.StatusOK, types)
}

func (h *Handler) CheckHypervisorConnection(w http.ResponseWriter, r *http.Request) {
	var req HypervisorCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	client, err := h.hvFactory.CreateClient(req.Type)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Unsupported type")
		return
	}
	ctx := r.Context()
	err = client.Connect(ctx, &hypervisor.Server{Host: req.Host, Port: req.Port, Type: req.Type, Username: req.Username, Password: req.Password})
	if err != nil {
		h.sendError(w, http.StatusBadGateway, "Connect failed: "+err.Error())
		return
	}
	err = client.TestConnection(ctx)
	if err != nil {
		h.sendError(w, http.StatusBadGateway, "Test failed: "+err.Error())
		return
	}
	h.sendJSON(w, http.StatusOK, map[string]string{"result": "ok"})
}

func (h *Handler) GetServerConnection(w http.ResponseWriter, r *http.Request) {
	uid := atoi(r.Header.Get("X-User-ID"))
	vars := mux.Vars(r)
	sid, _ := strconv.Atoi(vars["id"])
	srv, err := h.serverService.GetServerByID(sid, uid)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Server not found")
		return
	}
	resp := map[string]interface{}{
		"host":     srv.Host,
		"port":     srv.Port,
		"type":     srv.Type,
		"username": srv.UsernameDecrypted,
		// пароль не отдаём
	}
	h.sendJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpdateServerConnection(w http.ResponseWriter, r *http.Request) {
	uid := atoi(r.Header.Get("X-User-ID"))
	vars := mux.Vars(r)
	sid, _ := strconv.Atoi(vars["id"])
	var req UpdateConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	up := coreServer.UpdateServerRequest{}
	if req.Host != nil {
		up.Host = *req.Host
	}
	if req.Port != nil {
		up.Port = *req.Port
	}
	if req.Username != nil {
		up.Username = *req.Username
	}
	if req.Password != nil {
		up.Password = *req.Password
	}
	_, err := h.serverService.UpdateServer(sid, uid, &up)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendJSON(w, http.StatusOK, map[string]string{"result": "ok"})
}

// POST /api/servers/{id}/connection/check
func (h *Handler) CheckServerConnection(w http.ResponseWriter, r *http.Request) {
	uid := atoi(r.Header.Get("X-User-ID"))
	vars := mux.Vars(r)
	sid, _ := strconv.Atoi(vars["id"])
	srv, err := h.serverService.GetServerByID(sid, uid)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Server not found")
		return
	}
	client, err := h.hvFactory.CreateClient(srv.Type)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Unsupported type")
		return
	}
	ctx := r.Context()
	err = client.Connect(ctx, &hypervisor.Server{Host: srv.Host, Port: srv.Port, Type: srv.Type, Username: srv.UsernameDecrypted, Password: srv.PasswordDecrypted})
	if err != nil {
		h.sendError(w, http.StatusBadGateway, "Connect failed: "+err.Error())
		return
	}
	err = client.TestConnection(ctx)
	if err != nil {
		h.sendError(w, http.StatusBadGateway, "Test failed: "+err.Error())
		return
	}
	h.sendJSON(w, http.StatusOK, map[string]string{"result": "ok"})
}

func atoi(s string) int { i, _ := strconv.Atoi(s); return i }
