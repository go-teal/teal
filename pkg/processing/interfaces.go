package processing

// TaskContext holds runtime context for task execution
type TaskContext struct {
	TaskID       string                 // Task identifier from Push method
	TaskUUID     string                 // Unique UUID assigned in Push method
	InstanceName string                 // DAG instance name
	InstanceUUID string                 // Unique UUID assigned in constructor
	Input        map[string]interface{} // Input data from upstream tasks
}

// TestStatus represents the status of a test execution
type TestStatus string

const (
	TestStatusSuccess  TestStatus = "SUCCESS"
	TestStatusFailed   TestStatus = "FAILED"
	TestStatusNotFound TestStatus = "NOT_FOUND"
)

// TestResult represents the result of a single test execution
type TestResult struct {
	TestName   string     `json:"testName"`
	Status     TestStatus `json:"status"`
	Error      error      `json:"error,omitempty"`
	DurationMs int64      `json:"durationMs"`
	Message    string     `json:"message,omitempty"`
}

type Asset interface {
	Execute(ctx *TaskContext) (interface{}, error)
	RunTests(ctx *TaskContext, testsMap map[string]ModelTesting) []TestResult
	GetUpstreams() []string
	GetDownstreams() []string
	GetName() string
	GetDescriptor() any
}

type ModelTesting interface {
	Execute(ctx *TaskContext) (bool, string, error)
	GetDescriptor() any
}
