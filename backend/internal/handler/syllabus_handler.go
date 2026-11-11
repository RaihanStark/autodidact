package handler

import (
	"encoding/json"
	"net/http"

	"autodidact-backend/internal/model"
	"autodidact-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type SyllabusHandler struct {
	svc *service.SyllabusService
}

func NewSyllabusHandler(svc *service.SyllabusService) *SyllabusHandler {
	return &SyllabusHandler{svc: svc}
}

func (h *SyllabusHandler) List(w http.ResponseWriter, r *http.Request) {
	syllabi, err := h.svc.ListSyllabi()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, syllabi)
}

func (h *SyllabusHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	syl, err := h.svc.GetSyllabusByID(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if syl == nil {
		writeJSONError(w, "syllabus not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, syl)
}

func (h *SyllabusHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSyllabusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	syl, err := h.svc.CreateSyllabus(req)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, syl)
}

func (h *SyllabusHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateSyllabusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	syl, err := h.svc.UpdateSyllabus(id, req)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, syl)
}

func (h *SyllabusHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteSyllabus(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "syllabus deleted successfully"})
}

func (h *SyllabusHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	dup, err := h.svc.DuplicateSyllabus(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, dup)
}

func (h *SyllabusHandler) GetActiveStep(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	active, err := h.svc.GetActiveStep(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, active)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
