package storage

import (
	"os"
	"testing"
	"time"

	"GoTask_Management/internal/models"
)

func TestPostgreSQLStorage(t *testing.T) {
	// Skip if PostgreSQL is not available
	if os.Getenv("POSTGRES_TEST_DSN") == "" {
		t.Skip("PostgreSQL test skipped: POSTGRES_TEST_DSN not set")
	}

	config := PostgreSQLConfig{
		Host:     getEnvOrDefault("POSTGRES_TEST_HOST", "localhost"),
		Port:     5432,
		User:     getEnvOrDefault("POSTGRES_TEST_USER", "postgres"),
		Password: getEnvOrDefault("POSTGRES_TEST_PASSWORD", "password"),
		DBName:   getEnvOrDefault("POSTGRES_TEST_DB", "gotask_test"),
		SSLMode:  "disable",
		TimeZone: "UTC",
	}

	storage, err := NewPostgreSQLStorage(config)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL storage: %v", err)
	}
	defer storage.Close()

	// Clean up before tests
	cleanupPostgreSQL(t, storage)

	// Run storage compliance tests
	testStorageCompliance(t, storage)
}

func TestPostgreSQLStorage_AdditionalMethods(t *testing.T) {
	// Skip if PostgreSQL is not available
	if os.Getenv("POSTGRES_TEST_DSN") == "" {
		t.Skip("PostgreSQL test skipped: POSTGRES_TEST_DSN not set")
	}

	config := PostgreSQLConfig{
		Host:     getEnvOrDefault("POSTGRES_TEST_HOST", "localhost"),
		Port:     5432,
		User:     getEnvOrDefault("POSTGRES_TEST_USER", "postgres"),
		Password: getEnvOrDefault("POSTGRES_TEST_PASSWORD", "password"),
		DBName:   getEnvOrDefault("POSTGRES_TEST_DB", "gotask_test"),
		SSLMode:  "disable",
		TimeZone: "UTC",
	}

	storage, err := NewPostgreSQLStorage(config)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL storage: %v", err)
	}
	defer storage.Close()

	// Clean up before tests
	cleanupPostgreSQL(t, storage)

	t.Run("GetTasksByStatus", func(t *testing.T) {
		// Create test tasks
		task1 := createTestTask("pg_status_1", "Task 1", false)
		task2 := createTestTask("pg_status_2", "Task 2", true)
		task3 := createTestTask("pg_status_3", "Task 3", false)

		err := storage.Create(task1)
		if err != nil {
			t.Fatalf("Failed to create task1: %v", err)
		}
		err = storage.Create(task2)
		if err != nil {
			t.Fatalf("Failed to create task2: %v", err)
		}
		err = storage.Create(task3)
		if err != nil {
			t.Fatalf("Failed to create task3: %v", err)
		}

		// Test getting done tasks
		doneTasks, err := storage.GetTasksByStatus(true)
		if err != nil {
			t.Fatalf("Failed to get done tasks: %v", err)
		}
		if len(doneTasks) != 1 {
			t.Errorf("Expected 1 done task, got %d", len(doneTasks))
		}

		// Test getting undone tasks
		undoneTasks, err := storage.GetTasksByStatus(false)
		if err != nil {
			t.Fatalf("Failed to get undone tasks: %v", err)
		}
		if len(undoneTasks) != 2 {
			t.Errorf("Expected 2 undone tasks, got %d", len(undoneTasks))
		}
	})

	t.Run("GetTasksDueBefore", func(t *testing.T) {
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		tomorrow := now.Add(24 * time.Hour)

		// Create tasks with different due dates
		task1 := createTestTaskWithDueDate("pg_due_1", "Overdue Task", &yesterday)
		task2 := createTestTaskWithDueDate("pg_due_2", "Future Task", &tomorrow)
		task3 := createTestTask("pg_due_3", "No Due Date", false)

		err := storage.Create(task1)
		if err != nil {
			t.Fatalf("Failed to create task1: %v", err)
		}
		err = storage.Create(task2)
		if err != nil {
			t.Fatalf("Failed to create task2: %v", err)
		}
		err = storage.Create(task3)
		if err != nil {
			t.Fatalf("Failed to create task3: %v", err)
		}

		// Get tasks due before now
		dueTasks, err := storage.GetTasksDueBefore(now)
		if err != nil {
			t.Fatalf("Failed to get due tasks: %v", err)
		}
		if len(dueTasks) != 1 {
			t.Errorf("Expected 1 due task, got %d", len(dueTasks))
		}
		if dueTasks[0].ID != "pg_due_1" {
			t.Errorf("Expected overdue task, got %s", dueTasks[0].ID)
		}
	})

	t.Run("CountMethods", func(t *testing.T) {
		// Clean up first
		cleanupPostgreSQL(t, storage)

		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)

		// Create test tasks
		task1 := createTestTask("pg_count_1", "Task 1", false)
		task2 := createTestTask("pg_count_2", "Task 2", true)
		task3 := createTestTaskWithDueDate("pg_count_3", "Overdue Task", &yesterday)

		err := storage.Create(task1)
		if err != nil {
			t.Fatalf("Failed to create task1: %v", err)
		}
		err = storage.Create(task2)
		if err != nil {
			t.Fatalf("Failed to create task2: %v", err)
		}
		err = storage.Create(task3)
		if err != nil {
			t.Fatalf("Failed to create task3: %v", err)
		}

		// Test total count
		totalCount, err := storage.GetTasksCount()
		if err != nil {
			t.Fatalf("Failed to get total count: %v", err)
		}
		if totalCount != 3 {
			t.Errorf("Expected total count 3, got %d", totalCount)
		}

		// Test done count
		doneCount, err := storage.GetTasksCountByStatus(true)
		if err != nil {
			t.Fatalf("Failed to get done count: %v", err)
		}
		if doneCount != 1 {
			t.Errorf("Expected done count 1, got %d", doneCount)
		}

		// Test undone count
		undoneCount, err := storage.GetTasksCountByStatus(false)
		if err != nil {
			t.Fatalf("Failed to get undone count: %v", err)
		}
		if undoneCount != 2 {
			t.Errorf("Expected undone count 2, got %d", undoneCount)
		}

		// Test overdue count
		overdueCount, err := storage.GetOverdueTasksCount()
		if err != nil {
			t.Fatalf("Failed to get overdue count: %v", err)
		}
		if overdueCount != 1 {
			t.Errorf("Expected overdue count 1, got %d", overdueCount)
		}
	})

	t.Run("HealthCheck", func(t *testing.T) {
		err := storage.HealthCheck()
		if err != nil {
			t.Errorf("Health check failed: %v", err)
		}
	})

	t.Run("Transaction", func(t *testing.T) {
		tx := storage.BeginTransaction()
		if tx == nil {
			t.Error("Failed to begin transaction")
		}

		// Test rollback
		tx.Rollback()
	})
}

