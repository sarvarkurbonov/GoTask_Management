package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	taskService TaskService
	router      *mux.Router
	httpServer  *http.Server
	port        int
}

func NewServer(taskService TaskService, port int) *Server {
	s := &Server{
		taskService: taskService,
		port:        port,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()

	// Add middleware
	s.router.Use(loggingMiddleware)
	s.router.Use(jsonMiddleware)

	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Task routes - specific routes must come before parameterized routes
	api.HandleFunc("/tasks", s.handleGetTasks).Methods("GET")
	api.HandleFunc("/tasks", s.handleCreateTask).Methods("POST")
	api.HandleFunc("/tasks/due", s.handleGetDueTasks).Methods("GET")
	api.HandleFunc("/tasks/{id}", s.handleGetTask).Methods("GET")
	api.HandleFunc("/tasks/{id}", s.handleUpdateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}", s.handleDeleteTask).Methods("DELETE")

	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	if s.httpServer == nil {
		return nil // Server was never started
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
