package storage

import "GoTask_Management/internal/models"

type Storage interface {
	Create(task *models.Task) error
	GetAll() ([]*models.Task, error)
	GetByID(id string) (*models.Task, error)
	Update(task *models.Task) error
	Delete(id string) error
	Close() error
}
