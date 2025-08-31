package api

import (
	"net/http"
	"testing"
	"time"

	"GoTask_Management/internal/models"
)

func TestHandleGetTasks(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.GetMockService().Reset()

	// Seed with test data
	task1 := helper.CreateSampleTask("task1", "Task 1")
	task2 := &models.Task{
		ID:        "task2",
		Title:     "Task 2",
		Done:      true,
		CreatedAt: time.Now(),
		DueDate:   nil,
	}
	task3 := helper.CreateSampleTask("task3", "Task 3")

	helper.GetMockService().AddTask(task1)
	helper.GetMockService().AddTask(task2)
	helper.GetMockService().AddTask(task3)

	t.Run("gets all tasks", func(t *testing.T) {
		req := helper.CreateRequest("GET", "/api/v1/tasks", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)
		helper.AssertContentType(rr, "application/json")

		var responseTasks []*models.Task
		helper.AssertJSONResponse(rr, &responseTasks)

		if len(responseTasks) != 3 {
			t.Errorf("Expected 3 tasks, got %d", len(responseTasks))
		}
	})

	t.Run("filters done tasks", func(t *testing.T) {
		req := helper.CreateRequest("GET", "/api/v1/tasks?status=done", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		var responseTasks []*models.Task
		helper.AssertJSONResponse(rr, &responseTasks)

		if len(responseTasks) != 1 {
			t.Errorf("Expected 1 done task, got %d", len(responseTasks))
		}

		if !responseTasks[0].Done {
			t.Error("Expected task to be done")
		}
	})

	t.Run("filters undone tasks", func(t *testing.T) {
		req := helper.CreateRequest("GET", "/api/v1/tasks?status=undone", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		var responseTasks []*models.Task
		helper.AssertJSONResponse(rr, &responseTasks)

		if len(responseTasks) != 2 {
			t.Errorf("Expected 2 undone tasks, got %d", len(responseTasks))
		}

		for _, task := range responseTasks {
			if task.Done {
				t.Error("Expected task to be undone")
			}
		}
	})

	t.Run("handles service error", func(t *testing.T) {
		helper.GetMockService().SetError(true, "service error")

		req := helper.CreateRequest("GET", "/api/v1/tasks", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusInternalServerError)
		helper.AssertErrorResponse(rr, "service error")

		// Reset error state
		helper.GetMockService().SetError(false, "")
	})
}

func TestHandleCreateTask(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.GetMockService().Reset()

	t.Run("creates task successfully", func(t *testing.T) {
		dueDate := time.Now().Add(24 * time.Hour)
		taskReq := TaskRequest{
			Title:   "New Task",
			Done:    false,
			DueDate: &dueDate,
		}

		req := helper.CreateRequest("POST", "/api/v1/tasks", taskReq)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusCreated)
		helper.AssertContentType(rr, "application/json")

		var responseTask models.Task
		helper.AssertJSONResponse(rr, &responseTask)

		if responseTask.Title != "New Task" {
			t.Errorf("Expected title 'New Task', got '%s'", responseTask.Title)
		}

		if responseTask.Done {
			t.Error("Expected new task to not be done")
		}

		if responseTask.ID == "" {
			t.Error("Expected task to have an ID")
		}
	})

	t.Run("creates task without due date", func(t *testing.T) {
		taskReq := TaskRequest{
			Title: "Task without due date",
			Done:  false,
		}

		req := helper.CreateRequest("POST", "/api/v1/tasks", taskReq)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusCreated)

		var responseTask models.Task
		helper.AssertJSONResponse(rr, &responseTask)

		if responseTask.DueDate != nil {
			t.Error("Expected due date to be nil")
		}
	})

	t.Run("fails with empty title", func(t *testing.T) {
		taskReq := TaskRequest{
			Title: "",
			Done:  false,
		}

		req := helper.CreateRequest("POST", "/api/v1/tasks", taskReq)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusBadRequest)
		helper.AssertErrorResponse(rr, "Title is required")
	})

	t.Run("fails with invalid JSON", func(t *testing.T) {
		req := helper.CreateRequest("POST", "/api/v1/tasks", "invalid json")
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusBadRequest)
		helper.AssertErrorResponse(rr, "Invalid request body")
	})

	t.Run("handles service error", func(t *testing.T) {
		helper.GetMockService().SetError(true, "service error")

		taskReq := TaskRequest{
			Title: "Test Task",
			Done:  false,
		}

		req := helper.CreateRequest("POST", "/api/v1/tasks", taskReq)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusInternalServerError)
		helper.AssertErrorResponse(rr, "service error")

		// Reset error state
		helper.GetMockService().SetError(false, "")
	})
}

func TestHandleGetTask(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.GetMockService().Reset()

	// Seed with test data
	task := helper.CreateSampleTask("test_task", "Test Task")
	helper.GetMockService().AddTask(task)

	t.Run("gets existing task", func(t *testing.T) {
		req := helper.CreateRequest("GET", "/api/v1/tasks/test_task", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)
		helper.AssertContentType(rr, "application/json")

		var responseTask models.Task
		helper.AssertJSONResponse(rr, &responseTask)

		if responseTask.ID != task.ID {
			t.Errorf("Expected ID %s, got %s", task.ID, responseTask.ID)
		}
		if responseTask.Title != task.Title {
			t.Errorf("Expected title %s, got %s", task.Title, responseTask.Title)
		}
	})

	t.Run("fails for non-existent task", func(t *testing.T) {
		req := helper.CreateRequest("GET", "/api/v1/tasks/non_existent", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusNotFound)
		helper.AssertErrorResponse(rr, "Task not found")
	})

	t.Run("handles service error", func(t *testing.T) {
		helper.GetMockService().SetError(true, "service error")

		req := helper.CreateRequest("GET", "/api/v1/tasks/test_task", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusNotFound)
		helper.AssertErrorResponse(rr, "Task not found")

		// Reset error state
		helper.GetMockService().SetError(false, "")
	})
}

