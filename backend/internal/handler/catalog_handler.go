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
		writeError(w, http.StatusInternalServerError, "Error al cargar el menú")
		return
	}

	writeJSON(w, http.StatusOK, menu)
}

func (h *CatalogHandler) ToggleStock(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID de producto inválido")
		return
	}

	var req struct {
		InStock bool `json:"in_stock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Cuerpo de petición inválido")
		return
	}

	if err := h.catalogService.ToggleStock(r.Context(), int32(id), req.InStock); err != nil {
		writeError(w, http.StatusInternalServerError, "Error actualizando disponibilidad")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Disponibilidad actualizada exitosamente"})
}
