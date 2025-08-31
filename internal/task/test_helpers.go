package task

import (
	"errors"
	"testing"
	"time"

	"GoTask_Management/internal/models"
	"GoTask_Management/internal/storage"
)

// MockStorage implements the storage.Storage interface for testing
type MockStorage struct {
	tasks       map[string]*models.Task
	shouldError bool
	errorMsg    string
}

// NewMockStorage creates a new mock storage
func NewMockStorage() *MockStorage {
	return &MockStorage{
		tasks: make(map[string]*models.Task),
	}
}

// SetError configures the mock to return errors
func (m *MockStorage) SetError(shouldError bool, errorMsg string) {
	m.shouldError = shouldError
	m.errorMsg = errorMsg
}

// Create implements storage.Storage
func (m *MockStorage) Create(task *models.Task) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	m.tasks[task.ID] = task
	return nil
}

// GetAll implements storage.Storage
func (m *MockStorage) GetAll() ([]*models.Task, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	
	tasks := make([]*models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// GetByID implements storage.Storage
func (m *MockStorage) GetByID(id string) (*models.Task, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	
	task, exists := m.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}
	return task, nil
}

// Update implements storage.Storage
func (m *MockStorage) Update(task *models.Task) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	
	if _, exists := m.tasks[task.ID]; !exists {
		return errors.New("task not found")
	}
	m.tasks[task.ID] = task
	return nil
}

// Delete implements storage.Storage
func (m *MockStorage) Delete(id string) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	
	if _, exists := m.tasks[id]; !exists {
		return errors.New("task not found")
	}
	delete(m.tasks, id)
	return nil
}

// Close implements storage.Storage
func (m *MockStorage) Close() error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	return nil
}

// TestHelper provides utilities for task service testing
type TestHelper struct {
	t           *testing.T
	mockStorage *MockStorage
	service     *Service
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	mockStorage := NewMockStorage()
	service := NewService(mockStorage)
	
	return &TestHelper{
		t:           t,
		mockStorage: mockStorage,
		service:     service,
	}
}

// GetService returns the task service
func (h *TestHelper) GetService() *Service {
	return h.service
}

// GetMockStorage returns the mock storage
func (h *TestHelper) GetMockStorage() *MockStorage {
	return h.mockStorage
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

// CreateCompletedTask creates a completed task
func (h *TestHelper) CreateCompletedTask(id, title string) *models.Task {
	now := time.Now()
	return &models.Task{
		ID:        id,
		Title:     title,
		Done:      true,
		CreatedAt: now,
		DueDate:   nil,
	}
}

// CreateMultipleTasks creates multiple sample tasks
func (h *TestHelper) CreateMultipleTasks(count int) []*models.Task {
	tasks := make([]*models.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = h.CreateSampleTask(
			generateTestID(i),
			generateTestTitle(i),
		)
	}
	return tasks
}

// SeedMockStorage adds tasks to the mock storage
func (h *TestHelper) SeedMockStorage(tasks []*models.Task) {
	for _, task := range tasks {
		h.mockStorage.tasks[task.ID] = task
	}
}

// AssertTaskEqual compares two tasks for equality
func (h *TestHelper) AssertTaskEqual(expected, actual *models.Task) {
	if expected.ID != actual.ID {
		h.t.Errorf("Expected ID %s, got %s", expected.ID, actual.ID)
	}
	if expected.Title != actual.Title {
		h.t.Errorf("Expected Title %s, got %s", expected.Title, actual.Title)
	}
	if expected.Done != actual.Done {
		h.t.Errorf("Expected Done %v, got %v", expected.Done, actual.Done)
	}
	if !expected.CreatedAt.Equal(actual.CreatedAt) {
		h.t.Errorf("Expected CreatedAt %v, got %v", expected.CreatedAt, actual.CreatedAt)
	}
	
	// Handle nil due dates
	if expected.DueDate == nil && actual.DueDate != nil {
		h.t.Errorf("Expected DueDate to be nil, got %v", actual.DueDate)
	}
	if expected.DueDate != nil && actual.DueDate == nil {
		h.t.Errorf("Expected DueDate %v, got nil", expected.DueDate)
	}
	if expected.DueDate != nil && actual.DueDate != nil && !expected.DueDate.Equal(*actual.DueDate) {
		h.t.Errorf("Expected DueDate %v, got %v", expected.DueDate, actual.DueDate)
	}
}

// AssertTaskSliceEqual compares two task slices for equality
func (h *TestHelper) AssertTaskSliceEqual(expected, actual []*models.Task) {
	if len(expected) != len(actual) {
		h.t.Errorf("Expected %d tasks, got %d", len(expected), len(actual))
		return
	}

	for i, expectedTask := range expected {
		h.AssertTaskEqual(expectedTask, actual[i])
	}
}

// AssertError checks if an error occurred when expected
func (h *TestHelper) AssertError(err error, shouldError bool, message string) {
	if shouldError && err == nil {
		h.t.Errorf("Expected error for %s, but got none", message)
	}
	if !shouldError && err != nil {
		h.t.Errorf("Unexpected error for %s: %v", message, err)
	}
}

// AssertNoError checks that no error occurred
func (h *TestHelper) AssertNoError(err error, message string) {
	if err != nil {
		h.t.Errorf("Unexpected error for %s: %v", message, err)
	}
}

// AssertContainsTask checks if a task slice contains a specific task
func (h *TestHelper) AssertContainsTask(tasks []*models.Task, expectedTask *models.Task) {
	for _, task := range tasks {
		if task.ID == expectedTask.ID {
			h.AssertTaskEqual(expectedTask, task)
			return
		}
	}
	h.t.Errorf("Task %s not found in task slice", expectedTask.ID)
}

// AssertNotContainsTask checks if a task slice does not contain a specific task
func (h *TestHelper) AssertNotContainsTask(tasks []*models.Task, taskID string) {
	for _, task := range tasks {
		if task.ID == taskID {
			h.t.Errorf("Task %s should not be in task slice", taskID)
			return
		}
	}
}

// Helper functions for generating test data
func generateTestID(index int) string {
	return "test_task_" + string(rune('0'+index))
}

func generateTestTitle(index int) string {
	return "Test Task " + string(rune('0'+index))
}

// CreateOverdueTask creates a task that is overdue
func (h *TestHelper) CreateOverdueTask(id, title string) *models.Task {
	now := time.Now()
	overdueDate := now.Add(-24 * time.Hour) // 1 day ago
	return &models.Task{
		ID:        id,
		Title:     title,
		Done:      false,
		CreatedAt: now.Add(-48 * time.Hour), // Created 2 days ago
		DueDate:   &overdueDate,
	}
}

// CreateDueSoonTask creates a task that is due soon
func (h *TestHelper) CreateDueSoonTask(id, title string, hoursFromNow int) *models.Task {
	now := time.Now()
	dueDate := now.Add(time.Duration(hoursFromNow) * time.Hour)
	return &models.Task{
		ID:        id,
		Title:     title,
		Done:      false,
		CreatedAt: now,
		DueDate:   &dueDate,
	}
}

// Verify that MockStorage implements storage.Storage interface
var _ storage.Storage = (*MockStorage)(nil)
