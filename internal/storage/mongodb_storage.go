package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"GoTask_Management/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBStorage implements the Storage interface using MongoDB
type MongoDBStorage struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	ctx        context.Context
}

// MongoDBConfig holds the configuration for MongoDB connection
type MongoDBConfig struct {
	URI            string
	Database       string
	Collection     string
	ConnectTimeout time.Duration
	QueryTimeout   time.Duration
}

// NewMongoDBStorage creates a new MongoDB storage instance
func NewMongoDBStorage(config MongoDBConfig) (*MongoDBStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(config.URI)
	clientOptions.SetMaxPoolSize(100)
	clientOptions.SetMinPoolSize(10)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)
	collection := database.Collection(config.Collection)

	storage := &MongoDBStorage{
		client:     client,
		database:   database,
		collection: collection,
		ctx:        context.Background(),
	}

	// Create indexes
	if err := storage.createIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("✅ MongoDB storage initialized successfully")
	return storage, nil
}

// createIndexes creates necessary indexes for optimal performance
func (ms *MongoDBStorage) createIndexes() error {
	ctx, cancel := context.WithTimeout(ms.ctx, 10*time.Second)
	defer cancel()

	// Create index on ID field
	idIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// Create index on created_at field
	createdAtIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "created_at", Value: -1}},
	}

	// Create index on due_date field
	dueDateIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "due_date", Value: 1}},
	}

	// Create index on done field
	doneIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "done", Value: 1}},
	}

	// Create compound index for filtering
	compoundIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "done", Value: 1},
			{Key: "due_date", Value: 1},
		},
	}

	indexes := []mongo.IndexModel{idIndex, createdAtIndex, dueDateIndex, doneIndex, compoundIndex}

	_, err := ms.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Create implements Storage interface
func (ms *MongoDBStorage) Create(task *models.Task) error {
	ctx, cancel := context.WithTimeout(ms.ctx, 5*time.Second)
	defer cancel()

	_, err := ms.collection.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// GetAll implements Storage interface
func (ms *MongoDBStorage) GetAll() ([]*models.Task, error) {
	ctx, cancel := context.WithTimeout(ms.ctx, 10*time.Second)
	defer cancel()

	// Sort by created_at in descending order
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := ms.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []*models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

// GetByID implements Storage interface
func (ms *MongoDBStorage) GetByID(id string) (*models.Task, error) {
	ctx, cancel := context.WithTimeout(ms.ctx, 5*time.Second)
	defer cancel()

	var task models.Task
	filter := bson.D{{Key: "id", Value: id}}
	
	err := ms.collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task by ID: %w", err)
	}

	return &task, nil
}

// Update implements Storage interface
func (ms *MongoDBStorage) Update(task *models.Task) error {
	ctx, cancel := context.WithTimeout(ms.ctx, 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "id", Value: task.ID}}
	update := bson.D{{Key: "$set", Value: task}}

	result, err := ms.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// Delete implements Storage interface
func (ms *MongoDBStorage) Delete(id string) error {
	ctx, cancel := context.WithTimeout(ms.ctx, 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "id", Value: id}}
	result, err := ms.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// Close implements Storage interface
func (ms *MongoDBStorage) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return ms.client.Disconnect(ctx)
}

// GetTasksByStatus returns tasks filtered by status
func (ms *MongoDBStorage) GetTasksByStatus(done bool) ([]*models.Task, error) {
	ctx, cancel := context.WithTimeout(ms.ctx, 10*time.Second)
	defer cancel()

	filter := bson.D{{Key: "done", Value: done}}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := ms.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by status: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []*models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

// GetTasksDueBefore returns tasks due before the specified time
func (ms *MongoDBStorage) GetTasksDueBefore(deadline time.Time) ([]*models.Task, error) {
	ctx, cancel := context.WithTimeout(ms.ctx, 10*time.Second)
	defer cancel()

	filter := bson.D{
		{Key: "due_date", Value: bson.D{{Key: "$ne", Value: nil}}},
		{Key: "due_date", Value: bson.D{{Key: "$lte", Value: deadline}}},
	}
	opts := options.Find().SetSort(bson.D{{Key: "due_date", Value: 1}})

	cursor, err := ms.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks due before deadline: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []*models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

// GetTasksCount returns the total count of tasks
func (ms *MongoDBStorage) GetTasksCount() (int64, error) {
	ctx, cancel := context.WithTimeout(ms.ctx, 5*time.Second)
	defer cancel()

	count, err := ms.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}
	return count, nil
}

// GetTasksCountByStatus returns the count of tasks by status
func (ms *MongoDBStorage) GetTasksCountByStatus(done bool) (int64, error) {
	ctx, cancel := context.WithTimeout(ms.ctx, 5*time.Second)
	defer cancel()

	filter := bson.D{{Key: "done", Value: done}}
	count, err := ms.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks by status: %w", err)
	}
	return count, nil
}

// GetOverdueTasksCount returns the count of overdue tasks
func (ms *MongoDBStorage) GetOverdueTasksCount() (int64, error) {
	ctx, cancel := context.WithTimeout(ms.ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	filter := bson.D{
		{Key: "due_date", Value: bson.D{{Key: "$ne", Value: nil}}},
		{Key: "due_date", Value: bson.D{{Key: "$lt", Value: now}}},
		{Key: "done", Value: false},
	}

	count, err := ms.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count overdue tasks: %w", err)
	}
	return count, nil
}

// HealthCheck performs a health check on the database connection
func (ms *MongoDBStorage) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ms.client.Ping(ctx, nil)
}

// GetCollection returns the underlying MongoDB collection (for advanced operations)
func (ms *MongoDBStorage) GetCollection() *mongo.Collection {
	return ms.collection
}

// Verify that MongoDBStorage implements Storage interface
var _ Storage = (*MongoDBStorage)(nil)
