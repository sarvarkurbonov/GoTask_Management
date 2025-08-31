package api

import (
	"net/http"
	"testing"
	"time"

	"GoTask_Management/internal/models"
)

// TestAPIIntegration tests the complete API workflow
func TestAPIIntegration(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.GetMockService().Reset()

	t.Run("complete task management workflow", func(t *testing.T) {
		// 1. Start with empty task list
		req := helper.CreateRequest("GET", "/api/v1/tasks", nil)
		rr := helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusOK)

		var initialTasks []*models.Task
		helper.AssertJSONResponse(rr, &initialTasks)
		if len(initialTasks) != 0 {
			t.Errorf("Expected empty task list, got %d tasks", len(initialTasks))
		}

		// 2. Create multiple tasks
		taskData := []struct {
			title   string
			dueDate *time.Time
		}{
			{"Task 1", nil},
			{"Task 2", timePtr(time.Now().Add(24 * time.Hour))},
			{"Task 3", timePtr(time.Now().Add(48 * time.Hour))},
		}

		createdTaskIDs := make([]string, 0, len(taskData))

		for _, data := range taskData {
			createReq := TaskRequest{
				Title:   data.title,
				Done:    false,
				DueDate: data.dueDate,
			}

			req := helper.CreateRequest("POST", "/api/v1/tasks", createReq)
			rr := helper.ExecuteRequest(req)
			helper.AssertStatusCode(rr, http.StatusCreated)

			var createdTask models.Task
			helper.AssertJSONResponse(rr, &createdTask)

			if createdTask.Title != data.title {
				t.Errorf("Expected title %s, got %s", data.title, createdTask.Title)
			}

			createdTaskIDs = append(createdTaskIDs, createdTask.ID)
		}

		// 3. Verify all tasks were created
		req = helper.CreateRequest("GET", "/api/v1/tasks", nil)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusOK)

		var allTasks []*models.Task
		helper.AssertJSONResponse(rr, &allTasks)
		if len(allTasks) != len(taskData) {
			t.Errorf("Expected %d tasks, got %d", len(taskData), len(allTasks))
		}

		// 4. Update one task to be completed
		updateReq := TaskRequest{
			Title: "Updated Task 1",
			Done:  true,
		}

		req = helper.CreateRequest("PUT", "/api/v1/tasks/"+createdTaskIDs[0], updateReq)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusOK)

		var updatedTask models.Task
		helper.AssertJSONResponse(rr, &updatedTask)
		if !updatedTask.Done {
			t.Error("Expected task to be marked as done")
		}

		// 5. Filter completed tasks
		req = helper.CreateRequest("GET", "/api/v1/tasks?status=done", nil)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusOK)

		var doneTasks []*models.Task
		helper.AssertJSONResponse(rr, &doneTasks)
		if len(doneTasks) != 1 {
			t.Errorf("Expected 1 done task, got %d", len(doneTasks))
		}

		// 6. Filter uncompleted tasks
		req = helper.CreateRequest("GET", "/api/v1/tasks?status=undone", nil)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusOK)

		var undoneTasks []*models.Task
		helper.AssertJSONResponse(rr, &undoneTasks)
		if len(undoneTasks) != 2 {
			t.Errorf("Expected 2 undone tasks, got %d", len(undoneTasks))
		}

		// 7. Get due tasks
		req = helper.CreateRequest("GET", "/api/v1/tasks/due?days=3", nil)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusOK)

		var dueTasks []*models.Task
		helper.AssertJSONResponse(rr, &dueTasks)
		// Should include tasks with due dates within 3 days
		if len(dueTasks) != 2 {
			t.Errorf("Expected 2 due tasks, got %d", len(dueTasks))
		}

		// 8. Delete a task
		req = helper.CreateRequest("DELETE", "/api/v1/tasks/"+createdTaskIDs[1], nil)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusOK)

		// 9. Verify task was deleted
		req = helper.CreateRequest("GET", "/api/v1/tasks/"+createdTaskIDs[1], nil)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusNotFound)

		// 10. Verify remaining tasks
		req = helper.CreateRequest("GET", "/api/v1/tasks", nil)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusOK)

		var finalTasks []*models.Task
		helper.AssertJSONResponse(rr, &finalTasks)
		if len(finalTasks) != 2 {
			t.Errorf("Expected 2 remaining tasks, got %d", len(finalTasks))
		}
	})
}

