package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "strings"
)

// Global logger instance
var Log *zap.Logger

// Sets up a global Zap logger with the given log level.
func InitLogger(logLevel string) error {
    var level zapcore.Level

    // Convert string level to zapcore.Level
    switch strings.ToLower(logLevel) {
    case "debug":
        level = zapcore.DebugLevel
    case "info":
        level = zapcore.InfoLevel
    case "warn":
        level = zapcore.WarnLevel
    case "error":
        level = zapcore.ErrorLevel
    default:
        level = zapcore.InfoLevel // fallback
    }

    // Configure encoder
    config := zap.Config{
        Level:            zap.NewAtomicLevelAt(level),
        Development:      false,
        Encoding:         "json",          // structured JSON logs
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
        EncoderConfig: zapcore.EncoderConfig{
            MessageKey:   "message",
            LevelKey:     "level",
            TimeKey:      "time",
            NameKey:      "logger",
            CallerKey:    "caller",
            StacktraceKey: "stacktrace",
            LineEnding:   zapcore.DefaultLineEnding,
            EncodeLevel:  zapcore.LowercaseLevelEncoder,
            EncodeTime:   zapcore.ISO8601TimeEncoder,
            EncodeCaller: zapcore.ShortCallerEncoder,
        },
    }

    log, err := config.Build()
    if err != nil {
        return err
    }

    Log = log
    return nil
}
