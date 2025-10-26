package respond

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"user-aggregation/internal/models/response"
)

func Error(w http.ResponseWriter, log *slog.Logger, op string, code int, clientMsg string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if log != nil {
		attrs := []any{
			slog.String("op", op),
			slog.Int("status", code),
			slog.String("client_msg", clientMsg),
		}
		if err != nil {
			attrs = append(attrs, slog.Any("error", err))
		}
		log.Error("request failed", attrs...)
	}

	_ = json.NewEncoder(w).Encode(response.ErrorPayload{
		Error:  clientMsg,
		Op:     op,
		Status: code,
	})
}

func Writer(w http.ResponseWriter, log *slog.Logger, op string, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if log != nil {
		log.Info("response", slog.String("op", op), slog.Int("status", code))
	}
	if err := json.NewEncoder(w).Encode(v); err != nil && log != nil {
		log.Warn("write json failed",
			slog.String("op", op),
			slog.Int("status", code),
			slog.Any("error", err),
		)
	}
}
