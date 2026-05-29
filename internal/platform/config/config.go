package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App    AppConfig
	HTTP   HTTPConfig
	Log    LogConfig
	Redis  RedisConfig
	MySQL  MySQLConfig
	SQLite SQLiteConfig
	Mongo  MongoConfig
	Task   TaskConfig
	LLM    LLMConfig
	JWT    JWTConfig
	CORS   CORSConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Port    string
	GinMode string
}

type LogConfig struct {
	Level         slog.Level
	Format        string
	Dir           string
	FilePrefix    string
	Stdout        bool
	RetentionDays int
}

type RedisConfig struct {
	Enabled  bool
	Mode     string
	Addrs    []string
	Username string
	Password string
	DB       int
	Timeout  time.Duration
}

type MySQLConfig struct {
	Enabled bool
	DSN     string
	Timeout time.Duration
}

type SQLiteConfig struct {
	Enabled bool
	DSN     string
	Timeout time.Duration
}

type MongoConfig struct {
	Enabled  bool
	URI      string
	Database string
	Timeout  time.Duration
}

type TaskConfig struct {
	Enabled           bool
	StreamKey         string
	ConsumerGroup     string
	ConsumerName      string
	ReadBlock         time.Duration
	MaxLen            int64
	WorkerConcurrency int
	RetryJobs         bool
	MaxTries          int
	JobTimeout        time.Duration
}

type LLMConfig struct {
	Enabled     bool
	Provider    string
	BaseURL     string
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
	Timeout     time.Duration
}

type JWTConfig struct {
	Secret       string
	ExcludePaths []string
}

type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
	MaxAge       int
}

