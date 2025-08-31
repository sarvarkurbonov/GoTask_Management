package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"GoTask_Management/internal/models"
)

func TestNewJSONStorage(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("creates new storage with new file", func(t *testing.T) {
		filepath := helper.TempFilePath("test.json")

		storage, err := NewJSONStorage(filepath)
		helper.AssertNoError(err, "creating new JSON storage")

		if storage == nil {
			t.Error("Expected storage to be created")
		}

		// Verify file was created
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			t.Error("Expected file to be created")
		}

		// Verify file contains empty array
		tasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting all tasks from new storage")
		if len(tasks) != 0 {
			t.Errorf("Expected empty task list, got %d tasks", len(tasks))
		}
	})

	t.Run("opens existing file", func(t *testing.T) {
		filepath := helper.TempFilePath("existing.json")

		// Create file with sample data
		sampleTask := helper.CreateSampleTask("existing_1", "Existing Task")
		initialStorage, err := NewJSONStorage(filepath)
		helper.AssertNoError(err, "creating initial storage")

		err = initialStorage.Create(sampleTask)
		helper.AssertNoError(err, "creating sample task")

		// Open existing file
		storage, err := NewJSONStorage(filepath)
		helper.AssertNoError(err, "opening existing JSON storage")

		tasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting tasks from existing storage")

		if len(tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(tasks))
		}

		helper.AssertTaskEqual(sampleTask, tasks[0])
	})

	t.Run("fails with invalid directory", func(t *testing.T) {
		invalidPath := helper.CreateNonExistentPath("test.json")

		_, err := NewJSONStorage(invalidPath)
		helper.AssertError(err, true, "creating storage with invalid directory")
	})
}

func TestJSONStorage_Create(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("creates single task", func(t *testing.T) {
		storage, err := NewJSONStorage(helper.TempFilePath("create.json"))
		helper.AssertNoError(err, "creating storage")

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
		storage, err := NewJSONStorage(helper.TempFilePath("create_multiple.json"))
		helper.AssertNoError(err, "creating storage")

		tasks := helper.CreateMultipleTasks(3)

		for _, task := range tasks {
			err = storage.Create(task)
			helper.AssertNoError(err, "creating task")
		}

		// Verify all tasks were created
		allTasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting all tasks")

		helper.AssertTaskSliceEqual(tasks, allTasks)
	})

	t.Run("creates task with due date", func(t *testing.T) {
		storage, err := NewJSONStorage(helper.TempFilePath("create_due.json"))
		helper.AssertNoError(err, "creating storage")

		dueDate := time.Now().Add(24 * time.Hour)
		task := helper.CreateSampleTaskWithDueDate("due_1", "Task with Due Date", dueDate)

		err = storage.Create(task)
		helper.AssertNoError(err, "creating task with due date")

		// Verify task was created with due date
		retrievedTask, err := storage.GetByID("due_1")
		helper.AssertNoError(err, "getting task by ID")

		helper.AssertTaskEqual(task, retrievedTask)
	})
}

func TestJSONStorage_GetAll(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("returns empty list for new storage", func(t *testing.T) {
		storage, err := NewJSONStorage(helper.TempFilePath("empty.json"))
		helper.AssertNoError(err, "creating storage")

		tasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting all tasks")

		if len(tasks) != 0 {
			t.Errorf("Expected 0 tasks, got %d", len(tasks))
		}
	})

	t.Run("returns all tasks", func(t *testing.T) {
		storage, err := NewJSONStorage(helper.TempFilePath("getall.json"))
		helper.AssertNoError(err, "creating storage")

		expectedTasks := helper.CreateMultipleTasks(5)

		for _, task := range expectedTasks {
			err = storage.Create(task)
			helper.AssertNoError(err, "creating task")
		}

		tasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting all tasks")

		helper.AssertTaskSliceEqual(expectedTasks, tasks)
	})

	t.Run("handles corrupted file", func(t *testing.T) {
		corruptedFile := helper.CreateInvalidJSONFile("corrupted.json")
		storage := &JSONStorage{filepath: corruptedFile}

		_, err := storage.GetAll()
		helper.AssertError(err, true, "reading corrupted file")
	})
}

func TestJSONStorage_GetByID(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewJSONStorage(helper.TempFilePath("getbyid.json"))
	helper.AssertNoError(err, "creating storage")

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

		if err.Error() != "task not found" {
			t.Errorf("Expected 'task not found' error, got: %v", err)
		}
	})
}

func TestJSONStorage_Update(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewJSONStorage(helper.TempFilePath("update.json"))
	helper.AssertNoError(err, "creating storage")

	// Create initial task
	originalTask := helper.CreateSampleTask("update_1", "Original Task")
	err = storage.Create(originalTask)
	helper.AssertNoError(err, "creating original task")

	t.Run("updates existing task", func(t *testing.T) {
		updatedTask := &models.Task{
			ID:        "update_1",
			Title:     "Updated Task",
			Done:      true,
			CreatedAt: originalTask.CreatedAt,
			DueDate:   &time.Time{},
		}

		err = storage.Update(updatedTask)
		helper.AssertNoError(err, "updating task")

		// Verify task was updated
		retrievedTask, err := storage.GetByID("update_1")
		helper.AssertNoError(err, "getting updated task")

		helper.AssertTaskEqual(updatedTask, retrievedTask)
	})

	t.Run("returns error for non-existent task", func(t *testing.T) {
		nonExistentTask := helper.CreateSampleTask("non_existent", "Non-existent Task")

		err = storage.Update(nonExistentTask)
		helper.AssertError(err, true, "updating non-existent task")

		if err.Error() != "task not found" {
			t.Errorf("Expected 'task not found' error, got: %v", err)
		}
	})
}

