package handler

import (
	"encoding/json"
	"net/http"

	"autodidact-backend/internal/model"
	"autodidact-backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type FocusHandler struct {
	svc *service.FocusService
}

func NewFocusHandler(svc *service.FocusService) *FocusHandler {
	return &FocusHandler{svc: svc}
}

func (h *FocusHandler) List(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	items, err := h.svc.ListFocusItems(r.Context(), period)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *FocusHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateFocusItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	item, err := h.svc.CreateFocusItem(r.Context(), &req)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *FocusHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateFocusItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	item, err := h.svc.UpdateFocusItem(r.Context(), id, &req)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *FocusHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteFocusItem(r.Context(), id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *FocusHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	var req model.ReorderFocusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.ReorderFocusItems(r.Context(), req.IDs); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "reordered"})
}

func (h *FocusHandler) ListBacklogItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListBacklogItems(r.Context())
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, items)
}
