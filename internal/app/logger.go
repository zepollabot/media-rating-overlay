package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
)

// LogConfig holds configuration for the logger
type LogConfig struct {
	// LogFilePath is the path where logs will be written
	LogFilePath string
	// LogLevel is the minimum log level to record
	LogLevel zapcore.Level
	// MaxSize is the maximum size in megabytes of the log file before it gets rotated
	MaxSize int
	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int
	// MaxAge is the maximum number of days to retain old log files
	MaxAge int
	// Compress determines if the rotated log files should be compressed
	Compress bool
	// ServiceName is the name of the service for structured logging
	ServiceName string
	// ServiceVersion is the version of the service for structured logging
	ServiceVersion string
	// UseJSON determines if logs should be in JSON format (true) or console format (false)
	UseJSON bool
	// UseStdout determines if logs should be written to stdout in addition to the log file
	UseStdout bool
}

// DefaultConfig returns a default LogConfig
func DefaultConfig(logFilePath string) LogConfig {
	return LogConfig{
		LogFilePath:    logFilePath,
		LogLevel:       getDefaultLogLevel(),
		MaxSize:        100, // 100 MB
		MaxBackups:     5,
		MaxAge:         30, // 30 days
		Compress:       true,
		ServiceName:    "media-rating-overlay",
		ServiceVersion: "1.0.0",
		UseJSON:        true,
		UseStdout:      true,
	}
}

type ctxLogger struct{}

// NewLogger creates a new logger with the given configuration
func NewLogger(config LogConfig) (*zap.Logger, error) {
	logger, err := buildLogger(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	logger.Info("--------------------------------------------------")
	logger.Info("Welcome to Media Rating Overlay",
		zap.String("service", config.ServiceName),
		zap.String("version", config.ServiceVersion),
	)
	logger.Info("--------------------------------------------------")

	return logger, nil
}

// NewLoggerWithDefaults creates a new logger with default configuration
func NewLoggerWithDefaults(logFilePath string) (*zap.Logger, error) {
	config := DefaultConfig(logFilePath)
	return NewLogger(config)
}

// NewLoggerFromConfig creates a new logger from the application config
func NewLoggerFromConfig(config *config.Logger) (*zap.Logger, error) {
	// Convert string log level to zapcore.Level
	logLevel := getLogLevelFromString(config.LogLevel)

	// Create log config
	logConfig := LogConfig{
		LogFilePath:    config.LogFilePath,
		LogLevel:       logLevel,
		MaxSize:        config.MaxSize,
		MaxBackups:     config.MaxBackups,
		MaxAge:         config.MaxAge,
		Compress:       config.Compress,
		ServiceName:    config.ServiceName,
		ServiceVersion: config.ServiceVersion,
		UseJSON:        config.UseJSON,
		UseStdout:      config.UseStdout,
	}

	// If logFilePath is empty, use default
	if logConfig.LogFilePath == "" {
		logConfig.LogFilePath = "logs/media-rating-overlay.log"
	}

	// If service name is empty, use default
	if logConfig.ServiceName == "" {
		logConfig.ServiceName = "media-rating-overlay"
	}

	// If service version is empty, use default
	if logConfig.ServiceVersion == "" {
		logConfig.ServiceVersion = "1.0.0"
	}

	return NewLogger(logConfig)
}

// getLogLevelFromString converts a string log level to zapcore.Level
func getLogLevelFromString(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return getDefaultLogLevel()
	}
}

// setFilePermissions sets the correct permissions for the log file
func setFilePermissions(filename string) error {
	// Create the file if it doesn't exist
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	file.Close()

	// Set the correct permissions
	return os.Chmod(filename, 0644)
}

func buildLogger(config LogConfig) (*zap.Logger, error) {
	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.TimeKey = "timestamp"

	// Create encoder based on config
	var encoder zapcore.Encoder
	if config.UseJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create log directory if it doesn't exist
	if err := createLogsPathDirectory(config.LogFilePath); err != nil {
		return nil, fmt.Errorf("unable to create log directory: %w", err)
	}

	// Set up log rotation
	writer := &lumberjack.Logger{
		Filename:   config.LogFilePath,
		MaxSize:    config.MaxSize, // megabytes
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge, // days
		Compress:   config.Compress,
	}

	// Set correct permissions for the log file
	if err := setFilePermissions(config.LogFilePath); err != nil {
		return nil, fmt.Errorf("failed to set log file permissions: %w", err)
	}

	// Create core with file output
	fileCore := zapcore.NewCore(encoder, zapcore.AddSync(writer), config.LogLevel)

	// Create core with stdout output if enabled
	var core zapcore.Core
	if config.UseStdout {
		// Create core with both file and console output
		core = zapcore.NewTee(
			fileCore,
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(encoderConfig),
				zapcore.AddSync(os.Stdout),
				config.LogLevel,
			),
		)
	} else {
		// Use only file output
		core = fileCore
	}

	// Create logger with caller and stacktrace options
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		//zap.Fields(
		//	zap.String("service", config.ServiceName),
		//	zap.String("version", config.ServiceVersion),
		//),
	)

	return logger, nil
}

func getDefaultLogLevel() zapcore.Level {
	env := getEnv()
	switch env {
	case "PROD":
		return zapcore.InfoLevel
	default:
		return zapcore.DebugLevel
	}
}

func getEnv() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "DEV"
	}
	return env
}

// ContextWithLogger adds logger to context
func ContextWithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxLogger{}, l)
}

// LoggerFromContext returns logger from context
func LoggerFromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxLogger{}).(*zap.Logger); ok {
		return l
	}
	return nil
}

// WithFields adds fields to the logger
func WithFields(logger *zap.Logger, fields ...zap.Field) *zap.Logger {
	return logger.With(fields...)
}

// Sync flushes any buffered log entries
func Sync(logger *zap.Logger) error {
	return logger.Sync()
}

func createLogsPathDirectory(filePath string) error {
	path := filepath.Dir(filePath)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