func TestAPIErrorScenarios(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.GetMockService().Reset()

	t.Run("handles various error scenarios", func(t *testing.T) {
		// 1. Try to get non-existent task
		req := helper.CreateRequest("GET", "/api/v1/tasks/non-existent", nil)
		rr := helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusNotFound)

		// 2. Try to update non-existent task
		updateReq := TaskRequest{
			Title: "Updated Task",
			Done:  true,
		}
		req = helper.CreateRequest("PUT", "/api/v1/tasks/non-existent", updateReq)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusInternalServerError)

		// 3. Try to delete non-existent task
		req = helper.CreateRequest("DELETE", "/api/v1/tasks/non-existent", nil)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusNotFound)

		// 4. Try to create task with invalid data
		invalidReq := TaskRequest{
			Title: "", // Empty title
			Done:  false,
		}
		req = helper.CreateRequest("POST", "/api/v1/tasks", invalidReq)
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusBadRequest)

		// 5. Try to send malformed JSON
		req = helper.CreateRequest("POST", "/api/v1/tasks", "invalid json")
		rr = helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusBadRequest)
	})
}

func TestAPIContentNegotiation(t *testing.T) {
	helper := NewTestHelper(t)

	t.Run("handles content types correctly", func(t *testing.T) {
		// 1. All responses should be JSON
		endpoints := []string{
			"/health",
			"/api/v1/tasks",
			"/api/v1/tasks/due",
		}

		for _, endpoint := range endpoints {
			req := helper.CreateRequest("GET", endpoint, nil)
			rr := helper.ExecuteRequest(req)

			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json' for %s, got '%s'",
					endpoint, contentType)
			}
		}

		// 2. POST requests should accept JSON
		taskReq := TaskRequest{
			Title: "Content Type Test",
			Done:  false,
		}

		req := helper.CreateRequest("POST", "/api/v1/tasks", taskReq)
		rr := helper.ExecuteRequest(req)
		helper.AssertStatusCode(rr, http.StatusCreated)

		// Response should also be JSON
		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected JSON response for POST, got '%s'", contentType)
		}
	})
}

func TestAPIPerformance(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.GetMockService().Reset()

	t.Run("handles multiple concurrent requests", func(t *testing.T) {
		// Create some initial tasks
		for i := 0; i < 10; i++ {
			taskReq := TaskRequest{
				Title: "Concurrent Test Task",
				Done:  false,
			}

			req := helper.CreateRequest("POST", "/api/v1/tasks", taskReq)
			rr := helper.ExecuteRequest(req)
			helper.AssertStatusCode(rr, http.StatusCreated)
		}

		// Test concurrent reads
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()

				req := helper.CreateRequest("GET", "/api/v1/tasks", nil)
				rr := helper.ExecuteRequest(req)

				if rr.Code != http.StatusOK {
					t.Errorf("Concurrent request failed with status %d", rr.Code)
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestAPIValidation(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.GetMockService().Reset()

	t.Run("validates input data correctly", func(t *testing.T) {
		testCases := []struct {
			name           string
			request        interface{}
			expectedStatus int
		}{
			{
				name: "valid task",
				request: TaskRequest{
					Title: "Valid Task",
					Done:  false,
				},
				expectedStatus: http.StatusCreated,
			},
			{
				name: "empty title",
				request: TaskRequest{
					Title: "",
					Done:  false,
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				name: "whitespace only title",
				request: TaskRequest{
					Title: "   ",
					Done:  false,
				},
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := helper.CreateRequest("POST", "/api/v1/tasks", tc.request)
				rr := helper.ExecuteRequest(req)
				helper.AssertStatusCode(rr, tc.expectedStatus)
			})
		}
	})
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
