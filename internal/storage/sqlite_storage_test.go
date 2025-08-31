package storage

import (
	"database/sql"
	"testing"
	"time"

	"GoTask_Management/internal/models"

	_ "modernc.org/sqlite"
)

func TestNewSQLiteStorage(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("creates new storage with new database", func(t *testing.T) {
		dbPath := helper.TempFilePath("test.db")

		storage, err := NewSQLiteStorage(dbPath)
		helper.AssertNoError(err, "creating new SQLite storage")
		defer storage.Close()

		if storage == nil {
			t.Error("Expected storage to be created")
		}

		// Verify table was created by trying to query it
		tasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting all tasks from new storage")
		if len(tasks) != 0 {
			t.Errorf("Expected empty task list, got %d tasks", len(tasks))
		}
	})

	t.Run("opens existing database", func(t *testing.T) {
		dbPath := helper.TempFilePath("existing.db")

		// Create database with sample data
		initialStorage, err := NewSQLiteStorage(dbPath)
		helper.AssertNoError(err, "creating initial storage")

		sampleTask := helper.CreateSampleTask("existing_1", "Existing Task")
		err = initialStorage.Create(sampleTask)
		helper.AssertNoError(err, "creating sample task")
		initialStorage.Close()

		// Open existing database
		storage, err := NewSQLiteStorage(dbPath)
		helper.AssertNoError(err, "opening existing SQLite storage")
		defer storage.Close()

		tasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting tasks from existing storage")

		if len(tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(tasks))
		}

		helper.AssertTaskEqual(sampleTask, tasks[0])
	})

	t.Run("fails with invalid database path", func(t *testing.T) {
		invalidPath := "/invalid/path/test.db"

		_, err := NewSQLiteStorage(invalidPath)
		helper.AssertError(err, true, "creating storage with invalid path")
	})
}

func TestSQLiteStorage_Create(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewSQLiteStorage(helper.TempFilePath("create.db"))
	helper.AssertNoError(err, "creating storage")
	defer storage.Close()

	t.Run("creates single task", func(t *testing.T) {
		task := helper.CreateSampleTask("create_1", "Create Task")

		err = storage.Create(task)
		helper.AssertNoError(err, "creating task")

		// Verify task was created
		tasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting all tasks")

		if len(tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(tasks))
		}

		helper.AssertTaskEqual(task, tasks[0])
	})

	t.Run("creates multiple tasks", func(t *testing.T) {
		tasks := helper.CreateMultipleTasks(3)

		for _, task := range tasks {
			err = storage.Create(task)
			helper.AssertNoError(err, "creating task")
		}

		// Verify all tasks were created (including the one from previous test)
		allTasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting all tasks")

		if len(allTasks) != 4 { // 1 from previous test + 3 new
			t.Errorf("Expected 4 tasks, got %d", len(allTasks))
		}
	})

	t.Run("creates task with due date", func(t *testing.T) {
		dueDate := time.Now().Add(24 * time.Hour)
		task := helper.CreateSampleTaskWithDueDate("due_1", "Task with Due Date", dueDate)

		err = storage.Create(task)
		helper.AssertNoError(err, "creating task with due date")

		// Verify task was created with due date
		retrievedTask, err := storage.GetByID("due_1")
		helper.AssertNoError(err, "getting task by ID")

		helper.AssertTaskEqual(task, retrievedTask)
	})

	t.Run("creates task without due date", func(t *testing.T) {
		task := helper.CreateSampleTask("no_due_1", "Task without Due Date")

		err = storage.Create(task)
		helper.AssertNoError(err, "creating task without due date")

		// Verify task was created without due date
		retrievedTask, err := storage.GetByID("no_due_1")
		helper.AssertNoError(err, "getting task by ID")

		if retrievedTask.DueDate != nil {
			t.Error("Expected DueDate to be nil")
		}
	})
}

