package main

import (
	"fmt"
	"log"
	"os"

	"time"

	"GoTask_Management/internal/storage"
	"GoTask_Management/internal/task"

	"github.com/spf13/cobra"
)

var (
	taskService *task.Service
	rootCmd     = &cobra.Command{
		Use:   "gotasker",
		Short: "A task management CLI tool",
		Long:  `GoTasker is a simple and efficient task management tool for personal and team use.`,
	}
)

func init() {
	// Initialize storage
	store, err := storage.NewJSONStorage("tasks.json")
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}
	taskService = task.NewService(store)

	// Add commands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(dueCmd)
}

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		dueDateStr, _ := cmd.Flags().GetString("due")

		var dueDate *time.Time
		if dueDateStr != "" {
			parsed, err := time.Parse("2006-01-02", dueDateStr)
			if err != nil {
				fmt.Printf("Error parsing date: %v\n", err)
				return
			}
			dueDate = &parsed
		}

		task, err := taskService.CreateTask(title, dueDate)
		if err != nil {
			fmt.Printf("Error creating task: %v\n", err)
			return
		}
		fmt.Printf("Task created successfully: [%s] %s\n", task.ID, task.Title)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Run: func(cmd *cobra.Command, args []string) {
		statusFilter, _ := cmd.Flags().GetString("status")

		tasks, err := taskService.ListTasks(statusFilter)
		if err != nil {
			fmt.Printf("Error listing tasks: %v\n", err)
			return
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return
		}

		fmt.Println("\n📋 Tasks:")
		fmt.Println("─────────────────────────────────────────")
		for _, t := range tasks {
			status := "⬜"
			if t.Done {
				status = "✅"
			}

			dueStr := ""
			if t.DueDate != nil {
				dueStr = fmt.Sprintf(" (Due: %s)", t.DueDate.Format("2006-01-02"))
			}

			fmt.Printf("%s [%s] %s%s\n", status, t.ID, t.Title, dueStr)
		}
		fmt.Println("─────────────────────────────────────────")
	},
}

var doneCmd = &cobra.Command{
	Use:   "done [id]",
	Short: "Mark a task as done",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		err := taskService.MarkTaskDone(id, true)
		if err != nil {
			fmt.Printf("Error marking task as done: %v\n", err)
			return
		}
		fmt.Printf("Task %s marked as done ✅\n", id)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		err := taskService.DeleteTask(id)
		if err != nil {
			fmt.Printf("Error deleting task: %v\n", err)
			return
		}
		fmt.Printf("Task %s deleted successfully 🗑️\n", id)
	},
}

var dueCmd = &cobra.Command{
	Use:   "due",
	Short: "List tasks due today or within specified days",
	Run: func(cmd *cobra.Command, args []string) {
		days, _ := cmd.Flags().GetInt("days")

		tasks, err := taskService.GetDueTasks(days)
		if err != nil {
			fmt.Printf("Error getting due tasks: %v\n", err)
			return
		}

		if len(tasks) == 0 {
			fmt.Printf("No tasks due in the next %d days.\n", days)
			return
		}

		fmt.Printf("\n⏰ Tasks due in the next %d days:\n", days)
		fmt.Println("─────────────────────────────────────────")
		for _, t := range tasks {
			status := "⬜"
			if t.Done {
				status = "✅"
			}

			dueStr := ""
			if t.DueDate != nil {
				dueStr = t.DueDate.Format("2006-01-02")
			}

			fmt.Printf("%s [%s] %s - Due: %s\n", status, t.ID, t.Title, dueStr)
		}
		fmt.Println("─────────────────────────────────────────")
	},
}

func init() {
	addCmd.Flags().StringP("due", "d", "", "Due date (YYYY-MM-DD)")
	listCmd.Flags().StringP("status", "s", "", "Filter by status (done/undone)")
	dueCmd.Flags().IntP("days", "d", 7, "Number of days to look ahead")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
