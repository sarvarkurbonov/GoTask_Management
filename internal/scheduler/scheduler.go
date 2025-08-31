package scheduler

import (
	"log"
	"time"

	"GoTask_Management/internal/task"
)

type Scheduler struct {
	taskService *task.Service
	interval    int // in seconds
	ticker      *time.Ticker
	done        chan bool
}

func New(taskService *task.Service, interval int) *Scheduler {
	return &Scheduler{
		taskService: taskService,
		interval:    interval,
		done:        make(chan bool),
	}
}

func (s *Scheduler) Start() {
	s.ticker = time.NewTicker(time.Duration(s.interval) * time.Second)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.performBackup()
			case <-s.done:
				return
			}
		}
	}()

	log.Printf("Scheduler started with interval: %d seconds", s.interval)
}

func (s *Scheduler) Stop() {
	s.ticker.Stop()
	s.done <- true
	log.Println("Scheduler stopped")
}

func (s *Scheduler) performBackup() {
	total, done, overdue, err := s.taskService.GetTasksSummary()
	if err != nil {
		log.Printf("Error getting task summary: %v", err)
		return
	}

	log.Printf("📊 Task Summary - Total: %d | Done: %d | Overdue: %d", total, done, overdue)

	// Here you could add additional backup logic
	// For example, creating a backup file with timestamp
	// or sending data to a remote backup service
}
