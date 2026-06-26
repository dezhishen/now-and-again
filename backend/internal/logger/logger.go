// Package logger provides a zap-based logger with daily rotation and compression.
//
// Usage:
//
//	logger.Init("logs")        // default: logs/ directory
//	logger.Info("message", zap.String("key", "val"))
//	logger.Sync()              // flush on shutdown
//
// Old logs are compressed to .gz after rotation.
package logger

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	L      *zap.Logger
	S      *zap.SugaredLogger
	dir    string
	mu     sync.Mutex
	writer *dailyWriter
)

// Init initializes the global logger. Logs go to both file (daily rotation) and stdout.
// logDir is the directory for log files (default "./logs").
func Init(logDir string) (*zap.Logger, error) {
	if logDir == "" {
		logDir = "logs"
	}
	dir = logDir
	os.MkdirAll(dir, 0755)

	writer = newDailyWriter(dir)

	// Console encoder (colored, human-readable)
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		EncodeTime:    zapcore.TimeEncoderOfLayout("15:04:05"),
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	})

	// File encoder (JSON for machine readability)
	fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	})

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(writer), zapcore.InfoLevel),
	)

	L = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	S = L.Sugar()

	// Start compression goroutine
	go compressLoop(dir)

	return L, nil
}

// Sync flushes any buffered log entries.
func Sync() {
	if L != nil {
		L.Sync()
	}
}

// dailyWriter writes to a date-stamped file, rotating at midnight.
type dailyWriter struct {
	dir   string
	file  *os.File
	today string
	mu    sync.Mutex
}

func newDailyWriter(dir string) *dailyWriter {
	w := &dailyWriter{dir: dir}
	w.rotate()
	return w
}

func (w *dailyWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if today != w.today {
		w.rotate()
	}

	return w.file.Write(p)
}

func (w *dailyWriter) rotate() {
	if w.file != nil {
		w.file.Close()
		// Compress old file in background
		oldPath := filepath.Join(w.dir, w.today+".log")
		go compressFile(oldPath)
	}

	w.today = time.Now().Format("2006-01-02")
	path := filepath.Join(w.dir, w.today+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: failed to open %s: %v\n", path, err)
		return
	}
	w.file = f
}

// compressLoop periodically compresses old log files.
func compressLoop(dir string) {
	for {
		time.Sleep(5 * time.Minute)
		entries, _ := os.ReadDir(dir)
		today := time.Now().Format("2006-01-02")
		for _, e := range entries {
			name := e.Name()
			// Skip today's file, .gz files, and non-log files
			if name == today+".log" || filepath.Ext(name) == ".gz" || filepath.Ext(name) != ".log" {
				continue
			}
			compressFile(filepath.Join(dir, name))
		}
	}
}

func compressFile(path string) {
	if _, err := os.Stat(path); err != nil {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	gzPath := path + ".gz"
	f, err := os.Create(gzPath)
	if err != nil {
		return
	}
	defer f.Close()

	w := gzip.NewWriter(f)
	w.Write(data)
	w.Close()
	os.Remove(path)
}

// ─── Convenience wrappers ────────────────────────────────────────

func Info(msg string, fields ...zap.Field)  { L.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { L.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { L.Error(msg, fields...) }
func Debug(msg string, fields ...zap.Field) { L.Debug(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { L.Fatal(msg, fields...) }

// Infof uses sugared logger for printf-style.
func Infof(template string, args ...interface{})  { S.Infof(template, args...) }
func Warnf(template string, args ...interface{})  { S.Warnf(template, args...) }
func Errorf(template string, args ...interface{}) { S.Errorf(template, args...) }
func Debugf(template string, args ...interface{}) { S.Debugf(template, args...) }
func Fatalf(template string, args ...interface{}) { S.Fatalf(template, args...) }
