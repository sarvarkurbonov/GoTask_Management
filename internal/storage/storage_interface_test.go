package storage

import (
	"testing"
	"time"

	"GoTask_Management/internal/models"
)

// TestStorageInterface tests that both implementations comply with the Storage interface
func TestStorageInterface(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Test both implementations
	implementations := []struct {
		name    string
		factory func() (Storage, error)
	}{
		{
			name: "JSONStorage",
			factory: func() (Storage, error) {
				return NewJSONStorage(helper.TempFilePath("interface_json.json"))
			},
		},
		{
			name: "SQLiteStorage",
			factory: func() (Storage, error) {
				return NewSQLiteStorage(helper.TempFilePath("interface_sqlite.db"))
			},
		},
	}

	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			storage, err := impl.factory()
			helper.AssertNoError(err, "creating storage implementation")
			defer storage.Close()

			testStorageCompliance(t, helper, storage)
		})
	}
}

// testStorageCompliance runs a comprehensive test suite against any Storage implementation
func testStorageCompliance(t *testing.T, helper *TestHelper, storage Storage) {
	t.Run("CRUD operations", func(t *testing.T) {
		testCRUDOperations(t, helper, storage)
	})

	t.Run("error handling", func(t *testing.T) {
		testErrorHandling(t, helper, storage)
	})

	t.Run("data consistency", func(t *testing.T) {
		testDataConsistency(t, helper, storage)
	})

	t.Run("edge cases", func(t *testing.T) {
		testEdgeCases(t, helper, storage)
	})
}

func testCRUDOperations(t *testing.T, helper *TestHelper, storage Storage) {
	// Test Create
	task1 := helper.CreateSampleTask("crud_1", "CRUD Test 1")
	err := storage.Create(task1)
	helper.AssertNoError(err, "creating task1")

	task2 := helper.CreateSampleTaskWithDueDate("crud_2", "CRUD Test 2", time.Now().Add(24*time.Hour))
	err = storage.Create(task2)
	helper.AssertNoError(err, "creating task2")

	// Test GetAll
	allTasks, err := storage.GetAll()
	helper.AssertNoError(err, "getting all tasks")
	if len(allTasks) < 2 {
		t.Errorf("Expected at least 2 tasks, got %d", len(allTasks))
	}

	// Test GetByID
	retrievedTask1, err := storage.GetByID("crud_1")
	helper.AssertNoError(err, "getting task1 by ID")
	helper.AssertTaskEqual(task1, retrievedTask1)

	retrievedTask2, err := storage.GetByID("crud_2")
	helper.AssertNoError(err, "getting task2 by ID")
	helper.AssertTaskEqual(task2, retrievedTask2)

	// Test Update
	task1.Title = "Updated CRUD Test 1"
	task1.Done = true
	err = storage.Update(task1)
	helper.AssertNoError(err, "updating task1")

	updatedTask1, err := storage.GetByID("crud_1")
	helper.AssertNoError(err, "getting updated task1")
	helper.AssertTaskEqual(task1, updatedTask1)

	// Test Delete
	err = storage.Delete("crud_2")
	helper.AssertNoError(err, "deleting task2")

	_, err = storage.GetByID("crud_2")
	helper.AssertError(err, true, "getting deleted task2")
}

func testErrorHandling(t *testing.T, helper *TestHelper, storage Storage) {
	// Test GetByID with non-existent ID
	_, err := storage.GetByID("non_existent_id")
	helper.AssertError(err, true, "getting non-existent task")

	// Test Update with non-existent task
	nonExistentTask := helper.CreateSampleTask("non_existent_update", "Non-existent Update")
	err = storage.Update(nonExistentTask)
	helper.AssertError(err, true, "updating non-existent task")

	// Test Delete with non-existent ID
	err = storage.Delete("non_existent_delete")
	helper.AssertError(err, true, "deleting non-existent task")
}