func TestHandleUpdateTask(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.GetMockService().Reset()

	// Seed with test data
	task := helper.CreateSampleTask("update_task", "Original Task")
	helper.GetMockService().AddTask(task)

	t.Run("updates task successfully", func(t *testing.T) {
		dueDate := time.Now().Add(48 * time.Hour)
		updateReq := TaskRequest{
			Title:   "Updated Task",
			Done:    true,
			DueDate: &dueDate,
		}

		req := helper.CreateRequest("PUT", "/api/v1/tasks/update_task", updateReq)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		var responseTask models.Task
		helper.AssertJSONResponse(rr, &responseTask)

		if responseTask.Title != "Updated Task" {
			t.Errorf("Expected title 'Updated Task', got '%s'", responseTask.Title)
		}

		if !responseTask.Done {
			t.Error("Expected task to be done")
		}
	})

	t.Run("fails with invalid JSON", func(t *testing.T) {
		req := helper.CreateRequest("PUT", "/api/v1/tasks/update_task", "invalid json")
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusBadRequest)
		helper.AssertErrorResponse(rr, "Invalid request body")
	})

	t.Run("handles service error", func(t *testing.T) {
		helper.GetMockService().SetError(true, "service error")

		updateReq := TaskRequest{
			Title: "Updated Task",
			Done:  true,
		}

		req := helper.CreateRequest("PUT", "/api/v1/tasks/update_task", updateReq)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusInternalServerError)
		helper.AssertErrorResponse(rr, "service error")

		// Reset error state
		helper.GetMockService().SetError(false, "")
	})
}

func TestHandleDeleteTask(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.GetMockService().Reset()

	// Seed with test data
	task := helper.CreateSampleTask("delete_task", "Delete Task")
	helper.GetMockService().AddTask(task)

	t.Run("deletes task successfully", func(t *testing.T) {
		req := helper.CreateRequest("DELETE", "/api/v1/tasks/delete_task", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		var response map[string]string
		helper.AssertJSONResponse(rr, &response)

		if response["message"] != "Task deleted successfully" {
			t.Errorf("Expected success message, got '%s'", response["message"])
		}
	})

	t.Run("fails for non-existent task", func(t *testing.T) {
		req := helper.CreateRequest("DELETE", "/api/v1/tasks/non_existent", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusNotFound)
		helper.AssertErrorResponse(rr, "Task not found")
	})

	t.Run("handles service error", func(t *testing.T) {
		helper.GetMockService().SetError(true, "service error")

		req := helper.CreateRequest("DELETE", "/api/v1/tasks/any_task", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusNotFound)
		helper.AssertErrorResponse(rr, "Task not found")

		// Reset error state
		helper.GetMockService().SetError(false, "")
	})
}

func TestHandleGetDueTasks(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.GetMockService().Reset()

	// Seed with test data
	now := time.Now()
	task1 := helper.CreateSampleTaskWithDueDate("due_soon", "Due Soon", now.Add(12*time.Hour))
	task2 := helper.CreateSampleTaskWithDueDate("due_later", "Due Later", now.Add(168*time.Hour)) // 7 days
	task3 := helper.CreateSampleTask("no_due", "No Due Date")

	helper.GetMockService().AddTask(task1)
	helper.GetMockService().AddTask(task2)
	helper.GetMockService().AddTask(task3)

	t.Run("gets due tasks with default days", func(t *testing.T) {
		req := helper.CreateRequest("GET", "/api/v1/tasks/due", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		var responseTasks []*models.Task
		helper.AssertJSONResponse(rr, &responseTasks)

		// Should include both tasks with due dates (within 7 days)
		if len(responseTasks) != 2 {
			t.Errorf("Expected 2 due tasks, got %d", len(responseTasks))
		}
	})

	t.Run("gets due tasks with custom days", func(t *testing.T) {
		req := helper.CreateRequest("GET", "/api/v1/tasks/due?days=1", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		var responseTasks []*models.Task
		helper.AssertJSONResponse(rr, &responseTasks)

		// Should include only the task due soon (within 1 day)
		if len(responseTasks) != 1 {
			t.Errorf("Expected 1 due task, got %d", len(responseTasks))
		}
	})

	t.Run("handles invalid days parameter", func(t *testing.T) {
		req := helper.CreateRequest("GET", "/api/v1/tasks/due?days=invalid", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		var responseTasks []*models.Task
		helper.AssertJSONResponse(rr, &responseTasks)

		// Should use default value (7 days)
		if len(responseTasks) != 2 {
			t.Errorf("Expected 2 due tasks with default days, got %d", len(responseTasks))
		}
	})

	t.Run("handles service error", func(t *testing.T) {
		helper.GetMockService().SetError(true, "service error")

		req := helper.CreateRequest("GET", "/api/v1/tasks/due", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusInternalServerError)
		helper.AssertErrorResponse(rr, "service error")

		// Reset error state
		helper.GetMockService().SetError(false, "")
	})
}

func TestHandleHealth(t *testing.T) {
	helper := NewTestHelper(t)

	t.Run("returns health status", func(t *testing.T) {
		req := helper.CreateRequest("GET", "/health", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)
		helper.AssertContentType(rr, "application/json")

		var response map[string]string
		helper.AssertJSONResponse(rr, &response)

		if response["status"] != "healthy" {
			t.Errorf("Expected status 'healthy', got '%s'", response["status"])
		}
	})
}
