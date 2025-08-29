package logwriter

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"

	"github.com/rs/zerolog"
)

type contextKey string

const TaskIdKey contextKey = "taskId"

type StoringConsoleWriter struct {
	consoleWriter zerolog.ConsoleWriter
	storage       map[string][]interface{}
	mu            sync.RWMutex
	ctx           context.Context
}

func NewStoringConsoleWriter(ctx context.Context, output io.Writer) *StoringConsoleWriter {
	return &StoringConsoleWriter{
		consoleWriter: zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: "15:04:05",
			NoColor:    false,
		},
		storage: make(map[string][]interface{}),
		ctx:     ctx,
	}
}

func (w *StoringConsoleWriter) Write(p []byte) (n int, err error) {
	n, err = w.consoleWriter.Write(p)
	if err != nil {
		return n, err
	}

	var logData map[string]interface{}
	if err := json.Unmarshal(p, &logData); err == nil {
		w.storeLog(logData)
	}

	return n, nil
}

func (w *StoringConsoleWriter) storeLog(logData map[string]interface{}) {
	taskName := w.extractTaskName(logData)

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.storage[taskName]; !exists {
		w.storage[taskName] = make([]interface{}, 0)
	}
	w.storage[taskName] = append(w.storage[taskName], logData)
}

func (w *StoringConsoleWriter) extractTaskName(logData map[string]interface{}) string {
	if taskName, ok := logData["task_name"].(string); ok {
		return taskName
	}

	if w.ctx != nil {
		if taskName, ok := w.ctx.Value(TaskIdKey).(string); ok {
			return taskName
		}
	}

	return "default"
}

func (w *StoringConsoleWriter) GetLogs(taskId string) []interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if logs, exists := w.storage[taskId]; exists {
		result := make([]interface{}, len(logs))
		copy(result, logs)
		return result
	}

	return []interface{}{}
}

func (w *StoringConsoleWriter) GetAllLogs() map[string][]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make(map[string][]interface{})
	for taskId, logs := range w.storage {
		logsCopy := make([]interface{}, len(logs))
		copy(logsCopy, logs)
		result[taskId] = logsCopy
	}

	return result
}

func (w *StoringConsoleWriter) ClearLogs(taskId string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.storage, taskId)
}

func (w *StoringConsoleWriter) ClearAllLogs() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.storage = make(map[string][]interface{})
}

func (w *StoringConsoleWriter) SetNoColor(noColor bool) {
	w.consoleWriter.NoColor = noColor
}

func (w *StoringConsoleWriter) SetTimeFormat(format string) {
	w.consoleWriter.TimeFormat = format
}

func (w *StoringConsoleWriter) SetFieldsOrder(order []string) {
	w.consoleWriter.FieldsOrder = order
}

func (w *StoringConsoleWriter) UpdateContext(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ctx = ctx
}

type MultiWriter struct {
	writers []io.Writer
}

func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(bytes.Clone(p))
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}

func WithTaskName(ctx context.Context, taskId string) context.Context {
	return context.WithValue(ctx, TaskIdKey, taskId)
}

func GetTaskName(ctx context.Context) (string, bool) {
	taskName, ok := ctx.Value(TaskIdKey).(string)
	return taskName, ok
}
