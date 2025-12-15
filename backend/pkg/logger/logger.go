package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.SugaredLogger for structured logging
type Logger struct {
	*zap.SugaredLogger
}

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, console
}

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	// Parse log level
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, err
	}

	// Configure encoder
	var encoder zapcore.Encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar := zapLogger.Sugar()

	return &Logger{SugaredLogger: sugar}, nil
}

// WithField adds a single field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{SugaredLogger: l.SugaredLogger.With(key, value)}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	var args []interface{}
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{SugaredLogger: l.SugaredLogger.With(args...)}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{SugaredLogger: l.SugaredLogger.With("error", err.Error())}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.SugaredLogger.Sync()
}

// Default logger instance
var defaultLogger *Logger

// InitDefault initializes the default logger
func InitDefault(cfg Config) error {
	logger, err := New(cfg)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// GetDefault returns the default logger instance
func GetDefault() *Logger {
	if defaultLogger == nil {
		// Fallback to a basic logger if not initialized
		logger, _ := New(Config{Level: "info", Format: "console"})
		defaultLogger = logger
	}
	return defaultLogger
}

// Package-level convenience functions

func Debug(args ...interface{}) {
	GetDefault().Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	GetDefault().Debugf(template, args...)
}

func Info(args ...interface{}) {
	GetDefault().Info(args...)
}

func Infof(template string, args ...interface{}) {
	GetDefault().Infof(template, args...)
}

func Warn(args ...interface{}) {
	GetDefault().Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	GetDefault().Warnf(template, args...)
}

func Error(args ...interface{}) {
	GetDefault().Error(args...)
}

func Errorf(template string, args ...interface{}) {
	GetDefault().Errorf(template, args...)
}

func Fatal(args ...interface{}) {
	GetDefault().Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	GetDefault().Fatalf(template, args...)
}
