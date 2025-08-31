package task

import (
	"testing"
	"time"

	"GoTask_Management/internal/models"
)

func TestNewService(t *testing.T) {
	helper := NewTestHelper(t)

	service := NewService(helper.GetMockStorage())
	if service == nil {
		t.Error("Expected service to be created")
	}

	if service.storage == nil {
		t.Error("Expected service to have storage")
	}
}

func TestService_CreateTask(t *testing.T) {
	helper := NewTestHelper(t)
	service := helper.GetService()

	t.Run("creates task successfully", func(t *testing.T) {
		title := "Test Task"
		dueDate := time.Now().Add(24 * time.Hour)

		task, err := service.CreateTask(title, &dueDate)
		helper.AssertNoError(err, "creating task")

		if task == nil {
			t.Error("Expected task to be created")
		}

		if task.Title != title {
			t.Errorf("Expected title %s, got %s", title, task.Title)
		}

		if task.Done {
			t.Error("Expected new task to not be done")
		}

		if task.DueDate == nil || !task.DueDate.Equal(dueDate) {
			t.Errorf("Expected due date %v, got %v", dueDate, task.DueDate)
		}

		if task.ID == "" {
			t.Error("Expected task to have an ID")
		}

		if task.CreatedAt.IsZero() {
			t.Error("Expected task to have a creation time")
		}
	})

	t.Run("creates task without due date", func(t *testing.T) {
		title := "Task without due date"

		task, err := service.CreateTask(title, nil)
		helper.AssertNoError(err, "creating task without due date")

		if task.DueDate != nil {
			t.Error("Expected due date to be nil")
		}
	})

	t.Run("fails with empty title", func(t *testing.T) {
		_, err := service.CreateTask("", nil)
		helper.AssertError(err, true, "creating task with empty title")

		if err.Error() != "task title cannot be empty" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("handles storage error", func(t *testing.T) {
		helper.GetMockStorage().SetError(true, "storage error")

		_, err := service.CreateTask("Test Task", nil)
		helper.AssertError(err, true, "creating task with storage error")

		// Reset error state
		helper.GetMockStorage().SetError(false, "")
	})
}

func TestService_ListTasks(t *testing.T) {
	helper := NewTestHelper(t)
	service := helper.GetService()

	// Seed with test data
	tasks := []*models.Task{
		helper.CreateSampleTask("task1", "Task 1"),
		helper.CreateCompletedTask("task2", "Task 2"),
		helper.CreateSampleTask("task3", "Task 3"),
		helper.CreateCompletedTask("task4", "Task 4"),
	}
	helper.SeedMockStorage(tasks)

	t.Run("lists all tasks", func(t *testing.T) {
		allTasks, err := service.ListTasks("")
		helper.AssertNoError(err, "listing all tasks")

		if len(allTasks) != 4 {
			t.Errorf("Expected 4 tasks, got %d", len(allTasks))
		}
	})

	t.Run("filters done tasks", func(t *testing.T) {
		doneTasks, err := service.ListTasks("done")
		helper.AssertNoError(err, "listing done tasks")

		if len(doneTasks) != 2 {
			t.Errorf("Expected 2 done tasks, got %d", len(doneTasks))
		}

		for _, task := range doneTasks {
			if !task.Done {
				t.Error("Expected all tasks to be done")
			}
		}
	})

	t.Run("filters undone tasks", func(t *testing.T) {
		undoneTasks, err := service.ListTasks("undone")
		helper.AssertNoError(err, "listing undone tasks")

		if len(undoneTasks) != 2 {
			t.Errorf("Expected 2 undone tasks, got %d", len(undoneTasks))
		}

		for _, task := range undoneTasks {
			if task.Done {
				t.Error("Expected all tasks to be undone")
			}
		}
	})

	t.Run("handles storage error", func(t *testing.T) {
		helper.GetMockStorage().SetError(true, "storage error")

		_, err := service.ListTasks("")
		helper.AssertError(err, true, "listing tasks with storage error")

		// Reset error state
		helper.GetMockStorage().SetError(false, "")
	})
}

func TestService_GetTask(t *testing.T) {
	helper := NewTestHelper(t)
	service := helper.GetService()

	// Seed with test data
	task := helper.CreateSampleTask("test_task", "Test Task")
	helper.SeedMockStorage([]*models.Task{task})

	t.Run("gets existing task", func(t *testing.T) {
		retrievedTask, err := service.GetTask("test_task")
		helper.AssertNoError(err, "getting existing task")

		helper.AssertTaskEqual(task, retrievedTask)
	})

	t.Run("fails for non-existent task", func(t *testing.T) {
		_, err := service.GetTask("non_existent")
		helper.AssertError(err, true, "getting non-existent task")
	})

	t.Run("handles storage error", func(t *testing.T) {
		helper.GetMockStorage().SetError(true, "storage error")

		_, err := service.GetTask("test_task")
		helper.AssertError(err, true, "getting task with storage error")

		// Reset error state
		helper.GetMockStorage().SetError(false, "")
	})
}

func TestService_UpdateTask(t *testing.T) {
	helper := NewTestHelper(t)
	service := helper.GetService()

	// Seed with test data
	originalTask := helper.CreateSampleTask("update_task", "Original Task")
	helper.SeedMockStorage([]*models.Task{originalTask})

	t.Run("updates task successfully", func(t *testing.T) {
		newTitle := "Updated Task"
		newDueDate := time.Now().Add(48 * time.Hour)

		updatedTask, err := service.UpdateTask("update_task", newTitle, true, &newDueDate)
		helper.AssertNoError(err, "updating task")

		if updatedTask.Title != newTitle {
			t.Errorf("Expected title %s, got %s", newTitle, updatedTask.Title)
		}

		if !updatedTask.Done {
			t.Error("Expected task to be done")
		}

		if updatedTask.DueDate == nil || !updatedTask.DueDate.Equal(newDueDate) {
			t.Errorf("Expected due date %v, got %v", newDueDate, updatedTask.DueDate)
		}

		// CreatedAt should remain unchanged
		if !updatedTask.CreatedAt.Equal(originalTask.CreatedAt) {
			t.Error("CreatedAt should not change during update")
		}
	})

	t.Run("updates with empty title keeps original", func(t *testing.T) {
		updatedTask, err := service.UpdateTask("update_task", "", false, nil)
		helper.AssertNoError(err, "updating task with empty title")

		// Title should remain unchanged when empty string is provided
		if updatedTask.Title != "Updated Task" { // From previous test
			t.Error("Title should remain unchanged when empty string provided")
		}
	})

	t.Run("fails for non-existent task", func(t *testing.T) {
		_, err := service.UpdateTask("non_existent", "New Title", false, nil)
		helper.AssertError(err, true, "updating non-existent task")
	})

	t.Run("handles storage get error", func(t *testing.T) {
		helper.GetMockStorage().SetError(true, "get error")

		_, err := service.UpdateTask("update_task", "New Title", false, nil)
		helper.AssertError(err, true, "updating task with storage get error")

		// Reset error state
		helper.GetMockStorage().SetError(false, "")
	})
}

func TestService_MarkTaskDone(t *testing.T) {
	helper := NewTestHelper(t)
	service := helper.GetService()

	// Seed with test data
	task := helper.CreateSampleTask("mark_done_task", "Mark Done Task")
	helper.SeedMockStorage([]*models.Task{task})

	t.Run("marks task as done", func(t *testing.T) {
		err := service.MarkTaskDone("mark_done_task", true)
		helper.AssertNoError(err, "marking task as done")

		// Verify task is marked as done in storage
		updatedTask := helper.GetMockStorage().tasks["mark_done_task"]
		if !updatedTask.Done {
			t.Error("Expected task to be marked as done")
		}
	})

	t.Run("marks task as undone", func(t *testing.T) {
		err := service.MarkTaskDone("mark_done_task", false)
		helper.AssertNoError(err, "marking task as undone")

		// Verify task is marked as undone in storage
		updatedTask := helper.GetMockStorage().tasks["mark_done_task"]
		if updatedTask.Done {
			t.Error("Expected task to be marked as undone")
		}
	})

	t.Run("fails for non-existent task", func(t *testing.T) {
		err := service.MarkTaskDone("non_existent", true)
		helper.AssertError(err, true, "marking non-existent task as done")
	})

	t.Run("handles storage error", func(t *testing.T) {
		helper.GetMockStorage().SetError(true, "storage error")

		err := service.MarkTaskDone("mark_done_task", true)
		helper.AssertError(err, true, "marking task done with storage error")

		// Reset error state
		helper.GetMockStorage().SetError(false, "")
	})
}

func TestService_DeleteTask(t *testing.T) {
	helper := NewTestHelper(t)
	service := helper.GetService()

	// Seed with test data
	task := helper.CreateSampleTask("delete_task", "Delete Task")
	helper.SeedMockStorage([]*models.Task{task})

	t.Run("deletes task successfully", func(t *testing.T) {
		err := service.DeleteTask("delete_task")
		helper.AssertNoError(err, "deleting task")

		// Verify task is deleted from storage
		_, exists := helper.GetMockStorage().tasks["delete_task"]
		if exists {
			t.Error("Expected task to be deleted from storage")
		}
	})

	t.Run("fails for non-existent task", func(t *testing.T) {
		err := service.DeleteTask("non_existent")
		helper.AssertError(err, true, "deleting non-existent task")
	})

	t.Run("handles storage error", func(t *testing.T) {
		helper.GetMockStorage().SetError(true, "storage error")

		err := service.DeleteTask("any_task")
		helper.AssertError(err, true, "deleting task with storage error")

		// Reset error state
		helper.GetMockStorage().SetError(false, "")
	})
}

func TestService_GetDueTasks(t *testing.T) {
	helper := NewTestHelper(t)
	service := helper.GetService()

	// Seed with test data
	tasks := []*models.Task{
		helper.CreateDueSoonTask("due_today", "Due Today", 12),          // Due in 12 hours
		helper.CreateDueSoonTask("due_tomorrow", "Due Tomorrow", 36),    // Due in 36 hours
		helper.CreateDueSoonTask("due_next_week", "Due Next Week", 168), // Due in 7 days
		helper.CreateOverdueTask("overdue", "Overdue Task"),             // Already overdue
		helper.CreateSampleTask("no_due_date", "No Due Date"),           // No due date
	}
	helper.SeedMockStorage(tasks)

	t.Run("gets tasks due within specified days", func(t *testing.T) {
		dueTasks, err := service.GetDueTasks(2) // Next 2 days
		helper.AssertNoError(err, "getting due tasks")

		// Should include: due_today, due_tomorrow, and overdue
		expectedCount := 3
		if len(dueTasks) != expectedCount {
			t.Errorf("Expected %d due tasks, got %d", expectedCount, len(dueTasks))
		}

		// Verify correct tasks are included
		taskIDs := make(map[string]bool)
		for _, task := range dueTasks {
			taskIDs[task.ID] = true
		}

		if !taskIDs["due_today"] {
			t.Error("Expected 'due_today' task to be included")
		}
		if !taskIDs["due_tomorrow"] {
			t.Error("Expected 'due_tomorrow' task to be included")
		}
		if !taskIDs["overdue"] {
			t.Error("Expected 'overdue' task to be included")
		}
		if taskIDs["due_next_week"] {
			t.Error("Expected 'due_next_week' task to NOT be included")
		}
		if taskIDs["no_due_date"] {
			t.Error("Expected 'no_due_date' task to NOT be included")
		}
	})

	t.Run("gets tasks due within 7 days", func(t *testing.T) {
		dueTasks, err := service.GetDueTasks(7)
		helper.AssertNoError(err, "getting due tasks within 7 days")

		// Should include all tasks with due dates
		expectedCount := 4
		if len(dueTasks) != expectedCount {
			t.Errorf("Expected %d due tasks, got %d", expectedCount, len(dueTasks))
		}
	})

	t.Run("handles empty result", func(t *testing.T) {
		// Clear storage and add only tasks without due dates
		helper.GetMockStorage().tasks = make(map[string]*models.Task)
		noDueDateTask := helper.CreateSampleTask("no_due", "No Due Date")
		helper.SeedMockStorage([]*models.Task{noDueDateTask})

		dueTasks, err := service.GetDueTasks(7)
		helper.AssertNoError(err, "getting due tasks with no results")

		if len(dueTasks) != 0 {
			t.Errorf("Expected 0 due tasks, got %d", len(dueTasks))
		}
	})

	t.Run("handles storage error", func(t *testing.T) {
		helper.GetMockStorage().SetError(true, "storage error")

		_, err := service.GetDueTasks(7)
		helper.AssertError(err, true, "getting due tasks with storage error")

		// Reset error state
		helper.GetMockStorage().SetError(false, "")
	})
}

func TestService_GetTasksSummary(t *testing.T) {
	helper := NewTestHelper(t)
	service := helper.GetService()

	// Seed with test data
	tasks := []*models.Task{
		helper.CreateSampleTask("task1", "Task 1"),             // Undone, no due date
		helper.CreateCompletedTask("task2", "Task 2"),          // Done
		helper.CreateSampleTask("task3", "Task 3"),             // Undone, no due date
		helper.CreateCompletedTask("task4", "Task 4"),          // Done
		helper.CreateOverdueTask("overdue1", "Overdue Task 1"), // Undone, overdue
		helper.CreateOverdueTask("overdue2", "Overdue Task 2"), // Undone, overdue
	}

	// Mark one overdue task as done
	overdueButDone := helper.CreateOverdueTask("overdue_done", "Overdue but Done")
	overdueButDone.Done = true
	tasks = append(tasks, overdueButDone)

	helper.SeedMockStorage(tasks)

	t.Run("calculates summary correctly", func(t *testing.T) {
		total, done, overdue, err := service.GetTasksSummary()
		helper.AssertNoError(err, "getting tasks summary")

		expectedTotal := 7
		expectedDone := 3    // task2, task4, overdue_done
		expectedOverdue := 2 // overdue1, overdue2 (overdue_done is done so not counted)

		if total != expectedTotal {
			t.Errorf("Expected total %d, got %d", expectedTotal, total)
		}

		if done != expectedDone {
			t.Errorf("Expected done %d, got %d", expectedDone, done)
		}

		if overdue != expectedOverdue {
			t.Errorf("Expected overdue %d, got %d", expectedOverdue, overdue)
		}
	})

	t.Run("handles empty storage", func(t *testing.T) {
		// Clear storage
		helper.GetMockStorage().tasks = make(map[string]*models.Task)

		total, done, overdue, err := service.GetTasksSummary()
		helper.AssertNoError(err, "getting summary with empty storage")

		if total != 0 || done != 0 || overdue != 0 {
			t.Errorf("Expected all counts to be 0, got total=%d, done=%d, overdue=%d", total, done, overdue)
		}
	})

	t.Run("handles storage error", func(t *testing.T) {
		helper.GetMockStorage().SetError(true, "storage error")

		_, _, _, err := service.GetTasksSummary()
		helper.AssertError(err, true, "getting summary with storage error")

		// Reset error state
		helper.GetMockStorage().SetError(false, "")
	})
}

func TestGenerateID(t *testing.T) {
	t.Run("generates unique IDs", func(t *testing.T) {
		// Generate multiple IDs to increase chance of uniqueness
		ids := make(map[string]bool)
		for i := 0; i < 10; i++ {
			id := generateID()

			if id == "" {
				t.Error("Expected non-empty ID")
			}

			// Check format
			if len(id) < 5 || id[:5] != "task_" {
				t.Errorf("Expected ID to start with 'task_', got %s", id)
			}

			// Check for duplicates
			if ids[id] {
				t.Errorf("Generated duplicate ID: %s", id)
			}
			ids[id] = true

			// Small delay to ensure different timestamps
			time.Sleep(1 * time.Microsecond)
		}

		// Should have generated 10 unique IDs
		if len(ids) != 10 {
			t.Errorf("Expected 10 unique IDs, got %d", len(ids))
		}
	})
}

func TestService_EdgeCases(t *testing.T) {
	helper := NewTestHelper(t)
	service := helper.GetService()

	t.Run("handles tasks with special characters", func(t *testing.T) {
		specialTitle := "Task with special chars: áéíóú ñ 中文 🚀 \"quotes\" 'apostrophes'"

		task, err := service.CreateTask(specialTitle, nil)
		helper.AssertNoError(err, "creating task with special characters")

		if task.Title != specialTitle {
			t.Errorf("Expected title to be preserved: %s", specialTitle)
		}
	})

	t.Run("handles very long titles", func(t *testing.T) {
		longTitle := string(make([]byte, 1000))
		for i := range longTitle {
			longTitle = longTitle[:i] + "a" + longTitle[i+1:]
		}

		task, err := service.CreateTask(longTitle, nil)
		helper.AssertNoError(err, "creating task with long title")

		if task.Title != longTitle {
			t.Error("Expected long title to be preserved")
		}
	})

	t.Run("handles extreme due dates", func(t *testing.T) {
		// Very far future date
		futureDate := time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC)

		task, err := service.CreateTask("Future Task", &futureDate)
		helper.AssertNoError(err, "creating task with future date")

		if task.DueDate == nil || !task.DueDate.Equal(futureDate) {
			t.Errorf("Expected future date to be preserved: %v", futureDate)
		}

		// Very old date
		oldDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

		task2, err := service.CreateTask("Old Task", &oldDate)
		helper.AssertNoError(err, "creating task with old date")

		if task2.DueDate == nil || !task2.DueDate.Equal(oldDate) {
			t.Errorf("Expected old date to be preserved: %v", oldDate)
		}
	})
}
