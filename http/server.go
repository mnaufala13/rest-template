package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/mnaufala13/rest-template/config"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

var (
	server *http.Server
	router *chi.Mux

	unHealthy bool
	mut       sync.RWMutex
)

func Start(cfg config.ServerConfig, dep ServerDependency) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	router = chi.NewRouter()
	router.Use(recoverPanic)
	router.Use(requestID)
	router.Use(jsonResponse)

	router.NotFound(notFoundHandler)
	router.MethodNotAllowed(methodNotAllowedHandler)

	registerHandlers(dep)

	server = &http.Server{
		Addr:         addr,
		IdleTimeout:  time.Duration(cfg.IdleTimeout),
		ReadTimeout:  time.Duration(cfg.ReadTimeout),
		WriteTimeout: time.Duration(cfg.WriteTimeout),
		Handler:      router,
	}

	slog.Info(fmt.Sprintf("starting HTTP on server %s", addr))
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("can't start serving HTTP requests", "error", err)
		}
	}
}

func Stop(delay, timeout time.Duration) {
	slog.Info("stopping HTTP server")
	slog.Info("update health check to return service unavailable (503)")
	setUnhealthy()
	slog.Info(fmt.Sprintf("wait %s before calling http shutdown", delay.String()))
	time.Sleep(delay)
	c, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := server.Shutdown(c)
	if err != nil {
		slog.Error(fmt.Sprintf("failed shutdown http server: %v", err))
	} else {
		slog.Info("finish shutdown HTTP server")
	}
}

func setUnhealthy() {
	mut.Lock()
	unHealthy = true
	mut.Unlock()
}

func isUnhealthy() bool {
	mut.RLock()
	uh := unHealthy
	mut.RUnlock()
	return uh
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if isUnhealthy() {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status": 0}`))
		return
	}
	w.Write([]byte(`{"status": 1}`))
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"status": 0, "error": "request method not allowed"}`))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"status": 0, "error": "requested url not found"}`))
}
