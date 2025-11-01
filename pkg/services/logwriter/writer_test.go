package logwriter

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStoringConsoleWriter(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)

	assert.NotNil(t, writer)
	assert.NotNil(t, writer.storage)
	assert.NotNil(t, writer.consoleWriter)
	assert.Equal(t, ctx, writer.ctx)
}

func TestStoringConsoleWriter_Write(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)

	logger := zerolog.New(writer).With().Timestamp().Logger()
	logger.Info().Str("task_name", "test_task").Msg("test message")

	logs := writer.GetLogs("test_task")
	assert.Len(t, logs, 1)

	logData, ok := logs[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test message", logData["message"])
	assert.Equal(t, "test_task", logData["task_name"])

	assert.Contains(t, buf.String(), "test message")
}

func TestStoringConsoleWriter_ExtractTaskName(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		logData  map[string]interface{}
		expected string
	}{
		{
			name:     "task name from log data",
			ctx:      context.Background(),
			logData:  map[string]interface{}{"task_name": "log_task"},
			expected: "log_task",
		},
		{
			name:     "task name from context",
			ctx:      WithTaskName(context.Background(), "ctx_task"),
			logData:  map[string]interface{}{},
			expected: "ctx_task",
		},
		{
			name:     "log data takes precedence over context",
			ctx:      WithTaskName(context.Background(), "ctx_task"),
			logData:  map[string]interface{}{"task_name": "log_task"},
			expected: "log_task",
		},
		{
			name:     "default task name",
			ctx:      context.Background(),
			logData:  map[string]interface{}{},
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			writer := NewStoringConsoleWriter(tt.ctx, buf)
			result := writer.extractTaskName(tt.logData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStoringConsoleWriter_GetLogs(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)

	logger := zerolog.New(writer).With().Timestamp().Logger()
	logger.Info().Str("task_name", "task1").Msg("message 1")
	logger.Info().Str("task_name", "task1").Msg("message 2")
	logger.Info().Str("task_name", "task2").Msg("message 3")

	task1Logs := writer.GetLogs("task1")
	assert.Len(t, task1Logs, 2)

	task2Logs := writer.GetLogs("task2")
	assert.Len(t, task2Logs, 1)

	emptyLogs := writer.GetLogs("nonexistent")
	assert.Len(t, emptyLogs, 0)
}

func TestStoringConsoleWriter_GetAllLogs(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)

	logger := zerolog.New(writer).With().Timestamp().Logger()
	logger.Info().Str("task_name", "task1").Msg("message 1")
	logger.Info().Str("task_name", "task2").Msg("message 2")
	logger.Info().Msg("message without task")

	allLogs := writer.GetAllLogs()
	assert.Len(t, allLogs, 3)
	assert.Len(t, allLogs["task1"], 1)
	assert.Len(t, allLogs["task2"], 1)
	assert.Len(t, allLogs["default"], 1)
}

func TestStoringConsoleWriter_ClearLogs(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)

	logger := zerolog.New(writer).With().Timestamp().Logger()
	logger.Info().Str("task_name", "task1").Msg("message 1")
	logger.Info().Str("task_name", "task2").Msg("message 2")

	writer.ClearLogs("task1")
	
	task1Logs := writer.GetLogs("task1")
	assert.Len(t, task1Logs, 0)

	task2Logs := writer.GetLogs("task2")
	assert.Len(t, task2Logs, 1)
}

func TestStoringConsoleWriter_ClearAllLogs(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)

	logger := zerolog.New(writer).With().Timestamp().Logger()
	logger.Info().Str("task_name", "task1").Msg("message 1")
	logger.Info().Str("task_name", "task2").Msg("message 2")

	writer.ClearAllLogs()
	
	allLogs := writer.GetAllLogs()
	assert.Len(t, allLogs, 0)
}

func TestStoringConsoleWriter_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)

	var wg sync.WaitGroup
	numGoroutines := 10
	numLogs := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			logger := zerolog.New(writer).With().Timestamp().Logger()
			for j := 0; j < numLogs; j++ {
				logger.Info().
					Str("task_name", "concurrent_task").
					Int("goroutine", id).
					Int("log_num", j).
					Msg("concurrent message")
			}
		}(i)
	}

	wg.Wait()

	logs := writer.GetLogs("concurrent_task")
	assert.Len(t, logs, numGoroutines*numLogs)
}

func TestStoringConsoleWriter_UpdateContext(t *testing.T) {
	ctx1 := WithTaskName(context.Background(), "task1")
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx1, buf)

	logger := zerolog.New(writer).With().Timestamp().Logger()
	logger.Info().Msg("message with context task1")

	ctx2 := WithTaskName(context.Background(), "task2")
	writer.UpdateContext(ctx2)
	logger.Info().Msg("message with context task2")

	task1Logs := writer.GetLogs("task1")
	assert.Len(t, task1Logs, 1)

	task2Logs := writer.GetLogs("task2")
	assert.Len(t, task2Logs, 1)
}

