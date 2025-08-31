package storage

import (
	"os"
	"testing"
	"time"

	"GoTask_Management/internal/models"
)

func TestMySQLStorage(t *testing.T) {
	// Skip if MySQL is not available
	if os.Getenv("MYSQL_TEST_DSN") == "" {
		t.Skip("MySQL test skipped: MYSQL_TEST_DSN not set")
	}

	config := MySQLConfig{
		Host:      getEnvOrDefault("MYSQL_TEST_HOST", "localhost"),
		Port:      3306,
		User:      getEnvOrDefault("MYSQL_TEST_USER", "root"),
		Password:  getEnvOrDefault("MYSQL_TEST_PASSWORD", "password"),
		DBName:    getEnvOrDefault("MYSQL_TEST_DB", "gotask_test"),
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}

	storage, err := NewMySQLStorage(config)
	if err != nil {
		t.Fatalf("Failed to create MySQL storage: %v", err)
	}
	defer storage.Close()

	// Clean up before tests
	cleanupMySQL(t, storage)

	// Run storage compliance tests
	testStorageCompliance(t, storage)
}

func TestMySQLStorage_AdditionalMethods(t *testing.T) {
	// Skip if MySQL is not available
	if os.Getenv("MYSQL_TEST_DSN") == "" {
		t.Skip("MySQL test skipped: MYSQL_TEST_DSN not set")
	}

	config := MySQLConfig{
		Host:      getEnvOrDefault("MYSQL_TEST_HOST", "localhost"),
		Port:      3306,
		User:      getEnvOrDefault("MYSQL_TEST_USER", "root"),
		Password:  getEnvOrDefault("MYSQL_TEST_PASSWORD", "password"),
		DBName:    getEnvOrDefault("MYSQL_TEST_DB", "gotask_test"),
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}

	storage, err := NewMySQLStorage(config)
	if err != nil {
		t.Fatalf("Failed to create MySQL storage: %v", err)
	}
	defer storage.Close()

	// Clean up before tests
	cleanupMySQL(t, storage)

	t.Run("GetTasksByStatus", func(t *testing.T) {
		// Create test tasks
		task1 := createTestTask("mysql_status_1", "Task 1", false)
		task2 := createTestTask("mysql_status_2", "Task 2", true)
		task3 := createTestTask("mysql_status_3", "Task 3", false)

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
		task1 := createTestTaskWithDueDate("mysql_due_1", "Overdue Task", &yesterday)
		task2 := createTestTaskWithDueDate("mysql_due_2", "Future Task", &tomorrow)
		task3 := createTestTask("mysql_due_3", "No Due Date", false)

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
		if dueTasks[0].ID != "mysql_due_1" {
			t.Errorf("Expected overdue task, got %s", dueTasks[0].ID)
		}
	})

	t.Run("CountMethods", func(t *testing.T) {
		// Clean up first
		cleanupMySQL(t, storage)

		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)

		// Create test tasks
		task1 := createTestTask("mysql_count_1", "Task 1", false)
		task2 := createTestTask("mysql_count_2", "Task 2", true)
		task3 := createTestTaskWithDueDate("mysql_count_3", "Overdue Task", &yesterday)

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

	t.Run("UTF8Support", func(t *testing.T) {
		// Test UTF-8 characters including emojis
		task := createTestTask("mysql_utf8", "Task with UTF-8: 你好 🚀 ñáéíóú", false)
		
		err := storage.Create(task)
		if err != nil {
			t.Fatalf("Failed to create UTF-8 task: %v", err)
		}

		retrievedTask, err := storage.GetByID("mysql_utf8")
		if err != nil {
			t.Fatalf("Failed to get UTF-8 task: %v", err)
		}

		if retrievedTask.Title != task.Title {
			t.Errorf("UTF-8 title not preserved: expected %s, got %s", task.Title, retrievedTask.Title)
		}
	})
}

func TestMySQLStorage_ErrorCases(t *testing.T) {
	t.Run("InvalidConnection", func(t *testing.T) {
		config := MySQLConfig{
			Host:      "invalid-host",
			Port:      3306,
			User:      "invalid-user",
			Password:  "invalid-password",
			DBName:    "invalid-db",
			Charset:   "utf8mb4",
			ParseTime: true,
			Loc:       "Local",
		}

		_, err := NewMySQLStorage(config)
		if err == nil {
			t.Error("Expected error for invalid connection, got nil")
		}
	})
}

func cleanupMySQL(t *testing.T, storage *MySQLStorage) {
	// Delete all tasks for clean test environment
	db := storage.GetDB()
	if err := db.Exec("DELETE FROM tasks").Error; err != nil {
		t.Logf("Warning: Failed to cleanup MySQL: %v", err)
	}
}
