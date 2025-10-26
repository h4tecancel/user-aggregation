package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"user-aggregation/internal/models"
	"user-aggregation/internal/models/response"
	"user-aggregation/internal/repo"
	"user-aggregation/internal/transport/http/respond"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type HTTP struct {
	Logger *slog.Logger
	DB     repo.Repo
}

func New(logger *slog.Logger, db repo.Repo) *HTTP {
	return &HTTP{Logger: logger, DB: db}
}

// LoadNewInfo godoc
// @Summary Create new user info
// @Description Create a new user subscription information record
// @Tags users
// @Accept json
// @Produce json
// @Param userInfo body models.UserInfo true "User subscription information"
// @Success 201 {object} models.UserInfo
// @Failure 400 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /users [post]
func (h *HTTP) LoadNewInfo(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.load_new_info"
	ctx := r.Context()

	var userInfo models.UserInfo
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&userInfo); err != nil {
		respond.Error(w, h.Logger, op, http.StatusBadRequest, "invalid JSON body", err)
		return
	}

	if err := h.DB.Insert(ctx, &userInfo); err != nil {
		respond.Error(w, h.Logger, op, http.StatusInternalServerError, "failed to save record", err)
		return
	}

	respond.Writer(w, h.Logger, op, http.StatusCreated, userInfo)
}

// GetInfo godoc
// @Summary Get user info by ID
// @Description Get all subscription information for a specific user
// @Tags users
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {array} models.UserInfo
// @Failure 400 {object} response.ErrorPayload
// @Failure 404 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /users/{id} [get]
func (h *HTTP) GetInfo(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.get_info"
	ctx := r.Context()

	id, err := parseUUIDVar(r, "id")
	if err != nil {
		respond.Error(w, h.Logger, op, http.StatusBadRequest, "invalid user_id", err)
		return
	}

	users, err := h.DB.GetByUserID(ctx, id)
	if err != nil {
		respond.Error(w, h.Logger, op, http.StatusInternalServerError, "failed to fetch records", err)
		return
	}

	respond.Writer(w, h.Logger, op, http.StatusOK, users)
}

// DeleteInfo godoc
// @Summary Delete user info by ID
// @Description Delete all subscription information for a specific user
// @Tags users
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {integer} int64 "Number of deleted records, but not in json. this will need to be done"
// @Failure 400 {object} response.ErrorPayload
// @Failure 404 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /users/{id} [delete]
func (h *HTTP) DeleteInfo(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.delete_info"
	ctx := r.Context()

	id, err := parseUUIDVar(r, "id")
	if err != nil {
		respond.Error(w, h.Logger, op, http.StatusBadRequest, "invalid user_id", err)
		return
	}

	n, err := h.DB.DeleteByUserID(ctx, id)
	if err != nil {
		if err == repo.ErrNotFound {
			respond.Error(w, h.Logger, op, http.StatusNotFound, "nothing to delete", err)
			return
		}
		respond.Error(w, h.Logger, op, http.StatusInternalServerError, "failed to delete", err)
		return
	}

	respond.Writer(w, h.Logger, op, http.StatusOK, map[string]any{"deleted": n})
}

// GetAllInfo godoc
// @Summary List of all users
// @Description Get all user subscription information records
// @Tags users
// @Produce json
// @Success 200 {array} models.UserInfo
// @Failure 500 {object} response.ErrorPayload
// @Router /users [get]
func (h *HTTP) GetAllInfo(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.get_all_info"
	ctx := r.Context()

	listInfo, err := h.DB.List(ctx)
	if err != nil {
		respond.Error(w, h.Logger, op, http.StatusInternalServerError, "failed to list", err)
		return
	}

	respond.Writer(w, h.Logger, op, http.StatusOK, listInfo)
}

// PatchUserInfo godoc
// @Summary Update user info
// @Description Partially update user subscription information (price and/or end date)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param update body models.UpdateUserInfo true "Update fields (price and/or end_date)"
// @Success 200 {integer} int64 "Number of updated records, but not in json, this will need to be done"
// @Failure 400 {object} response.ErrorPayload
// @Failure 404 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /users/{id} [patch]
func (h *HTTP) PatchUserInfo(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.patch_user_info"
	ctx := r.Context()

	id, err := parseUUIDVar(r, "id")
	if err != nil {
		respond.Error(w, h.Logger, op, http.StatusBadRequest, "invalid user_id", err)
		return
	}
	defer r.Body.Close()

	var patch models.UpdateUserInfo
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&patch); err != nil {
		respond.Error(w, h.Logger, op, http.StatusBadRequest, "invalid JSON body", err)
		return
	}

	if patch.Price == nil && patch.EndDate == nil {
		respond.Error(w, h.Logger, op, http.StatusBadRequest, "no fields to update", nil)
		return
	}

	if patch.EndDate != nil {
		t := patch.EndDate.UTC().Truncate(time.Second)
		patch.EndDate = &t
	}

	ui, err := h.DB.UpdateUserInfo(ctx, id, patch.Price, patch.EndDate)
	if err != nil {
		if err == repo.ErrNotFound {
			respond.Error(w, h.Logger, op, http.StatusNotFound, "nothing to update", err)
			return
		}
		respond.Error(w, h.Logger, op, http.StatusInternalServerError, "failed to update user", err)
		return
	}

	respond.Writer(w, h.Logger, op, http.StatusOK, ui)
}

// GetFilterSummary godoc
// @Summary Get filtered summary
// @Description Get total cost summary with optional filters
// @Tags summary
// @Produce json
// @Param service_name query string false "Filter by service name"
// @Param user_id query string false "Filter by user ID (UUID)"
// @Param start_date query string false "Filter by start date (RFC3339 format)"
// @Param end_date query string false "Filter by end date (RFC3339 format)"
// @Success 200 {object} response.Summary
// @Failure 400 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /summary [get]
func (h *HTTP) GetFilterSummary(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.get_filter_summary"
	ctx := r.Context()

	q := r.URL.Query()

	var serviceName *string
	if s := q.Get("service_name"); s != "" {
		serviceName = &s
	}

	var userID *uuid.UUID
	if uid := q.Get("user_id"); uid != "" {
		if parsed, err := uuid.Parse(uid); err == nil {
			userID = &parsed
		} else {
			respond.Error(w, h.Logger, op, http.StatusBadRequest, "invalid user_id", err)
			return
		}
	}

	var startDate, endDate *time.Time
	if s := q.Get("start_date"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			startDate = &t
		} else {
			respond.Error(w, h.Logger, op, http.StatusBadRequest, "invalid start_date (use RFC3339)", err)
			return
		}
	}
	if s := q.Get("end_date"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			endDate = &t
		} else {
			respond.Error(w, h.Logger, op, http.StatusBadRequest, "invalid end_date (use RFC3339)", err)
			return
		}
	}

	total, err := h.DB.FilterSum(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		respond.Error(w, h.Logger, op, http.StatusInternalServerError, "failed to calculate summary", err)
		return
	}

	out := response.Summary{
		TotalCost: total,
	}
	respond.Writer(w, h.Logger, op, http.StatusOK, out)
}

func parseUUIDVar(r *http.Request, key string) (uuid.UUID, error) {
	idStr := mux.Vars(r)[key]
	return uuid.Parse(idStr)
}
