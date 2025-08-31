package storage

import (
	"testing"
	"time"

	"GoTask_Management/internal/models"
)

// testStorageCompliance runs a comprehensive test suite that all storage implementations should pass
func testStorageCompliance(t *testing.T, storage Storage) {
	t.Run("Create", func(t *testing.T) {
		task := &models.Task{
			ID:        "test_create",
			Title:     "Test Create Task",
			Done:      false,
			CreatedAt: time.Now(),
			DueDate:   nil,
		}

		err := storage.Create(task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		// Verify task was created
		retrievedTask, err := storage.GetByID("test_create")
		if err != nil {
			t.Fatalf("Failed to retrieve created task: %v", err)
		}

		if retrievedTask.ID != task.ID {
			t.Errorf("Expected ID %s, got %s", task.ID, retrievedTask.ID)
		}
		if retrievedTask.Title != task.Title {
			t.Errorf("Expected title %s, got %s", task.Title, retrievedTask.Title)
		}
		if retrievedTask.Done != task.Done {
			t.Errorf("Expected done %v, got %v", task.Done, retrievedTask.Done)
		}
	})

	t.Run("CreateWithDueDate", func(t *testing.T) {
		dueDate := time.Now().Add(24 * time.Hour)
		task := &models.Task{
			ID:        "test_create_due",
			Title:     "Test Create Task with Due Date",
			Done:      false,
			CreatedAt: time.Now(),
			DueDate:   &dueDate,
		}

		err := storage.Create(task)
		if err != nil {
			t.Fatalf("Failed to create task with due date: %v", err)
		}

		// Verify task was created with due date
		retrievedTask, err := storage.GetByID("test_create_due")
		if err != nil {
			t.Fatalf("Failed to retrieve created task: %v", err)
		}

		if retrievedTask.DueDate == nil {
			t.Error("Expected due date to be preserved")
		} else {
			// Allow small time differences due to storage precision
			diff := retrievedTask.DueDate.Sub(dueDate)
			if diff < -time.Second || diff > time.Second {
				t.Errorf("Due date not preserved correctly: expected %v, got %v", dueDate, *retrievedTask.DueDate)
			}
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		// Create multiple tasks
		tasks := []*models.Task{
			{
				ID:        "test_getall_1",
				Title:     "Task 1",
				Done:      false,
				CreatedAt: time.Now(),
			},
			{
				ID:        "test_getall_2",
				Title:     "Task 2",
				Done:      true,
				CreatedAt: time.Now(),
			},
			{
				ID:        "test_getall_3",
				Title:     "Task 3",
				Done:      false,
				CreatedAt: time.Now(),
			},
		}

		for _, task := range tasks {
			err := storage.Create(task)
			if err != nil {
				t.Fatalf("Failed to create task %s: %v", task.ID, err)
			}
		}

		// Get all tasks
		allTasks, err := storage.GetAll()
		if err != nil {
			t.Fatalf("Failed to get all tasks: %v", err)
		}

		// Should have at least the tasks we created (might have more from other tests)
		if len(allTasks) < len(tasks) {
			t.Errorf("Expected at least %d tasks, got %d", len(tasks), len(allTasks))
		}

		// Verify our tasks are in the result
		taskMap := make(map[string]*models.Task)
		for _, task := range allTasks {
			taskMap[task.ID] = task
		}

		for _, expectedTask := range tasks {
			if retrievedTask, exists := taskMap[expectedTask.ID]; !exists {
				t.Errorf("Task %s not found in GetAll result", expectedTask.ID)
			} else {
				if retrievedTask.Title != expectedTask.Title {
					t.Errorf("Task %s title mismatch: expected %s, got %s", 
						expectedTask.ID, expectedTask.Title, retrievedTask.Title)
				}
			}
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		task := &models.Task{
			ID:        "test_getbyid",
			Title:     "Test GetByID Task",
			Done:      true,
			CreatedAt: time.Now(),
		}

		err := storage.Create(task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		// Test successful retrieval
		retrievedTask, err := storage.GetByID("test_getbyid")
		if err != nil {
			t.Fatalf("Failed to get task by ID: %v", err)
		}

		if retrievedTask.ID != task.ID {
			t.Errorf("Expected ID %s, got %s", task.ID, retrievedTask.ID)
		}
		if retrievedTask.Title != task.Title {
			t.Errorf("Expected title %s, got %s", task.Title, retrievedTask.Title)
		}
		if retrievedTask.Done != task.Done {
			t.Errorf("Expected done %v, got %v", task.Done, retrievedTask.Done)
		}

		// Test non-existent task
		_, err = storage.GetByID("non_existent_task")
		if err == nil {
			t.Error("Expected error for non-existent task, got nil")
		}
	})

	t.Run("Update", func(t *testing.T) {
		// Create initial task
		task := &models.Task{
			ID:        "test_update",
			Title:     "Original Title",
			Done:      false,
			CreatedAt: time.Now(),
		}

		err := storage.Create(task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		// Update the task
		dueDate := time.Now().Add(48 * time.Hour)
		task.Title = "Updated Title"
		task.Done = true
		task.DueDate = &dueDate

		err = storage.Update(task)
		if err != nil {
			t.Fatalf("Failed to update task: %v", err)
		}

		// Verify update
		retrievedTask, err := storage.GetByID("test_update")
		if err != nil {
			t.Fatalf("Failed to retrieve updated task: %v", err)
		}

		if retrievedTask.Title != "Updated Title" {
			t.Errorf("Expected updated title 'Updated Title', got %s", retrievedTask.Title)
		}
		if !retrievedTask.Done {
			t.Error("Expected task to be done after update")
		}
		if retrievedTask.DueDate == nil {
			t.Error("Expected due date to be set after update")
		}

		// Test updating non-existent task
		nonExistentTask := &models.Task{
			ID:        "non_existent_update",
			Title:     "Non-existent",
			Done:      false,
			CreatedAt: time.Now(),
		}

		err = storage.Update(nonExistentTask)
		if err == nil {
			t.Error("Expected error for updating non-existent task, got nil")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		// Create task to delete
		task := &models.Task{
			ID:        "test_delete",
			Title:     "Task to Delete",
			Done:      false,
			CreatedAt: time.Now(),
		}

		err := storage.Create(task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		// Verify task exists
		_, err = storage.GetByID("test_delete")
		if err != nil {
			t.Fatalf("Task should exist before deletion: %v", err)
		}

		// Delete the task
		err = storage.Delete("test_delete")
		if err != nil {
			t.Fatalf("Failed to delete task: %v", err)
		}

		// Verify task is deleted
		_, err = storage.GetByID("test_delete")
		if err == nil {
			t.Error("Expected error for deleted task, got nil")
		}

		// Test deleting non-existent task
		err = storage.Delete("non_existent_delete")
		if err == nil {
			t.Error("Expected error for deleting non-existent task, got nil")
		}
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		// Test with special characters, Unicode, emojis
		task := &models.Task{
			ID:        "test_special_chars",
			Title:     "Special chars: áéíóú ñ 中文 🚀 \"quotes\" 'apostrophes' & symbols",
			Done:      false,
			CreatedAt: time.Now(),
		}

		err := storage.Create(task)
		if err != nil {
			t.Fatalf("Failed to create task with special characters: %v", err)
		}

		retrievedTask, err := storage.GetByID("test_special_chars")
		if err != nil {
			t.Fatalf("Failed to retrieve task with special characters: %v", err)
		}

		if retrievedTask.Title != task.Title {
			t.Errorf("Special characters not preserved: expected %s, got %s", 
				task.Title, retrievedTask.Title)
		}
	})

	t.Run("EmptyTitle", func(t *testing.T) {
		// Test with empty title (should be allowed at storage level)
		task := &models.Task{
			ID:        "test_empty_title",
			Title:     "",
			Done:      false,
			CreatedAt: time.Now(),
		}

		err := storage.Create(task)
		if err != nil {
			t.Fatalf("Failed to create task with empty title: %v", err)
		}

		retrievedTask, err := storage.GetByID("test_empty_title")
		if err != nil {
			t.Fatalf("Failed to retrieve task with empty title: %v", err)
		}

		if retrievedTask.Title != "" {
			t.Errorf("Expected empty title, got %s", retrievedTask.Title)
		}
	})

	t.Run("TimePreservation", func(t *testing.T) {
		// Test that timestamps are preserved accurately
		now := time.Now().Truncate(time.Millisecond) // Truncate to millisecond precision
		dueDate := now.Add(24 * time.Hour)

		task := &models.Task{
			ID:        "test_time_preservation",
			Title:     "Time Preservation Test",
			Done:      false,
			CreatedAt: now,
			DueDate:   &dueDate,
		}

		err := storage.Create(task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		retrievedTask, err := storage.GetByID("test_time_preservation")
		if err != nil {
			t.Fatalf("Failed to retrieve task: %v", err)
		}

		// Allow small differences due to storage precision
		createdAtDiff := retrievedTask.CreatedAt.Sub(now)
		if createdAtDiff < -time.Second || createdAtDiff > time.Second {
			t.Errorf("CreatedAt not preserved: expected %v, got %v", now, retrievedTask.CreatedAt)
		}

		if retrievedTask.DueDate == nil {
			t.Error("DueDate should not be nil")
		} else {
			dueDateDiff := retrievedTask.DueDate.Sub(dueDate)
			if dueDateDiff < -time.Second || dueDateDiff > time.Second {
				t.Errorf("DueDate not preserved: expected %v, got %v", dueDate, *retrievedTask.DueDate)
			}
		}
	})
}
