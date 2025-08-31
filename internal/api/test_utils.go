package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"GoTask_Management/internal/models"
)

// MockTaskService implements TaskService interface for testing
type MockTaskService struct {
	tasks       map[string]*models.Task
	shouldError bool
	errorMsg    string
	idCounter   int
}

// NewMockTaskService creates a new mock task service
func NewMockTaskService() *MockTaskService {
	return &MockTaskService{
		tasks: make(map[string]*models.Task),
	}
}

// SetError configures the mock to return errors
func (m *MockTaskService) SetError(shouldError bool, errorMsg string) {
	m.shouldError = shouldError
	m.errorMsg = errorMsg
}

// Reset clears all tasks and error state
func (m *MockTaskService) Reset() {
	m.tasks = make(map[string]*models.Task)
	m.shouldError = false
	m.errorMsg = ""
	m.idCounter = 0
}

// AddTask adds a task to the mock storage
func (m *MockTaskService) AddTask(task *models.Task) {
	m.tasks[task.ID] = task
}

// CreateTask implements TaskService interface
func (m *MockTaskService) CreateTask(title string, dueDate *time.Time) (*models.Task, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("task title cannot be empty")
	}
	
	m.idCounter++
	task := &models.Task{
		ID:        m.generateID(),
		Title:     title,
		Done:      false,
		CreatedAt: time.Now(),
		DueDate:   dueDate,
	}
	
	m.tasks[task.ID] = task
	return task, nil
}

// ListTasks implements TaskService interface
func (m *MockTaskService) ListTasks(status string) ([]*models.Task, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	
	tasks := make([]*models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		if status == "" || 
		   (status == "done" && task.Done) || 
		   (status == "undone" && !task.Done) {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

// GetTask implements TaskService interface
func (m *MockTaskService) GetTask(id string) (*models.Task, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	
	task, exists := m.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}
	return task, nil
}

// UpdateTask implements TaskService interface
func (m *MockTaskService) UpdateTask(id string, title string, done bool, dueDate *time.Time) (*models.Task, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	
	task, exists := m.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}
	
	if title != "" {
		task.Title = title
	}
	task.Done = done
	task.DueDate = dueDate
	
	return task, nil
}

// DeleteTask implements TaskService interface
func (m *MockTaskService) DeleteTask(id string) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	
	if _, exists := m.tasks[id]; !exists {
		return errors.New("task not found")
	}
	
	delete(m.tasks, id)
	return nil
}

// GetDueTasks implements TaskService interface
func (m *MockTaskService) GetDueTasks(days int) ([]*models.Task, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	
	now := time.Now()
	deadline := now.AddDate(0, 0, days)
	
	dueTasks := make([]*models.Task, 0)
	for _, task := range m.tasks {
		if task.DueDate != nil && !task.DueDate.After(deadline) {
			dueTasks = append(dueTasks, task)
		}
	}
	
	return dueTasks, nil
}

// GetTasksSummary implements TaskService interface
func (m *MockTaskService) GetTasksSummary() (int, int, int, error) {
	if m.shouldError {
		return 0, 0, 0, errors.New(m.errorMsg)
	}
	
	total := len(m.tasks)
	done := 0
	overdue := 0
	now := time.Now()
	
	for _, task := range m.tasks {
		if task.Done {
			done++
		}
		if task.DueDate != nil && task.DueDate.Before(now) && !task.Done {
			overdue++
		}
	}
	
	return total, done, overdue, nil
}

// TestHelper provides utilities for API testing
type TestHelper struct {
	t           *testing.T
	mockService *MockTaskService
	server      *Server
}

// NewTestHelper creates a new API test helper
func NewTestHelper(t *testing.T) *TestHelper {
	mockService := NewMockTaskService()
	server := NewServer(mockService, 8080)

	return &TestHelper{
		t:           t,
		mockService: mockService,
		server:      server,
	}
}

// GetMockService returns the mock task service
func (h *TestHelper) GetMockService() *MockTaskService {
	return h.mockService
}

// CreateRequest creates an HTTP request for testing
func (h *TestHelper) CreateRequest(method, url string, body interface{}) *http.Request {
	var reqBody io.Reader
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			h.t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}
	
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		h.t.Fatalf("Failed to create request: %v", err)
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	return req
}

// ExecuteRequest executes an HTTP request and returns the response
func (h *TestHelper) ExecuteRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h.server.router.ServeHTTP(rr, req)
	return rr
}

// AssertStatusCode checks the HTTP status code
func (h *TestHelper) AssertStatusCode(rr *httptest.ResponseRecorder, expectedCode int) {
	if rr.Code != expectedCode {
		h.t.Errorf("Expected status code %d, got %d. Response body: %s", 
			expectedCode, rr.Code, rr.Body.String())
	}
}

// AssertContentType checks the Content-Type header
func (h *TestHelper) AssertContentType(rr *httptest.ResponseRecorder, expectedType string) {
	contentType := rr.Header().Get("Content-Type")
	if contentType != expectedType {
		h.t.Errorf("Expected Content-Type %s, got %s", expectedType, contentType)
	}
}

// AssertJSONResponse checks that the response is valid JSON and unmarshals it
func (h *TestHelper) AssertJSONResponse(rr *httptest.ResponseRecorder, target interface{}) {
	if err := json.Unmarshal(rr.Body.Bytes(), target); err != nil {
		h.t.Errorf("Failed to unmarshal JSON response: %v", err)
		h.t.Errorf("Response body: %s", rr.Body.String())
	}
}

// AssertErrorResponse checks that the response contains an error message
func (h *TestHelper) AssertErrorResponse(rr *httptest.ResponseRecorder, expectedMessage string) {
	var errorResp ErrorResponse
	h.AssertJSONResponse(rr, &errorResp)
	
	if errorResp.Error != expectedMessage {
		h.t.Errorf("Expected error message '%s', got '%s'", expectedMessage, errorResp.Error)
	}
}

// CreateSampleTask creates a sample task for testing
func (h *TestHelper) CreateSampleTask(id, title string) *models.Task {
	now := time.Now()
	return &models.Task{
		ID:        id,
		Title:     title,
		Done:      false,
		CreatedAt: now,
		DueDate:   nil,
	}
}

// CreateSampleTaskWithDueDate creates a sample task with due date
func (h *TestHelper) CreateSampleTaskWithDueDate(id, title string, dueDate time.Time) *models.Task {
	now := time.Now()
	return &models.Task{
		ID:        id,
		Title:     title,
		Done:      false,
		CreatedAt: now,
		DueDate:   &dueDate,
	}
}

// generateID generates a unique ID for the mock service
func (m *MockTaskService) generateID() string {
	return fmt.Sprintf("mock_task_%d", m.idCounter)
}

// Verify that MockTaskService implements TaskService interface
var _ TaskService = (*MockTaskService)(nil)
