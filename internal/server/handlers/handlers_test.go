package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"user-aggregation/internal/models"
	"user-aggregation/internal/models/response"
	"user-aggregation/internal/repo"
	"user-aggregation/internal/server/handlers/mocks"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func withVars(r *http.Request, key, val string) *http.Request {
	return mux.SetURLVars(r, map[string]string{key: val})
}

func toJSON(v any) *bytes.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

func TestLoadNewInfo_OK(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	u := models.UserInfo{
		UserID: uuid.New(), ServiceName: "Netflix", Price: 999,
		StartDate: time.Now().UTC().Truncate(time.Second),
		EndDate:   time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second),
	}

	m.On("Insert", mock.Anything, mock.AnythingOfType("*models.UserInfo")).
		Return(nil).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/users", toJSON(u))
	w := httptest.NewRecorder()

	h.LoadNewInfo(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	m.AssertExpectations(t)
}

func TestLoadNewInfo_BadJSON(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(`{"user_id":}`)))
	w := httptest.NewRecorder()

	h.LoadNewInfo(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetInfo_OK(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	uid := uuid.New()
	expected := []models.UserInfo{{UserID: uid, ServiceName: "A", Price: 10}}

	m.On("GetByUserID", mock.Anything, uid).
		Return(expected, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/users/"+uid.String(), nil)
	req = withVars(req, "id", uid.String())
	w := httptest.NewRecorder()

	h.GetInfo(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}

func TestGetInfo_BadUUID(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	req := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
	req = withVars(req, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.GetInfo(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
	m.AssertNotCalled(t, "GetByUserID", mock.Anything, mock.Anything)
}

func TestDeleteInfo_OK(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	uid := uuid.New()
	m.On("DeleteByUserID", mock.Anything, uid).
		Return(int64(2), nil).
		Once()

	req := httptest.NewRequest(http.MethodDelete, "/users/"+uid.String(), nil)
	req = withVars(req, "id", uid.String())
	w := httptest.NewRecorder()

	h.DeleteInfo(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}

func TestDeleteInfo_NotFound(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	uid := uuid.New()
	m.On("DeleteByUserID", mock.Anything, uid).
		Return(int64(0), repo.ErrNotFound).
		Once()

	req := httptest.NewRequest(http.MethodDelete, "/users/"+uid.String(), nil)
	req = withVars(req, "id", uid.String())
	w := httptest.NewRecorder()

	h.DeleteInfo(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
	m.AssertExpectations(t)
}

func TestDeleteInfo_BadUUID(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	req := httptest.NewRequest(http.MethodDelete, "/users/bad", nil)
	req = withVars(req, "id", "bad")
	w := httptest.NewRecorder()

	h.DeleteInfo(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
	m.AssertNotCalled(t, "DeleteByUserID", mock.Anything, mock.Anything)
}

func TestGetAllInfo_OK(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	m.On("List", mock.Anything).
		Return([]models.UserInfo{{ServiceName: "X"}}, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	h.GetAllInfo(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}

func TestPatchUserInfo_EndOnly_OK(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	uid := uuid.New()
	end := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

	m.
		On("UpdateUserInfo", mock.Anything, uid,
			(*int64)(nil), // важно: явное nil нужного типа
			mock.MatchedBy(func(t *time.Time) bool {
				if t == nil {
					return false
				}
				// Нормализуем секунды, чтобы не споткнуться о наносекунды и зону
				got := t.UTC().Truncate(time.Second)
				want := end.UTC().Truncate(time.Second)
				return got.Equal(want)
			}),
		).
		Return(int64(1), nil).
		Once()

	body := models.UpdateUserInfo{EndDate: &end}
	req := httptest.NewRequest(http.MethodPatch, "/users/"+uid.String(), toJSON(body))
	req = withVars(req, "id", uid.String())
	w := httptest.NewRecorder()

	h.PatchUserInfo(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}

func TestPatchUserInfo_PriceOnly_OK(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	uid := uuid.New()
	price := int64(1999)

	m.
		On("UpdateUserInfo", mock.Anything, uid,
			mock.MatchedBy(func(p *int64) bool { return p != nil && *p == price }),
			(*time.Time)(nil),
		).
		Return(int64(1), nil). // не 2, если хендлер ждёт “ровно 1 обновлена”
		Once()

	body := models.UpdateUserInfo{Price: &price}
	req := httptest.NewRequest(http.MethodPatch, "/users/"+uid.String(), toJSON(body))
	req = withVars(req, "id", uid.String())
	w := httptest.NewRecorder()

	h.PatchUserInfo(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}

func TestPatchUserInfo_BothFields_OK(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	uid := uuid.New()
	price := int64(2500)
	end := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)

	m.
		On("UpdateUserInfo", mock.Anything, uid,
			mock.MatchedBy(func(p *int64) bool { return p != nil && *p == price }),
			mock.MatchedBy(func(t *time.Time) bool {
				if t == nil {
					return false
				}
				got := t.UTC().Truncate(time.Second)
				want := end.UTC().Truncate(time.Second)
				return got.Equal(want)
			}),
		).
		Return(int64(1), nil).
		Once()

	body := models.UpdateUserInfo{Price: &price, EndDate: &end}
	req := httptest.NewRequest(http.MethodPatch, "/users/"+uid.String(), toJSON(body))
	req = withVars(req, "id", uid.String())
	w := httptest.NewRecorder()

	h.PatchUserInfo(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	m.AssertExpectations(t)
}

func TestPatchUserInfo_NoFields_BadRequest(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	uid := uuid.New()
	req := httptest.NewRequest(http.MethodPatch, "/users/"+uid.String(), bytes.NewReader([]byte(`{}`)))
	req = withVars(req, "id", uid.String())
	w := httptest.NewRecorder()

	h.PatchUserInfo(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	m.AssertNotCalled(t, "UpdateUserInfo", mock.Anything, mock.Anything, mock.Anything)
}

func TestPatchUserInfo_BadJSON(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	uid := uuid.New()
	req := httptest.NewRequest(http.MethodPatch, "/users/"+uid.String(), bytes.NewReader([]byte(`{"end_date":`)))
	req = withVars(req, "id", uid.String())
	w := httptest.NewRecorder()

	h.PatchUserInfo(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	m.AssertNotCalled(t, "UpdateUserInfo", mock.Anything, mock.Anything, mock.Anything)
}

func TestPatchUserInfo_BadUUID(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	req := httptest.NewRequest(http.MethodPatch, "/users/not-a-uuid", bytes.NewReader([]byte(`{"price": 100}`)))
	req = withVars(req, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.PatchUserInfo(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	m.AssertNotCalled(t, "UpdateUserInfo", mock.Anything, mock.Anything, mock.Anything)
}

func TestGetFilterSummary_OK(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	uid := uuid.New()
	svc := "Netflix"
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(72 * time.Hour)
	total := int64(1500)

	m.
		On("FilterSum", mock.Anything,
			mock.MatchedBy(func(p *uuid.UUID) bool { return p != nil && *p == uid }),
			mock.MatchedBy(func(p *string) bool { return p != nil && *p == svc }),
			mock.MatchedBy(func(p *time.Time) bool { return p != nil && p.Equal(start) }),
			mock.MatchedBy(func(p *time.Time) bool { return p != nil && p.Equal(end) }),
		).
		Return(total, nil).
		Once()

	url := "/summary?user_id=" + uid.String() +
		"&service_name=" + svc +
		"&start_date=" + start.Format(time.RFC3339) +
		"&end_date=" + end.Format(time.RFC3339)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	h.GetFilterSummary(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// при желании — проверим JSON
	var out response.Summary
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	require.Equal(t, total, out.TotalCost)
	m.AssertExpectations(t)
}

func TestGetFilterSummary_BadUserID(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	req := httptest.NewRequest(http.MethodGet, "/summary?user_id=bad", nil)
	w := httptest.NewRecorder()

	h.GetFilterSummary(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
	m.AssertNotCalled(t, "FilterSum", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetFilterSummary_BadDates(t *testing.T) {
	m := new(mocks.RepoMock)
	h := New(slog.Default(), m)

	req := httptest.NewRequest(http.MethodGet, "/summary?start_date=2025-13-40T00:00:00Z", nil)
	w := httptest.NewRecorder()
	h.GetFilterSummary(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/summary?end_date=not-a-date", nil)
	w2 := httptest.NewRecorder()
	h.GetFilterSummary(w2, req2)
	require.Equal(t, http.StatusBadRequest, w2.Code)

	m.AssertNotCalled(t, "FilterSum", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
