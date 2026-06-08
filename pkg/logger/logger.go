package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger.
type Logger struct {
	*zap.Logger
}

// New creates a new Logger based on the given level and format.
func New(level, format string) (*Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
	return &Logger{zapLogger}, nil
}

// NewNop returns a no-op logger for testing.
func NewNop() *Logger {
	return &Logger{zap.NewNop()}
}

// With creates a child logger with additional fields.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

// WithRequestID adds a request ID field to the logger.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.With(zap.String("request_id", requestID))
}

// WithUserID adds a user ID field.
func (l *Logger) WithUserID(userID string) *Logger {
	return l.With(zap.String("user_id", userID))
}

// WithError adds an error field.
func (l *Logger) WithError(err error) *Logger {
	return l.With(zap.Error(err))
}
