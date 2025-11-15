package logging

import (
	"context"
	"fmt"
	"os"
	"time"

	"user-services/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	defaultLogger *logrus.Logger
)

// Logger wraps logrus with additional functionality
type Logger struct {
	*logrus.Logger
}

// InitLogger initializes the global logger
func InitLogger() {
	cfg := config.GetConfig()

	defaultLogger = logrus.New()

	// Set log level based on environment
	switch cfg.Environment {
	case "production":
		defaultLogger.SetLevel(logrus.InfoLevel)
		defaultLogger.SetFormatter(&logrus.JSONFormatter{})
	case "staging":
		defaultLogger.SetLevel(logrus.InfoLevel)
		defaultLogger.SetFormatter(&logrus.JSONFormatter{})
	default:
		defaultLogger.SetLevel(logrus.DebugLevel)
		defaultLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Set output to stdout by default
	defaultLogger.SetOutput(os.Stdout)
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		InitLogger()
	}
	return &Logger{Logger: defaultLogger}
}

// WithRequestID adds a request ID to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.WithField("request_id", requestID)}
}

// WithUserID adds a user ID to the logger
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{Logger: l.WithField("user_id", userID)}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields logrus.Fields) *Logger {
	return &Logger{Logger: l.WithFields(fields)}
}

// WithError adds an error to the logger with stack trace if available
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.WithError(err)}
}

// RequestMiddleware adds request logging and correlation ID middleware
func RequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate or extract request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in response header and context
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		logger := GetLogger().WithRequestID(requestID)

		// Log request
		logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"user_agent": c.Request.UserAgent(),
			"ip":         c.ClientIP(),
		}).Info("Request started")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log response
		status := c.Writer.Status()
		logLevel := logrus.InfoLevel
		if status >= 400 && status < 500 {
			logLevel = logrus.WarnLevel
		} else if status >= 500 {
			logLevel = logrus.ErrorLevel
		}

		logger.WithFields(logrus.Fields{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status":      status,
			"duration":    duration,
			"size":        c.Writer.Size(),
			"error":       c.Errors.String(),
		}).Log(logLevel, "Request completed")
	}
}

// AuditLogger provides structured audit logging
type AuditLogger struct {
	*Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{Logger: GetLogger()}
}

// LogAuth logs authentication events
func (a *AuditLogger) LogAuth(ctx context.Context, userID, email, action, status string, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"event_type": "authentication",
		"user_id":    userID,
		"email":      email,
		"action":     action,
		"status":     status,
		"timestamp":  time.Now().UTC(),
	}

	// Add request ID if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	// Add metadata
	for k, v := range metadata {
		fields[k] = v
	}

	a.WithFields(fields).Info("Authentication event")
}

// LogSecurity logs security-related events
func (a *AuditLogger) LogSecurity(ctx context.Context, event string, severity string, details map[string]interface{}) {
	fields := logrus.Fields{
		"event_type": "security",
		"event":      event,
		"severity":   severity,
		"timestamp":  time.Now().UTC(),
	}

	// Add request ID if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	// Add details
	for k, v := range details {
		fields[k] = v
	}

	level := logrus.InfoLevel
	switch severity {
	case "high":
		level = logrus.ErrorLevel
	case "medium":
		level = logrus.WarnLevel
	}

	a.WithFields(fields).Log(level, "Security event")
}

// LogDataAccess logs data access events
func (a *AuditLogger) LogDataAccess(ctx context.Context, userID, resource, action string, success bool) {
	fields := logrus.Fields{
		"event_type": "data_access",
		"user_id":    userID,
		"resource":   resource,
		"action":     action,
		"success":    success,
		"timestamp":  time.Now().UTC(),
	}

	// Add request ID if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields["request_id"] = requestID
	}

	a.WithFields(fields).Info("Data access event")
}

// PerformanceLogger tracks performance metrics
type PerformanceLogger struct {
	*Logger
}

// NewPerformanceLogger creates a new performance logger
func NewPerformanceLogger() *PerformanceLogger {
	return &PerformanceLogger{Logger: GetLogger()}
}

// LogDatabaseQuery logs database query performance
func (p *PerformanceLogger) LogDatabaseQuery(query string, duration time.Duration, rows int, err error) {
	fields := logrus.Fields{
		"event_type": "database_query",
		"query":      query,
		"duration":   duration,
		"rows":       rows,
		"timestamp":  time.Now().UTC(),
	}

	if err != nil {
		fields["error"] = err.Error()
		p.WithFields(fields).Error("Database query failed")
	} else {
		p.WithFields(fields).Debug("Database query completed")
	}
}

// LogExternalAPI logs external API calls
func (p *PerformanceLogger) LogExternalAPI(service, endpoint string, duration time.Duration, statusCode int, err error) {
	fields := logrus.Fields{
		"event_type":  "external_api",
		"service":     service,
		"endpoint":    endpoint,
		"duration":    duration,
		"status_code": statusCode,
		"timestamp":   time.Now().UTC(),
	}

	if err != nil {
		fields["error"] = err.Error()
		p.WithFields(fields).Error("External API call failed")
	} else {
		p.WithFields(fields).Info("External API call completed")
	}
}

// LogCacheOperation logs cache operations
func (p *PerformanceLogger) LogCacheOperation(operation, key string, hit bool, duration time.Duration) {
	fields := logrus.Fields{
		"event_type": "cache_operation",
		"operation":  operation,
		"key":        key,
		"hit":        hit,
		"duration":   duration,
		"timestamp":  time.Now().UTC(),
	}

	p.WithFields(fields).Debug("Cache operation completed")
}

// GetContextLogger extracts logger from context or returns default
func GetContextLogger(ctx context.Context) *Logger {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return GetLogger().WithRequestID(id)
		}
	}
	return GetLogger()
}

// ContextWithLogger adds logger to context
func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, "logger", logger)
}