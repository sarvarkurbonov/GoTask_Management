package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type TaskRequest struct {
	Title   string     `json:"title"`
	Done    bool       `json:"done"`
	DueDate *time.Time `json:"due_date,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	tasks, err := s.taskService.ListTasks(status)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, tasks)
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	task, err := s.taskService.CreateTask(req.Title, req.DueDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, task)
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	task, err := s.taskService.GetTask(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Task not found")
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	task, err := s.taskService.UpdateTask(id, req.Title, req.Done, req.DueDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := s.taskService.DeleteTask(id); err != nil {
		respondWithError(w, http.StatusNotFound, "Task not found")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Task deleted successfully"})
}

func (s *Server) handleGetDueTasks(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 7 // default

	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	tasks, err := s.taskService.GetDueTasks(days)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, tasks)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": startTime.UTC(),
		"service":   "gotask-api",
		"version":   "1.0.0", // This could be injected from build
	}

	// Check storage health if available
	if healthChecker, ok := s.taskService.(interface{ HealthCheck() error }); ok {
		if err := healthChecker.HealthCheck(); err != nil {
			response["status"] = "unhealthy"
			response["storage"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			response["response_time_ms"] = time.Since(startTime).Milliseconds()
			respondWithJSON(w, http.StatusServiceUnavailable, response)
			return
		}
		response["storage"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Add response time
	response["response_time_ms"] = time.Since(startTime).Milliseconds()

	// Add uptime if available (this would need to be tracked globally)
	response["uptime"] = "unknown" // Placeholder for actual uptime tracking

	respondWithJSON(w, http.StatusOK, response)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		return
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}
