package task

import (
	"fmt"
	"strings"
	"time"

	"GoTask_Management/internal/models"
	"GoTask_Management/internal/storage"
)

type Service struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) CreateTask(title string, dueDate *time.Time) (*models.Task, error) {
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("task title cannot be empty")
	}

	task := &models.Task{
		ID:        generateID(),
		Title:     title,
		Done:      false,
		CreatedAt: time.Now(),
		DueDate:   dueDate,
	}

	if err := s.storage.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) ListTasks(status string) ([]*models.Task, error) {
	tasks, err := s.storage.GetAll()
	if err != nil {
		return nil, err
	}

	if status == "" {
		return tasks, nil
	}

	filtered := make([]*models.Task, 0)
	for _, task := range tasks {
		if (status == "done" && task.Done) || (status == "undone" && !task.Done) {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

func (s *Service) GetTask(id string) (*models.Task, error) {
	return s.storage.GetByID(id)
}

func (s *Service) UpdateTask(id string, title string, done bool, dueDate *time.Time) (*models.Task, error) {
	task, err := s.storage.GetByID(id)
	if err != nil {
		return nil, err
	}

	if title != "" {
		task.Title = title
	}
	task.Done = done
	if dueDate != nil {
		task.DueDate = dueDate
	}

	if err := s.storage.Update(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) MarkTaskDone(id string, done bool) error {
	task, err := s.storage.GetByID(id)
	if err != nil {
		return err
	}

	task.Done = done
	return s.storage.Update(task)
}

func (s *Service) DeleteTask(id string) error {
	return s.storage.Delete(id)
}

func (s *Service) GetDueTasks(days int) ([]*models.Task, error) {
	tasks, err := s.storage.GetAll()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	deadline := now.AddDate(0, 0, days)

	dueTasks := make([]*models.Task, 0)
	for _, task := range tasks {
		if task.DueDate != nil && !task.DueDate.After(deadline) {
			dueTasks = append(dueTasks, task)
		}
	}

	return dueTasks, nil
}

func (s *Service) GetTasksSummary() (int, int, int, error) {
	tasks, err := s.storage.GetAll()
	if err != nil {
		return 0, 0, 0, err
	}

	total := len(tasks)
	done := 0
	overdue := 0
	now := time.Now()

	for _, task := range tasks {
		if task.Done {
			done++
		}
		if task.DueDate != nil && task.DueDate.Before(now) && !task.Done {
			overdue++
		}
	}

	return total, done, overdue, nil
}

// HealthCheck performs a health check on the service and its dependencies
func (s *Service) HealthCheck() error {
	// Check if storage supports health checks
	if healthChecker, ok := s.storage.(interface{ HealthCheck() error }); ok {
		return healthChecker.HealthCheck()
	}

	// Fallback: try a simple operation to verify storage is working
	_, err := s.storage.GetAll()
	return err
}

func generateID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
