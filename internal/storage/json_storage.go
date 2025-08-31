package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"GoTask_Management/internal/models"
)

type JSONStorage struct {
	filepath string
	mu       sync.RWMutex
}

func NewJSONStorage(filepath string) (*JSONStorage, error) {
	js := &JSONStorage{
		filepath: filepath,
	}

	// Create file if it doesn't exist
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		if err := js.save(make([]*models.Task, 0)); err != nil {
			return nil, err
		}
	}

	return js, nil
}

func (js *JSONStorage) Create(task *models.Task) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	tasks, err := js.load()
	if err != nil {
		return err
	}

	tasks = append(tasks, task)
	return js.save(tasks)
}

func (js *JSONStorage) GetAll() ([]*models.Task, error) {
	js.mu.RLock()
	defer js.mu.RUnlock()

	return js.load()
}

func (js *JSONStorage) GetByID(id string) (*models.Task, error) {
	js.mu.RLock()
	defer js.mu.RUnlock()

	tasks, err := js.load()
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.ID == id {
			return task, nil
		}
	}

	return nil, fmt.Errorf("task not found")
}

func (js *JSONStorage) Update(task *models.Task) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	tasks, err := js.load()
	if err != nil {
		return err
	}

	for i, t := range tasks {
		if t.ID == task.ID {
			tasks[i] = task
			return js.save(tasks)
		}
	}

	return fmt.Errorf("task not found")
}

func (js *JSONStorage) Delete(id string) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	tasks, err := js.load()
	if err != nil {
		return err
	}

	filtered := make([]*models.Task, 0, len(tasks))
	found := false
	for _, task := range tasks {
		if task.ID != id {
			filtered = append(filtered, task)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("task not found")
	}

	return js.save(filtered)
}

func (js *JSONStorage) Close() error {
	return nil
}

func (js *JSONStorage) load() ([]*models.Task, error) {
	data, err := os.ReadFile(js.filepath)
	if err != nil {
		return nil, err
	}

	var tasks []*models.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (js *JSONStorage) save(tasks []*models.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first (atomic write)
	tempFile := js.filepath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	// Rename temp file to actual file
	return os.Rename(tempFile, js.filepath)
}
