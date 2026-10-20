package handler

import (
	"net/http"

	"mias-backend/internal/middleware"
	"mias-backend/internal/service"
)

type SSEHandler struct {
	hub *service.SSEHub
}

func NewSSEHandler(hub *service.SSEHub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

func (h *SSEHandler) StreamStaffEvents(w http.ResponseWriter, r *http.Request) {
	role, ok := middleware.GetUserRole(r.Context())
	if !ok || (role != "waiter" && role != "cashier" && role != "admin") {
		writeError(w, http.StatusForbidden, "No autorizado para eventos de personal")
		return
	}

	h.hub.ServeHTTP(w, r, role)
}

func (h *SSEHandler) StreamOrderEvents(w http.ResponseWriter, r *http.Request) {
	role, _ := middleware.GetUserRole(r.Context())
	if role == "" {
		role = "customer"
	}
	h.hub.ServeHTTP(w, r, role)
}