func TestJSONStorage_Delete(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewJSONStorage(helper.TempFilePath("delete.json"))
	helper.AssertNoError(err, "creating storage")

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

		if err.Error() != "task not found" {
			t.Errorf("Expected 'task not found' error, got: %v", err)
		}
	})
}

func TestJSONStorage_Close(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewJSONStorage(helper.TempFilePath("close.json"))
	helper.AssertNoError(err, "creating storage")

	err = storage.Close()
	helper.AssertNoError(err, "closing storage")
}

func TestJSONStorage_ConcurrentAccess(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewJSONStorage(helper.TempFilePath("concurrent.json"))
	helper.AssertNoError(err, "creating storage")

	t.Run("concurrent reads and writes", func(t *testing.T) {
		// Create initial tasks
		for i := 0; i < 5; i++ {
			task := helper.CreateSampleTask(generateTestID(i), generateTestTitle(i))
			err = storage.Create(task)
			helper.AssertNoError(err, "creating initial task")
		}

		// Test concurrent access
		done := make(chan bool, 10)

		// Concurrent readers
		for i := 0; i < 5; i++ {
			go func() {
				defer func() { done <- true }()
				_, err := storage.GetAll()
				if err != nil {
					t.Errorf("Concurrent read error: %v", err)
				}
			}()
		}

		// Concurrent writers
		for i := 5; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()
				task := helper.CreateSampleTask(generateTestID(index), generateTestTitle(index))
				err := storage.Create(task)
				if err != nil {
					t.Errorf("Concurrent write error: %v", err)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestJSONStorage_ErrorHandling(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("handles file permission errors", func(t *testing.T) {
		// Create a read-only file
		readOnlyFile := helper.CreateReadOnlyFile("readonly.json")
		storage := &JSONStorage{filepath: readOnlyFile}

		task := helper.CreateSampleTask("perm_1", "Permission Test")
		err := storage.Create(task)
		helper.AssertError(err, true, "creating task in read-only file")
	})

	t.Run("handles invalid file path on save", func(t *testing.T) {
		invalidPath := filepath.Join(helper.CreateNonExistentPath("invalid"), "test.json")
		storage := &JSONStorage{filepath: invalidPath}

		task := helper.CreateSampleTask("invalid_1", "Invalid Path Test")
		err := storage.Create(task)
		helper.AssertError(err, true, "creating task with invalid path")
	})

	t.Run("handles corrupted JSON on load", func(t *testing.T) {
		corruptedFile := helper.CreateInvalidJSONFile("corrupted_load.json")
		storage := &JSONStorage{filepath: corruptedFile}

		_, err := storage.GetByID("any_id")
		helper.AssertError(err, true, "loading corrupted JSON")
	})

	t.Run("handles missing file on load", func(t *testing.T) {
		nonExistentFile := helper.TempFilePath("missing.json")
		storage := &JSONStorage{filepath: nonExistentFile}

		_, err := storage.GetAll()
		helper.AssertError(err, true, "loading missing file")
	})
}

func TestJSONStorage_AtomicWrites(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	storage, err := NewJSONStorage(helper.TempFilePath("atomic.json"))
	helper.AssertNoError(err, "creating storage")

	t.Run("verifies atomic write behavior", func(t *testing.T) {
		// Create a task
		task := helper.CreateSampleTask("atomic_1", "Atomic Test")
		err = storage.Create(task)
		helper.AssertNoError(err, "creating task")

		// Verify temp file doesn't exist after successful write
		tempFile := storage.filepath + ".tmp"
		if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
			t.Error("Temporary file should not exist after successful write")
		}

		// Verify main file exists and contains correct data
		tasks, err := storage.GetAll()
		helper.AssertNoError(err, "getting tasks after atomic write")

		if len(tasks) != 1 {
			t.Errorf("Expected 1 task after atomic write, got %d", len(tasks))
		}

		helper.AssertTaskEqual(task, tasks[0])
	})
}

func TestJSONStorage_EdgeCases(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	t.Run("handles empty task list operations", func(t *testing.T) {
		storage, err := NewJSONStorage(helper.TempFilePath("empty_ops.json"))
		helper.AssertNoError(err, "creating storage")

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
		storage, err := NewJSONStorage(helper.TempFilePath("special_chars.json"))
		helper.AssertNoError(err, "creating storage")

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

		helper.AssertTaskEqual(specialTask, retrievedTask)
	})

	t.Run("handles very long task titles", func(t *testing.T) {
		storage, err := NewJSONStorage(helper.TempFilePath("long_title.json"))
		helper.AssertNoError(err, "creating storage")

		longTitle := string(make([]byte, 10000)) // Very long title
		for i := range longTitle {
			longTitle = longTitle[:i] + "a" + longTitle[i+1:]
		}

		longTask := helper.CreateSampleTask("long_1", longTitle)
		err = storage.Create(longTask)
		helper.AssertNoError(err, "creating task with long title")

		retrievedTask, err := storage.GetByID("long_1")
		helper.AssertNoError(err, "getting task with long title")

		helper.AssertTaskEqual(longTask, retrievedTask)
	})
}
