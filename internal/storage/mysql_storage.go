package storage

import (
	"fmt"
	"log"
	"time"

	"GoTask_Management/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLStorage implements the Storage interface using MySQL with GORM
type MySQLStorage struct {
	db *gorm.DB
}

// MySQLConfig holds the configuration for MySQL connection
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Charset  string
	ParseTime bool
	Loc      string
}

// NewMySQLStorage creates a new MySQL storage instance
func NewMySQLStorage(config MySQLConfig) (*MySQLStorage, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		config.User, config.Password, config.Host, config.Port, config.DBName,
		config.Charset, config.ParseTime, config.Loc)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	storage := &MySQLStorage{db: db}

	// Auto-migrate the schema
	if err := storage.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("✅ MySQL storage initialized successfully")
	return storage, nil
}

// migrate runs database migrations
func (ms *MySQLStorage) migrate() error {
	return ms.db.AutoMigrate(&models.Task{})
}

// Create implements Storage interface
func (ms *MySQLStorage) Create(task *models.Task) error {
	if err := ms.db.Create(task).Error; err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// GetAll implements Storage interface
func (ms *MySQLStorage) GetAll() ([]*models.Task, error) {
	var tasks []*models.Task
	if err := ms.db.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get all tasks: %w", err)
	}
	return tasks, nil
}

// GetByID implements Storage interface
func (ms *MySQLStorage) GetByID(id string) (*models.Task, error) {
	var task models.Task
	if err := ms.db.First(&task, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task by ID: %w", err)
	}
	return &task, nil
}

// Update implements Storage interface
func (ms *MySQLStorage) Update(task *models.Task) error {
	result := ms.db.Save(task)
	if result.Error != nil {
		return fmt.Errorf("failed to update task: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// Delete implements Storage interface
func (ms *MySQLStorage) Delete(id string) error {
	result := ms.db.Delete(&models.Task{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete task: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// Close implements Storage interface
func (ms *MySQLStorage) Close() error {
	sqlDB, err := ms.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// GetTasksByStatus returns tasks filtered by status
func (ms *MySQLStorage) GetTasksByStatus(done bool) ([]*models.Task, error) {
	var tasks []*models.Task
	if err := ms.db.Where("done = ?", done).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tasks by status: %w", err)
	}
	return tasks, nil
}

// GetTasksDueBefore returns tasks due before the specified time
func (ms *MySQLStorage) GetTasksDueBefore(deadline time.Time) ([]*models.Task, error) {
	var tasks []*models.Task
	if err := ms.db.Where("due_date IS NOT NULL AND due_date <= ?", deadline).
		Order("due_date ASC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tasks due before deadline: %w", err)
	}
	return tasks, nil
}

// GetTasksCount returns the total count of tasks
func (ms *MySQLStorage) GetTasksCount() (int64, error) {
	var count int64
	if err := ms.db.Model(&models.Task{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}
	return count, nil
}

// GetTasksCountByStatus returns the count of tasks by status
func (ms *MySQLStorage) GetTasksCountByStatus(done bool) (int64, error) {
	var count int64
	if err := ms.db.Model(&models.Task{}).Where("done = ?", done).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count tasks by status: %w", err)
	}
	return count, nil
}

// GetOverdueTasksCount returns the count of overdue tasks
func (ms *MySQLStorage) GetOverdueTasksCount() (int64, error) {
	var count int64
	now := time.Now()
	if err := ms.db.Model(&models.Task{}).
		Where("due_date IS NOT NULL AND due_date < ? AND done = false", now).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count overdue tasks: %w", err)
	}
	return count, nil
}

// HealthCheck performs a health check on the database connection
func (ms *MySQLStorage) HealthCheck() error {
	sqlDB, err := ms.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Ping()
}

// GetDB returns the underlying GORM database instance (for advanced operations)
func (ms *MySQLStorage) GetDB() *gorm.DB {
	return ms.db
}

// BeginTransaction starts a new database transaction
func (ms *MySQLStorage) BeginTransaction() *gorm.DB {
	return ms.db.Begin()
}

// Verify that MySQLStorage implements Storage interface
var _ Storage = (*MySQLStorage)(nil)
