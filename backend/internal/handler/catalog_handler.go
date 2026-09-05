package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"mias-backend/internal/service"
)

type CatalogHandler struct {
	catalogService *service.CatalogService
}

func NewCatalogHandler(catalogService *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService}
}

func (h *CatalogHandler) GetPublicMenu(w http.ResponseWriter, r *http.Request) {
	menu, err := h.catalogService.GetPublicMenu(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Error al cargar el menú"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menu)
}

func (h *CatalogHandler) ToggleStock(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"ID de producto inválido"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		InStock bool `json:"in_stock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de petición inválido"}`, http.StatusBadRequest)
		return
	}

	if err := h.catalogService.ToggleStock(r.Context(), int32(id), req.InStock); err != nil {
		http.Error(w, `{"error":"Error actualizando disponibilidad"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Disponibilidad actualizada exitosamente"}`))
}