// Load 是配置唯一真源；启动期校验失败必须立即返回错误。
func Load() (Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			return Config{}, fmt.Errorf("load .env: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("stat .env: %w", err)
	}

	cfg := Config{
		App: AppConfig{
			Name: envString("APP_NAME", ""),
			Env:  envString("APP_ENV", "local"),
		},
		HTTP: HTTPConfig{
			Port:    envString("HTTP_PORT", ""),
			GinMode: envString("HTTP_GIN_MODE", "debug"),
		},
		Redis: RedisConfig{
			Enabled:  envBool("REDIS_ENABLED", false),
			Mode:     envString("REDIS_MODE", "single"),
			Addrs:    splitCSV(envString("REDIS_ADDRS", "")),
			Username: envString("REDIS_USERNAME", ""),
			Password: envString("REDIS_PASSWORD", ""),
			DB:       envInt("REDIS_DB", 0),
			Timeout:  envDuration("REDIS_TIMEOUT", 5*time.Second),
		},
		MySQL: MySQLConfig{
			Enabled: envBool("MYSQL_ENABLED", false),
			DSN:     envString("MYSQL_DSN", ""),
			Timeout: envDuration("MYSQL_TIMEOUT", 5*time.Second),
		},
		SQLite: SQLiteConfig{
			Enabled: envBool("SQLITE_ENABLED", false),
			DSN:     envString("SQLITE_DSN", ""),
			Timeout: envDuration("SQLITE_TIMEOUT", 5*time.Second),
		},
		Mongo: MongoConfig{
			Enabled:  envBool("MONGO_ENABLED", false),
			URI:      envString("MONGO_URI", ""),
			Database: envString("MONGO_DATABASE", ""),
			Timeout:  envDuration("MONGO_TIMEOUT", 10*time.Second),
		},
		Task: TaskConfig{
			Enabled:           envBool("TASK_ENABLED", false),
			StreamKey:         envString("TASK_STREAM_KEY", ""),
			ConsumerGroup:     envString("TASK_CONSUMER_GROUP", ""),
			ConsumerName:      envString("TASK_CONSUMER_NAME", ""),
			ReadBlock:         envDuration("TASK_READ_BLOCK", 3*time.Second),
			MaxLen:            envInt64("TASK_MAX_LEN", 10000),
			WorkerConcurrency: envInt("TASK_WORKER_CONCURRENCY", 2),
			RetryJobs:         envBool("TASK_RETRY_JOBS", true),
			MaxTries:          envInt("TASK_MAX_TRIES", 3),
			JobTimeout:        envDuration("TASK_JOB_TIMEOUT", time.Hour),
		},
		LLM: LLMConfig{
			Enabled:     envBool("LLM_ENABLED", false),
			Provider:    envString("LLM_PROVIDER", "openai"),
			BaseURL:     envString("LLM_BASE_URL", ""),
			APIKey:      envString("LLM_API_KEY", ""),
			Model:       envString("LLM_MODEL", ""),
			Temperature: envFloat("LLM_TEMPERATURE", 0.2),
			MaxTokens:   envInt("LLM_MAX_TOKENS", 1024),
			Timeout:     envDuration("LLM_TIMEOUT", 60*time.Second),
		},
		JWT: JWTConfig{
			Secret:       envString("JWT_SECRET", ""),
			ExcludePaths: splitCSV(envString("JWT_EXCLUDE_PATHS", "")),
		},
		CORS: CORSConfig{
			AllowOrigins: splitCSV(envString("CORS_ALLOW_ORIGINS", "")),
			AllowMethods: splitCSV(envString("CORS_ALLOW_METHODS", "")),
			AllowHeaders: splitCSV(envString("CORS_ALLOW_HEADERS", "")),
			MaxAge:       envInt("CORS_MAX_AGE", 600),
		},
	}

	logLevel, err := parseLogLevel(envString("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}
	logFormat, err := parseLogFormat(envString("LOG_FORMAT", "text"))
	if err != nil {
		return Config{}, err
	}
	cfg.Log = LogConfig{
		Level:         logLevel,
		Format:        logFormat,
		Dir:           envString("LOG_DIR", "var/log/anvil"),
		FilePrefix:    envString("LOG_FILE_PREFIX", "app"),
		Stdout:        envBool("LOG_STDOUT", true),
		RetentionDays: envInt("LOG_RETENTION_DAYS", 7),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

// Validate 在启动期集中校验，避免运行时才暴露配置缺失。
func (c Config) Validate() error {
	if strings.TrimSpace(c.App.Name) == "" {
		return fmt.Errorf("APP_NAME is required")
	}
	if strings.TrimSpace(c.HTTP.Port) == "" {
		return fmt.Errorf("HTTP_PORT is required")
	}
	if c.Log.Dir == "" {
		return fmt.Errorf("LOG_DIR is required")
	}

	if c.Task.Enabled {
		if strings.TrimSpace(c.Task.StreamKey) == "" {
			return fmt.Errorf("TASK_STREAM_KEY is required when task is enabled")
		}
		if strings.TrimSpace(c.Task.ConsumerGroup) == "" {
			return fmt.Errorf("TASK_CONSUMER_GROUP is required when task is enabled")
		}
		if strings.TrimSpace(c.Task.ConsumerName) == "" {
			return fmt.Errorf("TASK_CONSUMER_NAME is required when task is enabled")
		}
		if !c.Redis.Enabled {
			return fmt.Errorf("REDIS_ENABLED must be true when task is enabled")
		}
		if !c.Mongo.Enabled {
			return fmt.Errorf("MONGO_ENABLED must be true when task is enabled")
		}
	}

	if c.Redis.Enabled {
		if len(c.Redis.Addrs) == 0 {
			return fmt.Errorf("REDIS_ADDRS is required when redis is enabled")
		}
		if c.Redis.Mode != "single" && c.Redis.Mode != "cluster" {
			return fmt.Errorf("invalid REDIS_MODE: %s", c.Redis.Mode)
		}
	}

	if c.Mongo.Enabled {
		if strings.TrimSpace(c.Mongo.URI) == "" {
			return fmt.Errorf("MONGO_URI is required when mongo is enabled")
		}
		if strings.TrimSpace(c.Mongo.Database) == "" {
			return fmt.Errorf("MONGO_DATABASE is required when mongo is enabled")
		}
	}

	if c.LLM.Enabled {
		if strings.TrimSpace(c.LLM.APIKey) == "" {
			return fmt.Errorf("LLM_API_KEY is required when llm is enabled")
		}
		if strings.TrimSpace(c.LLM.Model) == "" {
			return fmt.Errorf("LLM_MODEL is required when llm is enabled")
		}
	}

	if c.MySQL.Enabled && strings.TrimSpace(c.MySQL.DSN) == "" {
		return fmt.Errorf("MYSQL_DSN is required when mysql is enabled")
	}
	if c.SQLite.Enabled && strings.TrimSpace(c.SQLite.DSN) == "" {
		return fmt.Errorf("SQLITE_DSN is required when sqlite is enabled")
	}

	return nil
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envFloat(key string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid LOG_LEVEL: %s", value)
	}
}

func parseLogFormat(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "text", "json":
		return strings.ToLower(strings.TrimSpace(value)), nil
	default:
		return "", fmt.Errorf("invalid LOG_FORMAT: %s", value)
	}
}