func TestPostgreSQLStorage_ErrorCases(t *testing.T) {
	t.Run("InvalidConnection", func(t *testing.T) {
		config := PostgreSQLConfig{
			Host:     "invalid-host",
			Port:     5432,
			User:     "invalid-user",
			Password: "invalid-password",
			DBName:   "invalid-db",
			SSLMode:  "disable",
			TimeZone: "UTC",
		}

		_, err := NewPostgreSQLStorage(config)
		if err == nil {
			t.Error("Expected error for invalid connection, got nil")
		}
	})
}

func cleanupPostgreSQL(t *testing.T, storage *PostgreSQLStorage) {
	// Delete all tasks for clean test environment
	db := storage.GetDB()
	if err := db.Exec("DELETE FROM tasks").Error; err != nil {
		t.Logf("Warning: Failed to cleanup PostgreSQL: %v", err)
	}
}

func createTestTask(id, title string, done bool) *models.Task {
	return &models.Task{
		ID:        id,
		Title:     title,
		Done:      done,
		CreatedAt: time.Now(),
		DueDate:   nil,
	}
}

func createTestTaskWithDueDate(id, title string, dueDate *time.Time) *models.Task {
	return &models.Task{
		ID:        id,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now(),
		DueDate:   dueDate,
	}
}
