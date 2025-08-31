package storage

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// StorageType represents the type of storage backend
type StorageType string

const (
	StorageTypeJSON       StorageType = "json"
	StorageTypeSQLite     StorageType = "sqlite"
	StorageTypePostgreSQL StorageType = "postgres"
	StorageTypeMySQL      StorageType = "mysql"
	StorageTypeMongoDB    StorageType = "mongodb"
)

// StorageConfig holds configuration for all storage types
type StorageConfig struct {
	Type     StorageType
	FilePath string // For JSON and SQLite

	// Database connection settings
	Host     string
	Port     int
	User     string
	Password string
	DBName   string

	// PostgreSQL specific
	SSLMode  string
	TimeZone string

	// MySQL specific
	Charset   string
	ParseTime bool
	Loc       string

	// MongoDB specific
	URI            string
	Collection     string
	ConnectTimeout time.Duration
	QueryTimeout   time.Duration
}

// NewStorageFromEnv creates a storage instance based on environment variables
func NewStorageFromEnv() (Storage, error) {
	config, err := LoadConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from environment: %w", err)
	}

	return NewStorage(config)
}

// LoadConfigFromEnv loads storage configuration from environment variables
func LoadConfigFromEnv() (*StorageConfig, error) {
	storageType := StorageType(getEnvOrDefault("STORAGE_TYPE", "json"))

	config := &StorageConfig{
		Type:     storageType,
		FilePath: getEnvOrDefault("STORAGE_FILE_PATH", "tasks.json"),
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		User:     getEnvOrDefault("DB_USER", ""),
		Password: getEnvOrDefault("DB_PASSWORD", ""),
		DBName:   getEnvOrDefault("DB_NAME", "gotask"),
	}

	// Parse port
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_PORT: %w", err)
		}
		config.Port = port
	} else {
		// Set default ports based on storage type
		switch storageType {
		case StorageTypePostgreSQL:
			config.Port = 5432
		case StorageTypeMySQL:
			config.Port = 3306
		case StorageTypeMongoDB:
			config.Port = 27017
		}
	}

	// PostgreSQL specific settings
	config.SSLMode = getEnvOrDefault("POSTGRES_SSL_MODE", "disable")
	config.TimeZone = getEnvOrDefault("POSTGRES_TIMEZONE", "UTC")

	// MySQL specific settings
	config.Charset = getEnvOrDefault("MYSQL_CHARSET", "utf8mb4")
	config.ParseTime = getEnvOrDefault("MYSQL_PARSE_TIME", "true") == "true"
	config.Loc = getEnvOrDefault("MYSQL_LOC", "Local")

	// MongoDB specific settings
	config.URI = getEnvOrDefault("MONGODB_URI", fmt.Sprintf("mongodb://%s:%d", config.Host, config.Port))
	config.Collection = getEnvOrDefault("MONGODB_COLLECTION", "tasks")
	
	// Parse timeouts
	connectTimeoutStr := getEnvOrDefault("MONGODB_CONNECT_TIMEOUT", "10s")
	connectTimeout, err := time.ParseDuration(connectTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MONGODB_CONNECT_TIMEOUT: %w", err)
	}
	config.ConnectTimeout = connectTimeout

	queryTimeoutStr := getEnvOrDefault("MONGODB_QUERY_TIMEOUT", "5s")
	queryTimeout, err := time.ParseDuration(queryTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MONGODB_QUERY_TIMEOUT: %w", err)
	}
	config.QueryTimeout = queryTimeout

	return config, nil
}

// NewStorage creates a storage instance based on the provided configuration
func NewStorage(config *StorageConfig) (Storage, error) {
	switch config.Type {
	case StorageTypeJSON:
		return NewJSONStorage(config.FilePath)

	case StorageTypeSQLite:
		return NewSQLiteStorage(config.FilePath)

	case StorageTypePostgreSQL:
		pgConfig := PostgreSQLConfig{
			Host:     config.Host,
			Port:     config.Port,
			User:     config.User,
			Password: config.Password,
			DBName:   config.DBName,
			SSLMode:  config.SSLMode,
			TimeZone: config.TimeZone,
		}
		return NewPostgreSQLStorage(pgConfig)

	case StorageTypeMySQL:
		mysqlConfig := MySQLConfig{
			Host:      config.Host,
			Port:      config.Port,
			User:      config.User,
			Password:  config.Password,
			DBName:    config.DBName,
			Charset:   config.Charset,
			ParseTime: config.ParseTime,
			Loc:       config.Loc,
		}
		return NewMySQLStorage(mysqlConfig)

	case StorageTypeMongoDB:
		mongoConfig := MongoDBConfig{
			URI:            config.URI,
			Database:       config.DBName,
			Collection:     config.Collection,
			ConnectTimeout: config.ConnectTimeout,
			QueryTimeout:   config.QueryTimeout,
		}
		return NewMongoDBStorage(mongoConfig)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ValidateConfig validates the storage configuration
func ValidateConfig(config *StorageConfig) error {
	switch config.Type {
	case StorageTypeJSON, StorageTypeSQLite:
		if config.FilePath == "" {
			return fmt.Errorf("file path is required for %s storage", config.Type)
		}

	case StorageTypePostgreSQL, StorageTypeMySQL:
		if config.Host == "" {
			return fmt.Errorf("host is required for %s storage", config.Type)
		}
		if config.User == "" {
			return fmt.Errorf("user is required for %s storage", config.Type)
		}
		if config.DBName == "" {
			return fmt.Errorf("database name is required for %s storage", config.Type)
		}
		if config.Port <= 0 {
			return fmt.Errorf("valid port is required for %s storage", config.Type)
		}

	case StorageTypeMongoDB:
		if config.URI == "" {
			return fmt.Errorf("URI is required for MongoDB storage")
		}
		if config.DBName == "" {
			return fmt.Errorf("database name is required for MongoDB storage")
		}
		if config.Collection == "" {
			return fmt.Errorf("collection name is required for MongoDB storage")
		}

	default:
		return fmt.Errorf("unsupported storage type: %s", config.Type)
	}

	return nil
}

// GetSupportedStorageTypes returns a list of supported storage types
func GetSupportedStorageTypes() []StorageType {
	return []StorageType{
		StorageTypeJSON,
		StorageTypeSQLite,
		StorageTypePostgreSQL,
		StorageTypeMySQL,
		StorageTypeMongoDB,
	}
}

// IsStorageTypeSupported checks if a storage type is supported
func IsStorageTypeSupported(storageType StorageType) bool {
	supportedTypes := GetSupportedStorageTypes()
	for _, supported := range supportedTypes {
		if supported == storageType {
			return true
		}
	}
	return false
}
