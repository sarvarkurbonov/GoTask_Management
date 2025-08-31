package api

import (
	"time"

	"GoTask_Management/internal/models"
)

// TaskService defines the interface for task operations
type TaskService interface {
	CreateTask(title string, dueDate *time.Time) (*models.Task, error)
	ListTasks(status string) ([]*models.Task, error)
	GetTask(id string) (*models.Task, error)
	UpdateTask(id string, title string, done bool, dueDate *time.Time) (*models.Task, error)
	DeleteTask(id string) error
	GetDueTasks(days int) ([]*models.Task, error)
	GetTasksSummary() (int, int, int, error)
}