func TestStoringConsoleWriter_Settings(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)

	writer.SetNoColor(true)
	assert.True(t, writer.consoleWriter.NoColor)

	writer.SetTimeFormat("2006-01-02")
	assert.Equal(t, "2006-01-02", writer.consoleWriter.TimeFormat)

	fieldsOrder := []string{"level", "message", "task_name"}
	writer.SetFieldsOrder(fieldsOrder)
	assert.Equal(t, fieldsOrder, writer.consoleWriter.FieldsOrder)
}

func TestMultiWriter(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	
	multiWriter := NewMultiWriter(buf1, buf2)
	
	testData := []byte("test message")
	n, err := multiWriter.Write(testData)
	
	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, string(testData), buf1.String())
	assert.Equal(t, string(testData), buf2.String())
}

func TestMultiWriter_WithStoringConsoleWriter(t *testing.T) {
	ctx := context.Background()
	consoleBuf := &bytes.Buffer{}
	jsonBuf := &bytes.Buffer{}
	
	storingWriter := NewStoringConsoleWriter(ctx, consoleBuf)
	jsonWriter := zerolog.ConsoleWriter{Out: jsonBuf, NoColor: true}
	
	multiWriter := NewMultiWriter(storingWriter, jsonWriter)
	logger := zerolog.New(multiWriter).With().Timestamp().Logger()
	
	logger.Info().Str("task_name", "multi_task").Msg("multi message")
	
	logs := storingWriter.GetLogs("multi_task")
	assert.Len(t, logs, 1)
	assert.Contains(t, consoleBuf.String(), "multi message")
	assert.Contains(t, jsonBuf.String(), "multi message")
}

func TestWithTaskName_GetTaskName(t *testing.T) {
	ctx := context.Background()
	
	taskName, ok := GetTaskName(ctx)
	assert.False(t, ok)
	assert.Empty(t, taskName)
	
	ctx = WithTaskName(ctx, "test_task")
	taskName, ok = GetTaskName(ctx)
	assert.True(t, ok)
	assert.Equal(t, "test_task", taskName)
}

func TestStoringConsoleWriter_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)
	
	invalidJSON := []byte("this is not json")
	n, err := writer.Write(invalidJSON)
	
	// ConsoleWriter will fail to parse invalid JSON
	assert.Error(t, err)
	assert.Equal(t, 0, n)
	
	allLogs := writer.GetAllLogs()
	assert.Len(t, allLogs, 0)
}

func TestStoringConsoleWriter_StructuredLogging(t *testing.T) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)

	logger := zerolog.New(writer).With().Timestamp().Logger()
	
	type User struct {
		Name  string
		Email string
		Age   int
	}
	
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	
	logger.Info().
		Str("task_name", "structured_task").
		Interface("user", user).
		Str("operation", "create").
		Int("status_code", 201).
		Msg("User created successfully")
	
	logs := writer.GetLogs("structured_task")
	assert.Len(t, logs, 1)
	
	logData, ok := logs[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "User created successfully", logData["message"])
	assert.Equal(t, "structured_task", logData["task_name"])
	assert.Equal(t, "create", logData["operation"])
	assert.Equal(t, float64(201), logData["status_code"])
	assert.NotNil(t, logData["user"])
}

func BenchmarkStoringConsoleWriter_Write(b *testing.B) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)
	logger := zerolog.New(writer).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().
			Str("task_name", "bench_task").
			Int("iteration", i).
			Msg("benchmark message")
	}
}

func BenchmarkStoringConsoleWriter_ConcurrentWrite(b *testing.B) {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)

	b.RunParallel(func(pb *testing.PB) {
		logger := zerolog.New(writer).With().Timestamp().Logger()
		for pb.Next() {
			logger.Info().
				Str("task_name", "concurrent_bench").
				Msg("benchmark message")
		}
	})
}

func ExampleStoringConsoleWriter() {
	ctx := WithTaskName(context.Background(), "example_task")
	buf := &bytes.Buffer{}
	writer := NewStoringConsoleWriter(ctx, buf)
	
	logger := zerolog.New(writer).With().Timestamp().Logger()
	logger.Info().Msg("This is an example log")
	
	logs := writer.GetLogs("example_task")
	if len(logs) > 0 {
		if logData, ok := logs[0].(map[string]interface{}); ok {
			if strings.Contains(logData["message"].(string), "example") {
				// Output verified
			}
		}
	}
}