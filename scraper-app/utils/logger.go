package utils

import (
	"os"

	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"

	"go.uber.org/zap/zapcore"
)

// Wrapper for zapcore.Level. Returns log level based on environment log level, else default
func logLevel() zapcore.Level {
	if level, err := zapcore.ParseLevel(GetEnvOrDefault("LOG_LEVEL", "INFO")); err != nil {
		return zapcore.InfoLevel
	} else {
		return level
	}
}

var jsonEncoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
var textEncoder = zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())

var fileLogger = zapcore.NewCore(
	jsonEncoder,
	zapcore.AddSync(&lumberjack.Logger{
		Filename:   GetEnvOrDefault("LOGS_PATH", "../tmp/log/scaper/scraper-app.json"),
		MaxSize:    GetIntEnvOrDefault("LOGS_MAX_SIZE", 500),
		MaxBackups: GetIntEnvOrDefault("LOGS_MAX_BACKUPS", 5),
		MaxAge:     GetIntEnvOrDefault("LOGS_MAX_AGE", 72),
	}),
	logLevel(),
)

var consoleLogger = zapcore.NewCore(
	textEncoder,
	zapcore.AddSync(os.Stdout),
	logLevel(),
)

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

var Logger = zap.New(zapcore.NewTee(loggers()...), zapOpts()...)
