package api

import (
	"net/http"
	"testing"
)

func TestNewServer(t *testing.T) {
	mockService := NewMockTaskService()

	t.Run("creates server with valid parameters", func(t *testing.T) {
		server := NewServer(mockService, 8080)

		if server == nil {
			t.Error("Expected server to be created")
		}

		if server.taskService == nil {
			t.Error("Expected server to have a task service")
		}

		if server.port != 8080 {
			t.Errorf("Expected port 8080, got %d", server.port)
		}

		if server.router == nil {
			t.Error("Expected server to have a router")
		}
	})

	t.Run("creates server with different port", func(t *testing.T) {
		server := NewServer(mockService, 3000)

		if server.port != 3000 {
			t.Errorf("Expected port 3000, got %d", server.port)
		}
	})

	t.Run("creates server with zero port", func(t *testing.T) {
		server := NewServer(mockService, 0)

		if server.port != 0 {
			t.Errorf("Expected port 0, got %d", server.port)
		}
	})
}



func TestServer_Integration(t *testing.T) {
	helper := NewTestHelper(t)

	t.Run("handles complete task workflow", func(t *testing.T) {
		// Create a task
		createReq := TaskRequest{
			Title: "Integration Test Task",
			Done:  false,
		}

		req := helper.CreateRequest("POST", "/api/v1/tasks", createReq)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusCreated)

		var createdTask map[string]interface{}
		helper.AssertJSONResponse(rr, &createdTask)

		taskID := createdTask["id"].(string)

		// Get the task
		req = helper.CreateRequest("GET", "/api/v1/tasks/"+taskID, nil)
		rr = helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		var retrievedTask map[string]interface{}
		helper.AssertJSONResponse(rr, &retrievedTask)

		if retrievedTask["title"] != "Integration Test Task" {
			t.Error("Retrieved task should match created task")
		}

		// Update the task
		updateReq := TaskRequest{
			Title: "Updated Integration Test Task",
			Done:  true,
		}

		req = helper.CreateRequest("PUT", "/api/v1/tasks/"+taskID, updateReq)
		rr = helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		var updatedTask map[string]interface{}
		helper.AssertJSONResponse(rr, &updatedTask)

		if updatedTask["title"] != "Updated Integration Test Task" {
			t.Error("Task should be updated")
		}

		if updatedTask["done"] != true {
			t.Error("Task should be marked as done")
		}

		// List all tasks
		req = helper.CreateRequest("GET", "/api/v1/tasks", nil)
		rr = helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		var allTasks []map[string]interface{}
		helper.AssertJSONResponse(rr, &allTasks)

		if len(allTasks) == 0 {
			t.Error("Should have at least one task")
		}

		// Delete the task
		req = helper.CreateRequest("DELETE", "/api/v1/tasks/"+taskID, nil)
		rr = helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusOK)

		// Verify task is deleted
		req = helper.CreateRequest("GET", "/api/v1/tasks/"+taskID, nil)
		rr = helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusNotFound)
	})
}

func TestServer_ErrorHandling(t *testing.T) {
	helper := NewTestHelper(t)

	t.Run("handles service errors gracefully", func(t *testing.T) {
		// Configure mock to return errors
		helper.GetMockService().SetError(true, "database connection failed")

		// Try to get tasks
		req := helper.CreateRequest("GET", "/api/v1/tasks", nil)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusInternalServerError)
		helper.AssertErrorResponse(rr, "database connection failed")

		// Reset error state
		helper.GetMockService().SetError(false, "")
	})

	t.Run("handles malformed JSON requests", func(t *testing.T) {
		req := helper.CreateRequest("POST", "/api/v1/tasks", "invalid json")
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusBadRequest)
		helper.AssertErrorResponse(rr, "Invalid request body")
	})

	t.Run("handles missing required fields", func(t *testing.T) {
		invalidReq := TaskRequest{
			Title: "", // Empty title should be rejected
			Done:  false,
		}

		req := helper.CreateRequest("POST", "/api/v1/tasks", invalidReq)
		rr := helper.ExecuteRequest(req)

		helper.AssertStatusCode(rr, http.StatusBadRequest)
		helper.AssertErrorResponse(rr, "Title is required")
	})
}

func TestServer_CORS_and_Headers(t *testing.T) {
	helper := NewTestHelper(t)

	t.Run("sets appropriate headers", func(t *testing.T) {
		req := helper.CreateRequest("GET", "/health", nil)
		rr := helper.ExecuteRequest(req)

		// Check that Content-Type is set by JSON middleware
		if rr.Header().Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type header to be set")
		}
	})
}