func TestSQLiteStorage_GetAll(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewSQLiteStorage(helper.TempFilePath("getall.db"))
	helper.AssertNoError(err, "creating storage")
	defer storage.Close()

	t.Run("returns empty list for new storage", func(t *testing.T) {
		tasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting all tasks")

		if len(tasks) != 0 {
			t.Errorf("Expected 0 tasks, got %d", len(tasks))
		}
	})

	t.Run("returns all tasks", func(t *testing.T) {
		expectedTasks := helper.CreateMultipleTasks(5)

		for _, task := range expectedTasks {
			err = storage.Create(task)
			helper.AssertNoError(err, "creating task")
		}

		tasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting all tasks")

		if len(tasks) != len(expectedTasks) {
			t.Errorf("Expected %d tasks, got %d", len(expectedTasks), len(tasks))
		}

		// Verify all tasks are present (order might be different)
		for _, expectedTask := range expectedTasks {
			found := false
			for _, task := range tasks {
				if task.ID == expectedTask.ID {
					helper.AssertTaskEqual(expectedTask, task)
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected task %s not found", expectedTask.ID)
			}
		}
	})
}

func TestSQLiteStorage_GetByID(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewSQLiteStorage(helper.TempFilePath("getbyid.db"))
	helper.AssertNoError(err, "creating storage")
	defer storage.Close()

	// Create test tasks
	tasks := helper.CreateMultipleTasks(3)
	for _, task := range tasks {
		err = storage.Create(task)
		helper.AssertNoError(err, "creating task")
	}

	t.Run("finds existing task", func(t *testing.T) {
		for _, expectedTask := range tasks {
			task, err := storage.GetByID(expectedTask.ID)
			helper.AssertNoError(err, "getting task by ID")
			helper.AssertTaskEqual(expectedTask, task)
		}
	})

	t.Run("returns error for non-existent task", func(t *testing.T) {
		_, err := storage.GetByID("non_existent")
		helper.AssertError(err, true, "getting non-existent task")

		if err != sql.ErrNoRows {
			t.Errorf("Expected sql.ErrNoRows, got: %v", err)
		}
	})
}

func TestSQLiteStorage_Update(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewSQLiteStorage(helper.TempFilePath("update.db"))
	helper.AssertNoError(err, "creating storage")
	defer storage.Close()

	// Create initial task
	originalTask := helper.CreateSampleTask("update_1", "Original Task")
	err = storage.Create(originalTask)
	helper.AssertNoError(err, "creating original task")

	t.Run("updates existing task", func(t *testing.T) {
		dueDate := time.Now().Add(48 * time.Hour)
		updatedTask := &models.Task{
			ID:        "update_1",
			Title:     "Updated Task",
			Done:      true,
			CreatedAt: originalTask.CreatedAt, // CreatedAt should not change
			DueDate:   &dueDate,
		}

		err = storage.Update(updatedTask)
		helper.AssertNoError(err, "updating task")

		// Verify task was updated
		retrievedTask, err := storage.GetByID("update_1")
		helper.AssertNoError(err, "getting updated task")

		if retrievedTask.Title != "Updated Task" {
			t.Errorf("Expected title 'Updated Task', got '%s'", retrievedTask.Title)
		}
		if !retrievedTask.Done {
			t.Error("Expected task to be done")
		}
		if retrievedTask.DueDate == nil || !retrievedTask.DueDate.Equal(dueDate) {
			t.Errorf("Expected due date %v, got %v", dueDate, retrievedTask.DueDate)
		}
	})

	t.Run("returns error for non-existent task", func(t *testing.T) {
		nonExistentTask := helper.CreateSampleTask("non_existent", "Non-existent Task")

		err = storage.Update(nonExistentTask)
		helper.AssertError(err, true, "updating non-existent task")

		if err != sql.ErrNoRows {
			t.Errorf("Expected sql.ErrNoRows, got: %v", err)
		}
	})
}

func TestSQLiteStorage_Delete(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewSQLiteStorage(helper.TempFilePath("delete.db"))
	helper.AssertNoError(err, "creating storage")
	defer storage.Close()

	// Create test tasks
	tasks := helper.CreateMultipleTasks(3)
	for _, task := range tasks {
		err = storage.Create(task)
		helper.AssertNoError(err, "creating task")
	}

	t.Run("deletes existing task", func(t *testing.T) {
		err = storage.Delete(tasks[1].ID)
		helper.AssertNoError(err, "deleting task")

		// Verify task was deleted
		_, err = storage.GetByID(tasks[1].ID)
		helper.AssertError(err, true, "getting deleted task")

		// Verify other tasks still exist
		remainingTasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting remaining tasks")

		if len(remainingTasks) != 2 {
			t.Errorf("Expected 2 remaining tasks, got %d", len(remainingTasks))
		}
	})

	t.Run("returns error for non-existent task", func(t *testing.T) {
		err = storage.Delete("non_existent")
		helper.AssertError(err, true, "deleting non-existent task")

		if err != sql.ErrNoRows {
			t.Errorf("Expected sql.ErrNoRows, got: %v", err)
		}
	})
}

func TestSQLiteStorage_Close(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewSQLiteStorage(helper.TempFilePath("close.db"))
	helper.AssertNoError(err, "creating storage")

	err = storage.Close()
	helper.AssertNoError(err, "closing storage")

	// Verify that operations fail after close
	task := helper.CreateSampleTask("after_close", "After Close")
	err = storage.Create(task)
	helper.AssertError(err, true, "creating task after close")
}

func TestSQLiteStorage_ConcurrentAccess(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewSQLiteStorage(helper.TempFilePath("concurrent.db"))
	helper.AssertNoError(err, "creating storage")
	defer storage.Close()

	t.Run("concurrent reads", func(t *testing.T) {
		// Create initial tasks
		for i := 0; i < 5; i++ {
			task := helper.CreateSampleTask(generateTestID(i), generateTestTitle(i))
			err = storage.Create(task)
			helper.AssertNoError(err, "creating initial task")
		}

		// Test concurrent reads (these should work fine)
		done := make(chan error, 5)

		for i := 0; i < 5; i++ {
			go func() {
				_, err := storage.GetAll()
				done <- err
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			err := <-done
			if err != nil {
				t.Errorf("Concurrent read error: %v", err)
			}
		}
	})

	t.Run("sequential writes after concurrent reads", func(t *testing.T) {
		// SQLite handles concurrent reads well, but writes need to be more careful
		// Test that we can still write after concurrent reads
		task := helper.CreateSampleTask("after_concurrent", "After Concurrent Reads")
		err := storage.Create(task)
		helper.AssertNoError(err, "creating task after concurrent reads")

		// Verify the task was created
		retrievedTask, err := storage.GetByID("after_concurrent")
		helper.AssertNoError(err, "getting task after concurrent operations")
		helper.AssertTaskEqual(task, retrievedTask)
	})
}

func TestSQLiteStorage_ErrorHandling(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("handles database constraint violations", func(t *testing.T) {
		storage, err := NewSQLiteStorage(helper.TempFilePath("constraints.db"))
		helper.AssertNoError(err, "creating storage")
		defer storage.Close()

		task := helper.CreateSampleTask("constraint_1", "Constraint Test")

		// Create task first time
		err = storage.Create(task)
		helper.AssertNoError(err, "creating task first time")

		// Try to create same task again (should fail due to primary key constraint)
		err = storage.Create(task)
		helper.AssertError(err, true, "creating duplicate task")
	})

	t.Run("handles database connection errors", func(t *testing.T) {
		storage, err := NewSQLiteStorage(helper.TempFilePath("connection.db"))
		helper.AssertNoError(err, "creating storage")

		// Close the database connection
		storage.Close()

		// Try to perform operations on closed database
		task := helper.CreateSampleTask("closed_1", "Closed DB Test")
		err = storage.Create(task)
		helper.AssertError(err, true, "creating task on closed database")

		_, err = storage.GetAll()
		helper.AssertError(err, true, "getting tasks from closed database")

		_, err = storage.GetByID("any_id")
		helper.AssertError(err, true, "getting task by ID from closed database")

		err = storage.Update(task)
		helper.AssertError(err, true, "updating task on closed database")

		err = storage.Delete("any_id")
		helper.AssertError(err, true, "deleting task from closed database")
	})
}

func TestSQLiteStorage_DataIntegrity(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewSQLiteStorage(helper.TempFilePath("integrity.db"))
	helper.AssertNoError(err, "creating storage")
	defer storage.Close()

	t.Run("preserves data types correctly", func(t *testing.T) {
		now := time.Now()
		dueDate := now.Add(24 * time.Hour)

		task := &models.Task{
			ID:        "integrity_1",
			Title:     "Data Integrity Test",
			Done:      true,
			CreatedAt: now,
			DueDate:   &dueDate,
		}

		err = storage.Create(task)
		helper.AssertNoError(err, "creating task")

		retrievedTask, err := storage.GetByID("integrity_1")
		helper.AssertNoError(err, "getting task")

		// Check all fields are preserved correctly
		if retrievedTask.ID != task.ID {
			t.Errorf("ID mismatch: expected %s, got %s", task.ID, retrievedTask.ID)
		}
		if retrievedTask.Title != task.Title {
			t.Errorf("Title mismatch: expected %s, got %s", task.Title, retrievedTask.Title)
		}
		if retrievedTask.Done != task.Done {
			t.Errorf("Done mismatch: expected %v, got %v", task.Done, retrievedTask.Done)
		}

		// Check timestamps (allow small difference due to precision)
		if abs(retrievedTask.CreatedAt.Sub(task.CreatedAt)) > time.Millisecond {
			t.Errorf("CreatedAt mismatch: expected %v, got %v", task.CreatedAt, retrievedTask.CreatedAt)
		}

		if retrievedTask.DueDate == nil {
			t.Error("DueDate should not be nil")
		} else if abs(retrievedTask.DueDate.Sub(*task.DueDate)) > time.Millisecond {
			t.Errorf("DueDate mismatch: expected %v, got %v", task.DueDate, retrievedTask.DueDate)
		}
	})

	t.Run("handles null due dates correctly", func(t *testing.T) {
		task := helper.CreateSampleTask("null_due", "Null Due Date Test")
		task.DueDate = nil

		err = storage.Create(task)
		helper.AssertNoError(err, "creating task with null due date")

		retrievedTask, err := storage.GetByID("null_due")
		helper.AssertNoError(err, "getting task with null due date")

		if retrievedTask.DueDate != nil {
			t.Errorf("Expected DueDate to be nil, got %v", retrievedTask.DueDate)
		}
	})
}

func TestSQLiteStorage_EdgeCases(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewSQLiteStorage(helper.TempFilePath("edge_cases.db"))
	helper.AssertNoError(err, "creating storage")
	defer storage.Close()

	t.Run("handles empty task list operations", func(t *testing.T) {
		// Try to get non-existent task from empty storage
		_, err = storage.GetByID("non_existent")
		helper.AssertError(err, true, "getting task from empty storage")

		// Try to update non-existent task
		task := helper.CreateSampleTask("update_empty", "Update Empty")
		err = storage.Update(task)
		helper.AssertError(err, true, "updating task in empty storage")

		// Try to delete non-existent task
		err = storage.Delete("delete_empty")
		helper.AssertError(err, true, "deleting task from empty storage")
	})

	t.Run("handles tasks with special characters", func(t *testing.T) {
		specialTask := &models.Task{
			ID:        "special_1",
			Title:     "Task with special chars: áéíóú ñ 中文 🚀 \"quotes\" 'apostrophes'",
			Done:      false,
			CreatedAt: time.Now(),
			DueDate:   nil,
		}

		err = storage.Create(specialTask)
		helper.AssertNoError(err, "creating task with special characters")

		retrievedTask, err := storage.GetByID("special_1")
		helper.AssertNoError(err, "getting task with special characters")

		if retrievedTask.Title != specialTask.Title {
			t.Errorf("Title mismatch: expected %s, got %s", specialTask.Title, retrievedTask.Title)
		}
	})

	t.Run("handles very long task titles", func(t *testing.T) {
		longTitle := string(make([]byte, 10000)) // Very long title
		for i := range longTitle {
			longTitle = longTitle[:i] + "a" + longTitle[i+1:]
		}

		longTask := helper.CreateSampleTask("long_1", longTitle)
		err = storage.Create(longTask)
		helper.AssertNoError(err, "creating task with long title")

		retrievedTask, err := storage.GetByID("long_1")
		helper.AssertNoError(err, "getting task with long title")

		if retrievedTask.Title != longTitle {
			t.Error("Long title was not preserved correctly")
		}
	})

	t.Run("handles extreme dates", func(t *testing.T) {
		// Test with very old date
		oldDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
		oldTask := helper.CreateSampleTaskWithDueDate("old_date", "Old Date Task", oldDate)

		err = storage.Create(oldTask)
		helper.AssertNoError(err, "creating task with old date")

		retrievedOldTask, err := storage.GetByID("old_date")
		helper.AssertNoError(err, "getting task with old date")

		if retrievedOldTask.DueDate == nil || !retrievedOldTask.DueDate.Equal(oldDate) {
			t.Errorf("Old date not preserved: expected %v, got %v", oldDate, retrievedOldTask.DueDate)
		}

		// Test with future date
		futureDate := time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC)
		futureTask := helper.CreateSampleTaskWithDueDate("future_date", "Future Date Task", futureDate)

		err = storage.Create(futureTask)
		helper.AssertNoError(err, "creating task with future date")

		retrievedFutureTask, err := storage.GetByID("future_date")
		helper.AssertNoError(err, "getting task with future date")

		if retrievedFutureTask.DueDate == nil || !retrievedFutureTask.DueDate.Equal(futureDate) {
			t.Errorf("Future date not preserved: expected %v, got %v", futureDate, retrievedFutureTask.DueDate)
		}
	})
}

// Helper function to calculate absolute duration
func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
