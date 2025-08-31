package storage

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"STORAGE_TYPE", "STORAGE_FILE_PATH", "DB_HOST", "DB_PORT", "DB_USER", 
		"DB_PASSWORD", "DB_NAME", "POSTGRES_SSL_MODE", "POSTGRES_TIMEZONE",
		"MYSQL_CHARSET", "MYSQL_PARSE_TIME", "MYSQL_LOC", "MONGODB_URI",
		"MONGODB_COLLECTION", "MONGODB_CONNECT_TIMEOUT", "MONGODB_QUERY_TIMEOUT",
	}
	
	for _, envVar := range envVars {
		originalEnv[envVar] = os.Getenv(envVar)
	}
	
	// Clean environment
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
	
	// Restore environment after test
	defer func() {
		for _, envVar := range envVars {
			if value, exists := originalEnv[envVar]; exists && value != "" {
				os.Setenv(envVar, value)
			} else {
				os.Unsetenv(envVar)
			}
		}
	}()

	t.Run("DefaultValues", func(t *testing.T) {
		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if config.Type != StorageTypeJSON {
			t.Errorf("Expected default storage type %s, got %s", StorageTypeJSON, config.Type)
		}
		if config.FilePath != "tasks.json" {
			t.Errorf("Expected default file path 'tasks.json', got %s", config.FilePath)
		}
		if config.Host != "localhost" {
			t.Errorf("Expected default host 'localhost', got %s", config.Host)
		}
	})

	t.Run("PostgreSQLConfig", func(t *testing.T) {
		os.Setenv("STORAGE_TYPE", "postgres")
		os.Setenv("DB_HOST", "pg-host")
		os.Setenv("DB_PORT", "5433")
		os.Setenv("DB_USER", "pguser")
		os.Setenv("DB_PASSWORD", "pgpass")
		os.Setenv("DB_NAME", "pgdb")
		os.Setenv("POSTGRES_SSL_MODE", "require")
		os.Setenv("POSTGRES_TIMEZONE", "America/New_York")

		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("Failed to load PostgreSQL config: %v", err)
		}

		if config.Type != StorageTypePostgreSQL {
			t.Errorf("Expected PostgreSQL storage type, got %s", config.Type)
		}
		if config.Host != "pg-host" {
			t.Errorf("Expected host 'pg-host', got %s", config.Host)
		}
		if config.Port != 5433 {
			t.Errorf("Expected port 5433, got %d", config.Port)
		}
		if config.User != "pguser" {
			t.Errorf("Expected user 'pguser', got %s", config.User)
		}
		if config.Password != "pgpass" {
			t.Errorf("Expected password 'pgpass', got %s", config.Password)
		}
		if config.DBName != "pgdb" {
			t.Errorf("Expected database 'pgdb', got %s", config.DBName)
		}
		if config.SSLMode != "require" {
			t.Errorf("Expected SSL mode 'require', got %s", config.SSLMode)
		}
		if config.TimeZone != "America/New_York" {
			t.Errorf("Expected timezone 'America/New_York', got %s", config.TimeZone)
		}
	})

	t.Run("MySQLConfig", func(t *testing.T) {
		// Clean previous env vars
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		os.Setenv("STORAGE_TYPE", "mysql")
		os.Setenv("DB_HOST", "mysql-host")
		os.Setenv("DB_PORT", "3307")
		os.Setenv("DB_USER", "mysqluser")
		os.Setenv("DB_PASSWORD", "mysqlpass")
		os.Setenv("DB_NAME", "mysqldb")
		os.Setenv("MYSQL_CHARSET", "utf8")
		os.Setenv("MYSQL_PARSE_TIME", "false")
		os.Setenv("MYSQL_LOC", "UTC")

		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("Failed to load MySQL config: %v", err)
		}

		if config.Type != StorageTypeMySQL {
			t.Errorf("Expected MySQL storage type, got %s", config.Type)
		}
		if config.Host != "mysql-host" {
			t.Errorf("Expected host 'mysql-host', got %s", config.Host)
		}
		if config.Port != 3307 {
			t.Errorf("Expected port 3307, got %d", config.Port)
		}
		if config.Charset != "utf8" {
			t.Errorf("Expected charset 'utf8', got %s", config.Charset)
		}
		if config.ParseTime != false {
			t.Errorf("Expected ParseTime false, got %v", config.ParseTime)
		}
		if config.Loc != "UTC" {
			t.Errorf("Expected location 'UTC', got %s", config.Loc)
		}
	})

	t.Run("MongoDBConfig", func(t *testing.T) {
		// Clean previous env vars
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		os.Setenv("STORAGE_TYPE", "mongodb")
		os.Setenv("DB_HOST", "mongo-host")
		os.Setenv("DB_PORT", "27018")
		os.Setenv("DB_NAME", "mongodb")
		os.Setenv("MONGODB_URI", "mongodb://mongo-host:27018")
		os.Setenv("MONGODB_COLLECTION", "custom_tasks")
		os.Setenv("MONGODB_CONNECT_TIMEOUT", "15s")
		os.Setenv("MONGODB_QUERY_TIMEOUT", "10s")

		config, err := LoadConfigFromEnv()
		if err != nil {
			t.Fatalf("Failed to load MongoDB config: %v", err)
		}

		if config.Type != StorageTypeMongoDB {
			t.Errorf("Expected MongoDB storage type, got %s", config.Type)
		}
		if config.URI != "mongodb://mongo-host:27018" {
			t.Errorf("Expected URI 'mongodb://mongo-host:27018', got %s", config.URI)
		}
		if config.Collection != "custom_tasks" {
			t.Errorf("Expected collection 'custom_tasks', got %s", config.Collection)
		}
		if config.ConnectTimeout != 15*time.Second {
			t.Errorf("Expected connect timeout 15s, got %v", config.ConnectTimeout)
		}
		if config.QueryTimeout != 10*time.Second {
			t.Errorf("Expected query timeout 10s, got %v", config.QueryTimeout)
		}
	})

	t.Run("InvalidPort", func(t *testing.T) {
		// Clean previous env vars
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		os.Setenv("DB_PORT", "invalid")

		_, err := LoadConfigFromEnv()
		if err == nil {
			t.Error("Expected error for invalid port, got nil")
		}
	})

	t.Run("InvalidTimeout", func(t *testing.T) {
		// Clean previous env vars
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		os.Setenv("STORAGE_TYPE", "mongodb")
		os.Setenv("MONGODB_CONNECT_TIMEOUT", "invalid")

		_, err := LoadConfigFromEnv()
		if err == nil {
			t.Error("Expected error for invalid timeout, got nil")
		}
	})
}

