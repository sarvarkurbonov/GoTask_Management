package models

import (
	"time"
)

type Task struct {
	ID        string     `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Title     string     `json:"title" gorm:"not null;type:text"`
	Done      bool       `json:"done" gorm:"default:false"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	DueDate   *time.Time `json:"due_date,omitempty" gorm:"index"`
}

type TaskFilter struct {
	Status  string // "done", "undone", or empty for all
	DueDays int    // Number of days to look ahead for due tasks
}
