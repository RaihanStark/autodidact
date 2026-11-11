package handler

import (
	"encoding/json"
	"net/http"

	"autodidact-backend/internal/model"
	"autodidact-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type ModuleHandler struct {
	svc *service.SyllabusService
}

func NewModuleHandler(svc *service.SyllabusService) *ModuleHandler {
	return &ModuleHandler{svc: svc}
}

func (h *ModuleHandler) Create(w http.ResponseWriter, r *http.Request) {
	syllabusID := chi.URLParam(r, "syllabusId")
	var req model.CreateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	mod, err := h.svc.CreateModule(syllabusID, req)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, mod)
}

func (h *ModuleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateModule(id, req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "module updated successfully"})
}

func (h *ModuleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteModule(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "module deleted successfully"})
}
