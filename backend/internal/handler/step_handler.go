package handler

import (
	"encoding/json"
	"net/http"

	"autodidact-backend/internal/model"
	"autodidact-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type StepHandler struct {
	svc *service.SyllabusService
}

func NewStepHandler(svc *service.SyllabusService) *StepHandler {
	return &StepHandler{svc: svc}
}

func (h *StepHandler) Create(w http.ResponseWriter, r *http.Request) {
	moduleID := chi.URLParam(r, "moduleId")
	var req model.CreateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	st, err := h.svc.CreateStep(moduleID, req)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, st)
}

func (h *StepHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	st, err := h.svc.GetStepByID(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if st == nil {
		writeJSONError(w, "step not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (h *StepHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateStep(id, req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "step updated successfully"})
}

func (h *StepHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteStep(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "step deleted successfully"})
}

func (h *StepHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateStepProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateStepProgress(id, req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "progress updated successfully"})
}