func testDataConsistency(t *testing.T, helper *TestHelper, storage Storage) {
	// Create tasks with various data types
	now := time.Now()
	dueDate := now.Add(48 * time.Hour)

	tasks := []*models.Task{
		{
			ID:        "consistency_1",
			Title:     "Task with no due date",
			Done:      false,
			CreatedAt: now,
			DueDate:   nil,
		},
		{
			ID:        "consistency_2",
			Title:     "Task with due date",
			Done:      true,
			CreatedAt: now,
			DueDate:   &dueDate,
		},
		{
			ID:        "consistency_3",
			Title:     "Task with special chars: áéíóú ñ 中文 🚀",
			Done:      false,
			CreatedAt: now,
			DueDate:   nil,
		},
	}

	// Create all tasks
	for _, task := range tasks {
		err := storage.Create(task)
		helper.AssertNoError(err, "creating consistency task")
	}

	// Retrieve and verify all tasks
	for _, expectedTask := range tasks {
		retrievedTask, err := storage.GetByID(expectedTask.ID)
		helper.AssertNoError(err, "getting consistency task")
		helper.AssertTaskEqual(expectedTask, retrievedTask)
	}

	// Test bulk retrieval
	allTasks, err := storage.GetAll()
	helper.AssertNoError(err, "getting all consistency tasks")

	// Verify all created tasks are in the result
	for _, expectedTask := range tasks {
		found := false
		for _, task := range allTasks {
			if task.ID == expectedTask.ID {
				helper.AssertTaskEqual(expectedTask, task)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Task %s not found in GetAll result", expectedTask.ID)
		}
	}
}

func testEdgeCases(t *testing.T, helper *TestHelper, storage Storage) {
	// Test empty string title
	emptyTitleTask := &models.Task{
		ID:        "empty_title",
		Title:     "",
		Done:      false,
		CreatedAt: time.Now(),
		DueDate:   nil,
	}
	err := storage.Create(emptyTitleTask)
	helper.AssertNoError(err, "creating task with empty title")

	retrievedEmptyTask, err := storage.GetByID("empty_title")
	helper.AssertNoError(err, "getting task with empty title")
	if retrievedEmptyTask.Title != "" {
		t.Errorf("Expected empty title, got '%s'", retrievedEmptyTask.Title)
	}

	// Test very long title
	longTitle := make([]byte, 1000)
	for i := range longTitle {
		longTitle[i] = 'a'
	}
	longTitleTask := &models.Task{
		ID:        "long_title",
		Title:     string(longTitle),
		Done:      false,
		CreatedAt: time.Now(),
		DueDate:   nil,
	}
	err = storage.Create(longTitleTask)
	helper.AssertNoError(err, "creating task with long title")

	retrievedLongTask, err := storage.GetByID("long_title")
	helper.AssertNoError(err, "getting task with long title")
	if retrievedLongTask.Title != string(longTitle) {
		t.Error("Long title was not preserved correctly")
	}

	// Test extreme dates
	extremeDate := time.Date(2200, 12, 31, 23, 59, 59, 0, time.UTC)
	extremeDateTask := &models.Task{
		ID:        "extreme_date",
		Title:     "Extreme Date Task",
		Done:      false,
		CreatedAt: time.Now(),
		DueDate:   &extremeDate,
	}
	err = storage.Create(extremeDateTask)
	helper.AssertNoError(err, "creating task with extreme date")

	retrievedExtremeTask, err := storage.GetByID("extreme_date")
	helper.AssertNoError(err, "getting task with extreme date")
	if retrievedExtremeTask.DueDate == nil || !retrievedExtremeTask.DueDate.Equal(extremeDate) {
		t.Errorf("Extreme date not preserved: expected %v, got %v", extremeDate, retrievedExtremeTask.DueDate)
	}

	// Test updating task to remove due date
	extremeDateTask.DueDate = nil
	err = storage.Update(extremeDateTask)
	helper.AssertNoError(err, "updating task to remove due date")

	updatedExtremeTask, err := storage.GetByID("extreme_date")
	helper.AssertNoError(err, "getting task after removing due date")
	if updatedExtremeTask.DueDate != nil {
		t.Errorf("Expected due date to be nil after update, got %v", updatedExtremeTask.DueDate)
	}
}

// TestStorageInterfaceCompliance verifies that both implementations satisfy the Storage interface
func TestStorageInterfaceCompliance(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Test that both implementations can be assigned to Storage interface
	var storage Storage

	// Test JSONStorage
	jsonStorage, err := NewJSONStorage(helper.TempFilePath("compliance_json.json"))
	helper.AssertNoError(err, "creating JSON storage")
	storage = jsonStorage
	if storage == nil {
		t.Error("JSONStorage does not implement Storage interface")
	}
	jsonStorage.Close()

	// Test SQLiteStorage
	sqliteStorage, err := NewSQLiteStorage(helper.TempFilePath("compliance_sqlite.db"))
	helper.AssertNoError(err, "creating SQLite storage")
	storage = sqliteStorage
	if storage == nil {
		t.Error("SQLiteStorage does not implement Storage interface")
	}
	sqliteStorage.Close()
}

// BenchmarkStorageOperations benchmarks both storage implementations
func BenchmarkStorageOperations(b *testing.B) {
	helper := NewTestHelper(&testing.T{})
	defer helper.Cleanup()

	implementations := []struct {
		name    string
		factory func() (Storage, error)
	}{
		{
			name: "JSONStorage",
			factory: func() (Storage, error) {
				return NewJSONStorage(helper.TempFilePath("bench_json.json"))
			},
		},
		{
			name: "SQLiteStorage",
			factory: func() (Storage, error) {
				return NewSQLiteStorage(helper.TempFilePath("bench_sqlite.db"))
			},
		},
	}

	for _, impl := range implementations {
		b.Run(impl.name, func(b *testing.B) {
			storage, err := impl.factory()
			if err != nil {
				b.Fatalf("Failed to create storage: %v", err)
			}
			defer storage.Close()

			// Benchmark Create operations
			b.Run("Create", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					task := helper.CreateSampleTask(generateTestID(i), generateTestTitle(i))
					storage.Create(task)
				}
			})
		})
	}
}
