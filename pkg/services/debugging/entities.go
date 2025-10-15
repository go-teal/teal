package debugging

type MaterializationType string

const (
	MaterializationTable       MaterializationType = "table"
	MaterializationIncremental MaterializationType = "incremental"
	MaterializationView        MaterializationType = "view"
	MaterializationCustom      MaterializationType = "custom"
	MaterializationRaw         MaterializationType = "raw"
)

type NodeState string

const (
	NodeStateInitial     NodeState = "INITIAL"
	NodeStateInProgress  NodeState = "IN_PROGRESS"
	NodeStateTesting     NodeState = "TESTING"
	NodeStateFailed      NodeState = "FAILED"
	NodeStateSuccess     NodeState = "SUCCESS"
	NodeStateTestsFailed NodeState = "TESTS_FAILED" // Asset succeeded but tests failed
)

type TestStatus string

const (
	TestStatusInitial    TestStatus = "INITIAL"
	TestStatusInProgress TestStatus = "IN_PROGRESS"
	TestStatusFailed     TestStatus = "FAILED"
	TestStatusSuccess    TestStatus = "SUCCESS"
)

type DagNodeDTO struct {
	Name                  string              `json:"name"`
	Description           string              `json:"description"`
	Downstreams           []string            `json:"downstreams"`
	Upstreams             []string            `json:"upstreams"`
	SQLSelectQuery        string              `json:"sqlSelectQuery"`
	SQLCompiledQuery      string              `json:"sqlCompiledQuery"`
	Materialization       MaterializationType `json:"materialization"`
	ConnectionType        string              `json:"connectionType"`
	ConnectionName        string              `json:"connectionName"`
	IsDataFramed          bool                `json:"isDataFramed"`
	PersistInputs         bool                `json:"persistInputs"`
	Tests                 []string            `json:"tests"`
	State                 NodeState           `json:"state"`
	TotalTests            int                 `json:"totalTests"`
	SuccessfulTests       int                 `json:"successfulTests"`
	LastExecutionDuration int64               `json:"lastExecutionDuration"` // Duration in milliseconds
	LastTestsDuration     int64               `json:"lastTestsDuration"`     // Duration of tests execution in milliseconds
	TaskGroupIndex        int                 `json:"TaskGroupIndex"`
}

type TestProfileDTO struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	SQL            string `json:"sql"`
	ConnectionName string `json:"connectionName"`
	ConnectionType string `json:"connectionType"`
}

type DagExecutionStatus string

const (
	DagExecutionStatusNotStarted  DagExecutionStatus = "NOT_STARTED"
	DagExecutionStatusInProgress  DagExecutionStatus = "IN_PROGRESS"
	DagExecutionStatusSuccess     DagExecutionStatus = "SUCCESS"
	DagExecutionStatusFailed      DagExecutionStatus = "FAILED"
	DagExecutionStatusPending     DagExecutionStatus = "PENDING"
	DagExecutionStatusTestsFailed DagExecutionStatus = "TESTS_FAILED" // All assets succeeded but some tests failed
)

// TestResultDTO represents a test result in API responses
type TestResultDTO struct {
	TestName   string `json:"testName"`
	Status     string `json:"status"` // SUCCESS, FAILED, NOT_FOUND
	ErrorMsg   string `json:"error,omitempty"`
	DurationMs int64  `json:"durationMs"`
}

type NodeStatusDTO struct {
	Name            string          `json:"name"`
	State           NodeState       `json:"state"`
	Order           int             `json:"order"`
	StartTime       *int64          `json:"startTime,omitempty"` // Unix timestamp in milliseconds
	EndTime         *int64          `json:"endTime,omitempty"`   // Unix timestamp in milliseconds
	ExecutionTimeMs int64           `json:"executionTimeMs,omitempty"`
	Message         string          `json:"message,omitempty"`
	TotalTests      int             `json:"totalTests"`
	PassedTests     int             `json:"passedTests"`
	FailedTests     int             `json:"failedTests"`
	TestResults     []TestResultDTO `json:"testResults,omitempty"`
}

type DagExecutionResponseDTO struct {
	TaskId           string             `json:"taskId"`
	TaskUUID         string             `json:"taskUuid,omitempty"`
	Status           DagExecutionStatus `json:"status"`
	NodesStatus      []NodeStatusDTO    `json:"nodes"`
	LastTaskName     string             `json:"lastTaskName"`
	CompletedAssets  int                `json:"completedAssets"`
	TotalAssets      int                `json:"totalAssets"`
	FailedAssets     int                `json:"failedAssets"`
	InProgressAssets int                `json:"inProgressAssets"`
	RootTestResults  []TestResultDTO    `json:"rootTestResults,omitempty"`
}

type DagRunRequestDTO struct {
	TaskId   string                 `json:"taskId"`
	TaskUUID string                 `json:"taskUuid,omitempty"`
	Data     map[string]interface{} `json:"data"`
}

// TaskSummaryDTO represents the summary of an entire task execution (identified by TaskId)
// It contains aggregate counts for all assets in the DAG execution
type TaskSummaryDTO struct {
	TaskId           string             `json:"taskId"`
	TaskUUID         string             `json:"taskUuid,omitempty"`
	Status           DagExecutionStatus `json:"status"`
	StartTime        *int64             `json:"startTime,omitempty"`
	EndTime          *int64             `json:"endTime,omitempty"`
	TotalAssets      int                `json:"totalAssets"`
	CompletedAssets  int                `json:"completedAssets"`
	FailedAssets     int                `json:"failedAssets"`
	InProgressAssets int                `json:"inProgressAssets"`
}

type TaskListResponseDTO struct {
	Tasks []TaskSummaryDTO `json:"tasks"`
	Total int              `json:"total"`
}

type AssetExecuteRequestDTO struct {
	TaskId   string `json:"taskId"`
	TaskUUID string `json:"taskUuid,omitempty"`
}

type AssetExecuteResponseDTO struct {
	TaskId          string      `json:"taskId"`
	AssetName       string      `json:"assetName"`
	Status          NodeState   `json:"status"`
	StartTime       *int64      `json:"startTime,omitempty"`
	EndTime         *int64      `json:"endTime,omitempty"`
	ExecutionTimeMs int64       `json:"executionTimeMs,omitempty"`
	Result          interface{} `json:"result,omitempty"`
	Error           string      `json:"error,omitempty"`
	UpstreamsUsed   []string    `json:"upstreamsUsed"`
	TotalRecords    int         `json:"totalRecords,omitempty"`
	Offset          int         `json:"offset,omitempty"`
	Limit           int         `json:"limit,omitempty"`
}

type AssetDataResponseDTO struct {
	AssetName    string      `json:"assetName"`
	HasData      bool        `json:"hasData"`
	DataType     string      `json:"dataType"` // "dataframe", "map", "string", etc.
	IsDataFramed bool        `json:"isDataFramed"`
	Data         interface{} `json:"data,omitempty"`
	RowCount     int         `json:"rowCount,omitempty"`    // For dataframes
	ColumnCount  int         `json:"columnCount,omitempty"` // For dataframes
	Columns      []string    `json:"columns,omitempty"`     // For dataframes
	Error        string      `json:"error,omitempty"`
}
