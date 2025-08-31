package storage

import (
	"database/sql"

	"GoTask_Management/internal/models"

	_ "modernc.org/sqlite"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS tasks (
        id TEXT PRIMARY KEY,
        title TEXT NOT NULL,
        done BOOLEAN DEFAULT 0,
        created_at DATETIME NOT NULL,
        due_date DATETIME
    );`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) Create(task *models.Task) error {
	query := `INSERT INTO tasks (id, title, done, created_at, due_date) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, task.ID, task.Title, task.Done, task.CreatedAt, task.DueDate)
	return err
}

func (s *SQLiteStorage) GetAll() ([]*models.Task, error) {
	query := `SELECT id, title, done, created_at, due_date FROM tasks`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		var dueDate sql.NullTime

		err := rows.Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt, &dueDate)
		if err != nil {
			return nil, err
		}

		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *SQLiteStorage) GetByID(id string) (*models.Task, error) {
	query := `SELECT id, title, done, created_at, due_date FROM tasks WHERE id = ?`
	row := s.db.QueryRow(query, id)

	task := &models.Task{}
	var dueDate sql.NullTime

	err := row.Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt, &dueDate)
	if err != nil {
		return nil, err
	}

	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}

	return task, nil
}

func (s *SQLiteStorage) Update(task *models.Task) error {
	query := `UPDATE tasks SET title = ?, done = ?, due_date = ? WHERE id = ?`
	result, err := s.db.Exec(query, task.Title, task.Done, task.DueDate, task.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *SQLiteStorage) Delete(id string) error {
	query := `DELETE FROM tasks WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
