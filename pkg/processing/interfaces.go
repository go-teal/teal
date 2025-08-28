package processing

// TestStatus represents the status of a test execution
type TestStatus string

const (
	TestStatusSuccess  TestStatus = "SUCCESS"
	TestStatusFailed   TestStatus = "FAILED"
	TestStatusNotFound TestStatus = "NOT_FOUND"
)

// TestResult represents the result of a single test execution
type TestResult struct {
	TestName    string     `json:"testName"`
	Status      TestStatus `json:"status"`
	Error       error      `json:"error,omitempty"`
	DurationMs  int64      `json:"durationMs"`
	Message     string     `json:"message,omitempty"`
}

type Asset interface {
	Execute(input map[string]interface{}) (interface{}, error)
	RunTests(testsMap map[string]ModelTesting) []TestResult
	GetUpstreams() []string
	GetDownstreams() []string
	GetName() string
	GetDescriptor() any
}

type ModelTesting interface {
	Execute() (bool, string, error)
	GetDescriptor() any
}
