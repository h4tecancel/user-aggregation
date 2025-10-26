package server

import (
	"context"
	"errors"
	"net/http"
	"time"
	"user-aggregation/internal/server/handlers"
	"user-aggregation/internal/server/handlers/swagger"

	"github.com/gorilla/mux"
)

type Server struct {
	httpHandlers *handlers.HTTP
}

func New(h *handlers.HTTP) *Server {
	return &Server{httpHandlers: h}
}

func (s *Server) Start(
	ctx context.Context,
	address string,
	idleTimeout, rwTimeout,
	shutdownTimeout time.Duration) error {
	r := mux.NewRouter()

	r.Methods(http.MethodPost).Path("/users").HandlerFunc(s.httpHandlers.LoadNewInfo)
	r.Methods(http.MethodGet).Path("/users/{id}").HandlerFunc(s.httpHandlers.GetInfo)
	r.Methods(http.MethodPatch).Path("/users/{id}").HandlerFunc(s.httpHandlers.PatchUserInfo)
	r.Methods(http.MethodGet).Path("/users").HandlerFunc(s.httpHandlers.GetAllInfo)
	r.Methods(http.MethodDelete).Path("/users/{id}").HandlerFunc(s.httpHandlers.DeleteInfo)

	r.Methods(http.MethodGet).Path("/summary").HandlerFunc(s.httpHandlers.GetFilterSummary)

	r.Methods(http.MethodGet).Path("/health").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	swagger.RegisterRoutes(r)
	srv := &http.Server{
		Addr:              address,
		Handler:           r,
		IdleTimeout:       idleTimeout,
		ReadTimeout:       rwTimeout,
		WriteTimeout:      rwTimeout,
		ReadHeaderTimeout: rwTimeout * 2,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	shCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		_ = srv.Close()
		return err
	}

	return <-errCh
}
