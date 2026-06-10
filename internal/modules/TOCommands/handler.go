package TOCommands

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	svc IService
}

func NewHandler(svc IService) *Handler {
	return &Handler{svc: svc}
}

// RecordCommand godoc
// @Summary      Ingest a teleop command
// @Tags         commands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateCommandRequest true "Command payload"
// @Success      204
// @Failure      400 {object} map[string]string
// @Router       /api/v1/commands [post]
func (h *Handler) RecordCommand(w http.ResponseWriter, r *http.Request) {
	var req CreateCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.Record(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetHistory godoc
// @Summary      Query command history for a robot
// @Tags         commands
// @Produce      json
// @Security     BearerAuth
// @Param        robot_id query int    true  "Robot ID"
// @Param        from     query string true  "Start time (RFC3339)"
// @Param        to       query string true  "End time (RFC3339)"
// @Success      200 {array}  CommandResponse
// @Failure      400 {object} map[string]string
// @Router       /api/v1/commands [get]
func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	robotIDStr := r.URL.Query().Get("robot_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	robotID, err := strconv.ParseUint(robotIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid robot_id")
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from time, use RFC3339 format")
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to time, use RFC3339 format")
		return
	}

	resp, err := h.svc.GetHistory(r.Context(), uint(robotID), from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
