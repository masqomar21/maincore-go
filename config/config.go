package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName          string
	AppVersion       string
	AppEnv           string
	Port             string
	DatabaseURL      string
	RedisHost        string
	RedisPort        string
	JWTSecret        string
	S3Endpoint       string
	S3Bucket         string
	S3Region         string
	S3AccessKeyID    string
	S3SecretAccessKey string
	S3ForcePathStyle bool
	FileSaveToBucket bool
}

var AppConfig Config

func InitConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	AppConfig = Config{
		AppName:          getEnv("APP_NAME", "Starter Kit"),
		AppVersion:       getEnv("APP_VERSION", "1.0.0"),
		AppEnv:           getEnv("APP_ENV", "development"),
		Port:             getEnv("PORT", "3000"),
		DatabaseURL:      getEnv("DATABASE_URL", "host=localhost user=root password=root dbname=starter_kit port=5432 sslmode=disable TimeZone=Asia/Jakarta"),
		RedisHost:        getEnv("REDIS_HOST", "localhost"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
		JWTSecret:        getEnv("AUTH_JWT_SECRET", "secret"),
		S3Endpoint:       getEnv("S3_ENDPOINT", ""),
		S3Bucket:         getEnv("S3_BUCKET", ""),
		S3Region:         getEnv("S3_REGION", "ap-southeast-1"),
		S3AccessKeyID:    getEnv("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", ""),
		S3ForcePathStyle: getEnvBool("S3_FORCE_PATH_STYLE", false),
		FileSaveToBucket: getEnvBool("FILE_SAVE_TO_BUCKET", true),
	}
}

func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return fallback
}
