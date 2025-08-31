package storage

import (
	"fmt"
	"log"
	"time"

	"GoTask_Management/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgreSQLStorage implements the Storage interface using PostgreSQL with GORM
type PostgreSQLStorage struct {
	db *gorm.DB
}

// PostgreSQLConfig holds the configuration for PostgreSQL connection
type PostgreSQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// NewPostgreSQLStorage creates a new PostgreSQL storage instance
func NewPostgreSQLStorage(config PostgreSQLConfig) (*PostgreSQLStorage, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode, config.TimeZone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	storage := &PostgreSQLStorage{db: db}

	// Auto-migrate the schema
	if err := storage.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("✅ PostgreSQL storage initialized successfully")
	return storage, nil
}

// migrate runs database migrations
func (ps *PostgreSQLStorage) migrate() error {
	return ps.db.AutoMigrate(&models.Task{})
}

// Create implements Storage interface
func (ps *PostgreSQLStorage) Create(task *models.Task) error {
	if err := ps.db.Create(task).Error; err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// GetAll implements Storage interface
func (ps *PostgreSQLStorage) GetAll() ([]*models.Task, error) {
	var tasks []*models.Task
	if err := ps.db.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get all tasks: %w", err)
	}
	return tasks, nil
}

// GetByID implements Storage interface
func (ps *PostgreSQLStorage) GetByID(id string) (*models.Task, error) {
	var task models.Task
	if err := ps.db.First(&task, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task by ID: %w", err)
	}
	return &task, nil
}

// Update implements Storage interface
func (ps *PostgreSQLStorage) Update(task *models.Task) error {
	result := ps.db.Save(task)
	if result.Error != nil {
		return fmt.Errorf("failed to update task: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// Delete implements Storage interface
func (ps *PostgreSQLStorage) Delete(id string) error {
	result := ps.db.Delete(&models.Task{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete task: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// Close implements Storage interface
func (ps *PostgreSQLStorage) Close() error {
	sqlDB, err := ps.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// GetTasksByStatus returns tasks filtered by status
func (ps *PostgreSQLStorage) GetTasksByStatus(done bool) ([]*models.Task, error) {
	var tasks []*models.Task
	if err := ps.db.Where("done = ?", done).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tasks by status: %w", err)
	}
	return tasks, nil
}

// GetTasksDueBefore returns tasks due before the specified time
func (ps *PostgreSQLStorage) GetTasksDueBefore(deadline time.Time) ([]*models.Task, error) {
	var tasks []*models.Task
	if err := ps.db.Where("due_date IS NOT NULL AND due_date <= ?", deadline).
		Order("due_date ASC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tasks due before deadline: %w", err)
	}
	return tasks, nil
}

// GetTasksCount returns the total count of tasks
func (ps *PostgreSQLStorage) GetTasksCount() (int64, error) {
	var count int64
	if err := ps.db.Model(&models.Task{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}
	return count, nil
}

// GetTasksCountByStatus returns the count of tasks by status
func (ps *PostgreSQLStorage) GetTasksCountByStatus(done bool) (int64, error) {
	var count int64
	if err := ps.db.Model(&models.Task{}).Where("done = ?", done).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count tasks by status: %w", err)
	}
	return count, nil
}

// GetOverdueTasksCount returns the count of overdue tasks
func (ps *PostgreSQLStorage) GetOverdueTasksCount() (int64, error) {
	var count int64
	now := time.Now()
	if err := ps.db.Model(&models.Task{}).
		Where("due_date IS NOT NULL AND due_date < ? AND done = false", now).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count overdue tasks: %w", err)
	}
	return count, nil
}

// HealthCheck performs a health check on the database connection
func (ps *PostgreSQLStorage) HealthCheck() error {
	sqlDB, err := ps.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Ping()
}

// GetDB returns the underlying GORM database instance (for advanced operations)
func (ps *PostgreSQLStorage) GetDB() *gorm.DB {
	return ps.db
}

// BeginTransaction starts a new database transaction
func (ps *PostgreSQLStorage) BeginTransaction() *gorm.DB {
	return ps.db.Begin()
}

// Verify that PostgreSQLStorage implements Storage interface
var _ Storage = (*PostgreSQLStorage)(nil)
