package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"GoTask_Management/internal/models"
)

// TestHelper provides utilities for storage testing
type TestHelper struct {
	t       *testing.T
	tempDir string
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	tempDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	return &TestHelper{
		t:       t,
		tempDir: tempDir,
	}
}

// Cleanup removes temporary files and directories
func (h *TestHelper) Cleanup() {
	if err := os.RemoveAll(h.tempDir); err != nil {
		h.t.Errorf("Failed to cleanup temp directory: %v", err)
	}
}

// TempFilePath returns a temporary file path
func (h *TestHelper) TempFilePath(filename string) string {
	return filepath.Join(h.tempDir, filename)
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

// Helper functions for generating test data
func generateTestID(index int) string {
	return "test_task_" + string(rune('0'+index))
}

func generateTestTitle(index int) string {
	return "Test Task " + string(rune('0'+index))
}

// CreateInvalidJSONFile creates a file with invalid JSON content
func (h *TestHelper) CreateInvalidJSONFile(filename string) string {
	path := h.TempFilePath(filename)
	content := `{"invalid": json content}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		h.t.Fatalf("Failed to create invalid JSON file: %v", err)
	}
	return path
}

// CreateReadOnlyFile creates a read-only file
func (h *TestHelper) CreateReadOnlyFile(filename string) string {
	path := h.TempFilePath(filename)
	if err := os.WriteFile(path, []byte("[]"), 0444); err != nil {
		h.t.Fatalf("Failed to create read-only file: %v", err)
	}
	return path
}

// CreateNonExistentPath returns a path that doesn't exist
func (h *TestHelper) CreateNonExistentPath(filename string) string {
	return filepath.Join(h.tempDir, "nonexistent", filename)
}
