package config

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

)

var DB *gorm.DB
var RedisClient *redis.Client

func InitDatabase() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	dsn := AppConfig.DatabaseURL

	// Ensure database exists before connecting
	if err := createDatabaseIfNotExist(dsn); err != nil {
		log.Printf("Warning: failed to check/create database: %v", err)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	log.Println("Database connection established")
}

func createDatabaseIfNotExist(dsn string) error {
	var dbname string
	var defaultDSN string

	// Support both Postgres URL and Key-Value formats
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return err
		}
		// u.Path contains "/" + dbname
		dbname = strings.TrimPrefix(u.Path, "/")
		if idx := strings.Index(dbname, "?"); idx != -1 {
			dbname = dbname[:idx]
		}
		// If dbname is empty or default, nothing to do
		if dbname == "" || dbname == "postgres" {
			return nil
		}
		u.Path = "/postgres"
		defaultDSN = u.String()
	} else {
		// Key-Value format: e.g. "host=localhost user=root dbname=starter_kit ..."
		fields := splitKV(dsn)
		var newFields []string
		for _, field := range fields {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if key == "dbname" {
					dbname = strings.Trim(val, `'"`)
					newFields = append(newFields, "dbname=postgres")
				} else {
					newFields = append(newFields, field)
				}
			} else {
				newFields = append(newFields, field)
			}
		}
		if dbname == "" || dbname == "postgres" {
			return nil
		}
		defaultDSN = strings.Join(newFields, " ")
	}

	// Connect to default 'postgres' database
	db, err := gorm.Open(postgres.Open(defaultDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to default postgres DB: %w", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	// Check if database exists
	var count int64
	err = db.Raw("SELECT count(*) FROM pg_database WHERE datname = ?", dbname).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it does not exist
	if count == 0 {
		log.Printf("Database '%s' not found. Creating database...", dbname)
		query := fmt.Sprintf(`CREATE DATABASE "%s"`, dbname)
		err = db.Exec(query).Error
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database '%s' created successfully.", dbname)
	}

	return nil
}

func splitKV(dsn string) []string {
	var fields []string
	var current strings.Builder
	inQuote := false
	var quoteChar rune

	runes := []rune(dsn)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inQuote {
			current.WriteRune(r)
			if r == quoteChar {
				inQuote = false
			}
		} else {
			if r == '\'' || r == '"' {
				inQuote = true
				quoteChar = r
				current.WriteRune(r)
			} else if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
				if current.Len() > 0 {
					fields = append(fields, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(r)
			}
		}
	}
	if current.Len() > 0 {
		fields = append(fields, current.String())
	}
	return fields
}

func InitRedis() {
	addr := fmt.Sprintf("%s:%s", AppConfig.RedisHost, AppConfig.RedisPort)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	RedisClient = client
	log.Println("Redis connection established")
}


