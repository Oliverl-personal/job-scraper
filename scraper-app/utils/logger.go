package utils

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"

	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var jsonEncoder = zapcore.NewJSONEncoder(jsonEncodeConfig)
var textEncoder = zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
var fileLogger zapcore.Core
var consoleLogger zapcore.Core

// initializes a logger. If environment is not loaded before calling this function
// the logger will take on default logging settings
func InitLogger() {
	fmt.Println("in initlogger")
	fileLogger = zapcore.NewCore(
		jsonEncoder,
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   GetEnvOrDefault("LOGS_PATH", "../tmp/log/scaper/scraper-app.json"),
			MaxSize:    GetIntEnvOrDefault("LOGS_MAX_SIZE", 500),
			MaxBackups: GetIntEnvOrDefault("LOGS_MAX_BACKUPS", 5),
			MaxAge:     GetIntEnvOrDefault("LOGS_MAX_AGE", 72),
		}),
		logLevel(),
	)
	consoleLogger = zapcore.NewCore(
		textEncoder,
		zapcore.AddSync(os.Stdout),
		logLevel(),
	)
	Logger = zap.New(zapcore.NewTee(loggers()...), zapOpts()...)
	fmt.Println("init logger complete")
}

// Wrapper for zapcore.Level. Returns log level based on environment log level, else default
func logLevel() zapcore.Level {
	fmt.Println("Log level")
	if level, err := zapcore.ParseLevel(GetEnvOrDefault("LOG_LEVEL", "INFO")); err == nil {
		return level
	} else {
		return zapcore.InfoLevel
	}
}

var jsonEncodeConfig = zapcore.EncoderConfig{
	TimeKey:        "ts",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	FunctionKey:    zapcore.OmitKey,
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

// loggers return array of log cores to be used in zap logger.
// Includes console logger, and if logging is enabled, the file logger
func loggers() []zapcore.Core {
	cores := make([]zapcore.Core, 0)
	cores = append(cores, consoleLogger)
	if GetBoolEnvOrDefault("LOGS_FILE_ENABLED", true) {
		cores = append(cores, fileLogger)
	}
	return cores
}

func zapOpts() []zap.Option {
	return []zap.Option{zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)}
}