func TestNewStorage(t *testing.T) {
	t.Run("JSONStorage", func(t *testing.T) {
		config := &StorageConfig{
			Type:     StorageTypeJSON,
			FilePath: "test_tasks.json",
		}

		storage, err := NewStorage(config)
		if err != nil {
			t.Fatalf("Failed to create JSON storage: %v", err)
		}
		defer storage.Close()

		if storage == nil {
			t.Error("Expected storage instance, got nil")
		}

		// Clean up test file
		os.Remove("test_tasks.json")
	})

	t.Run("SQLiteStorage", func(t *testing.T) {
		config := &StorageConfig{
			Type:     StorageTypeSQLite,
			FilePath: "test_tasks.db",
		}

		storage, err := NewStorage(config)
		if err != nil {
			t.Fatalf("Failed to create SQLite storage: %v", err)
		}
		defer storage.Close()

		if storage == nil {
			t.Error("Expected storage instance, got nil")
		}

		// Clean up test file
		os.Remove("test_tasks.db")
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		config := &StorageConfig{
			Type: StorageType("unsupported"),
		}

		_, err := NewStorage(config)
		if err == nil {
			t.Error("Expected error for unsupported storage type, got nil")
		}
	})
}

func TestValidateConfig(t *testing.T) {
	t.Run("ValidJSONConfig", func(t *testing.T) {
		config := &StorageConfig{
			Type:     StorageTypeJSON,
			FilePath: "tasks.json",
		}

		err := ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("InvalidJSONConfig", func(t *testing.T) {
		config := &StorageConfig{
			Type:     StorageTypeJSON,
			FilePath: "",
		}

		err := ValidateConfig(config)
		if err == nil {
			t.Error("Expected error for missing file path, got nil")
		}
	})

	t.Run("ValidPostgreSQLConfig", func(t *testing.T) {
		config := &StorageConfig{
			Type:     StorageTypePostgreSQL,
			Host:     "localhost",
			Port:     5432,
			User:     "user",
			Password: "password",
			DBName:   "database",
		}

		err := ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("InvalidPostgreSQLConfig", func(t *testing.T) {
		testCases := []struct {
			name   string
			config *StorageConfig
		}{
			{
				name: "MissingHost",
				config: &StorageConfig{
					Type:   StorageTypePostgreSQL,
					Port:   5432,
					User:   "user",
					DBName: "database",
				},
			},
			{
				name: "MissingUser",
				config: &StorageConfig{
					Type:   StorageTypePostgreSQL,
					Host:   "localhost",
					Port:   5432,
					DBName: "database",
				},
			},
			{
				name: "MissingDBName",
				config: &StorageConfig{
					Type: StorageTypePostgreSQL,
					Host: "localhost",
					Port: 5432,
					User: "user",
				},
			},
			{
				name: "InvalidPort",
				config: &StorageConfig{
					Type:   StorageTypePostgreSQL,
					Host:   "localhost",
					Port:   0,
					User:   "user",
					DBName: "database",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := ValidateConfig(tc.config)
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tc.name)
				}
			})
		}
	})

	t.Run("ValidMongoDBConfig", func(t *testing.T) {
		config := &StorageConfig{
			Type:       StorageTypeMongoDB,
			URI:        "mongodb://localhost:27017",
			DBName:     "database",
			Collection: "tasks",
		}

		err := ValidateConfig(config)
		if err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		config := &StorageConfig{
			Type: StorageType("unsupported"),
		}

		err := ValidateConfig(config)
		if err == nil {
			t.Error("Expected error for unsupported type, got nil")
		}
	})
}

func TestGetSupportedStorageTypes(t *testing.T) {
	types := GetSupportedStorageTypes()
	
	expectedTypes := []StorageType{
		StorageTypeJSON,
		StorageTypeSQLite,
		StorageTypePostgreSQL,
		StorageTypeMySQL,
		StorageTypeMongoDB,
	}

	if len(types) != len(expectedTypes) {
		t.Errorf("Expected %d storage types, got %d", len(expectedTypes), len(types))
	}

	for _, expectedType := range expectedTypes {
		found := false
		for _, actualType := range types {
			if actualType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected storage type %s not found", expectedType)
		}
	}
}

func TestIsStorageTypeSupported(t *testing.T) {
	supportedTypes := []StorageType{
		StorageTypeJSON,
		StorageTypeSQLite,
		StorageTypePostgreSQL,
		StorageTypeMySQL,
		StorageTypeMongoDB,
	}

	for _, storageType := range supportedTypes {
		if !IsStorageTypeSupported(storageType) {
			t.Errorf("Storage type %s should be supported", storageType)
		}
	}

	unsupportedTypes := []StorageType{
		StorageType("redis"),
		StorageType("cassandra"),
		StorageType("unknown"),
	}

	for _, storageType := range unsupportedTypes {
		if IsStorageTypeSupported(storageType) {
			t.Errorf("Storage type %s should not be supported", storageType)
		}
	}
}
