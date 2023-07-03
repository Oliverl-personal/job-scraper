package utils

import (
	"fmt"
	"os"
	"strconv"

	ms "scraper-app/model/storage"

	"github.com/joho/godotenv"
)

func LoadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		Logger.Warn(".env file is missing - not loaded")
	} else {
		Logger.Debug(".env loaded")
	}
}

func GetEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	} else {
		FatalError(fmt.Errorf("env variable %s is required", key))
		return ""
	}
}

func GetEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetIntEnvOrDefault(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if parse, err := strconv.Atoi(value); err == nil {
			return parse
		}
	}
	return fallback
}

func GetBoolEnvOrDefault(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if parse, err := strconv.ParseBool(value); err == nil {
			return parse
		}
	}
	return fallback
}

func GetProviderType() ms.ProviderType {
	switch GetEnv("STORAGE_PROVIDER") {
	case string(ms.AwsS3Bucket):
		return ms.AwsS3Bucket
	case string(ms.MariaDb):
		return ms.MariaDb
	case string(ms.Mysql):
		return ms.Mysql
	case string(ms.Postgres):
		return ms.Postgres
	default:
		return ms.InMemory
	}
}
