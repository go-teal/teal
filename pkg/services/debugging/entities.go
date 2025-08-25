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
	NodeStateInitial    NodeState = "INITIAL"
	NodeStateInProgress NodeState = "IN_PROGRESS"
	NodeStateTesting    NodeState = "TESTING"
	NodeStateFailed     NodeState = "FAILED"
	NodeStateSuccess    NodeState = "SUCCESS"
)

type TestStatus string

const (
	TestStatusInitial    TestStatus = "INITIAL"
	TestStatusInProgress TestStatus = "IN_PROGRESS"
	TestStatusFailed     TestStatus = "FAILED"
	TestStatusSuccess    TestStatus = "SUCCESS"
)

type DagNodeDTO struct {
	Name             string              `json:"name"`
	Downstreams      []string            `json:"downstreams"`
	Upstreams        []string            `json:"upstreams"`
	SQLSelectQuery   string              `json:"sqlSelectQuery"`
	SQLCompiledQuery string              `json:"sqlCompiledQuery"`
	Materialization  MaterializationType `json:"materialization"`
	ConnectionType   string              `json:"connectionType"`
	ConnectionName   string              `json:"connectionName"`
	IsDataFramed     bool                `json:"isDataFramed"`
	PersistInputs    bool                `json:"persistInputs"`
	Tests            []string            `json:"tests"`
	State                  NodeState           `json:"state"`
	TotalTests             int                 `json:"totalTests"`
	SuccessfulTests        int                 `json:"successfulTests"`
	LastExecutionDuration  int64               `json:"lastExecutionDuration"` // Duration in milliseconds
	LastTestsDuration      int64               `json:"lastTestsDuration"`     // Duration of tests execution in milliseconds
}

type TestProfileDTO struct {
	Name           string     `json:"name"`
	SQL            string     `json:"sql"`
	ConnectionName string     `json:"connectionName"`
	ConnectionType string     `json:"connectionType"`
	Status         TestStatus `json:"status"`
}

type DagExecutionStatus string

const (
	DagExecutionStatusNotStarted DagExecutionStatus = "NOT_STARTED"
	DagExecutionStatusInProgress DagExecutionStatus = "IN_PROGRESS"
	DagExecutionStatusSuccess    DagExecutionStatus = "SUCCESS"
	DagExecutionStatusFailed     DagExecutionStatus = "FAILED"
	DagExecutionStatusPending    DagExecutionStatus = "PENDING"
)

type TaskStatusDTO struct {
	Name             string    `json:"name"`
	State            NodeState `json:"state"`
	Order            int       `json:"order"`
	StartTime        *int64    `json:"startTime,omitempty"`        // Unix timestamp in milliseconds
	EndTime          *int64    `json:"endTime,omitempty"`          // Unix timestamp in milliseconds
	ExecutionTimeMs  int64     `json:"executionTimeMs,omitempty"`
	Message          string    `json:"message,omitempty"`
	CompletedAssets  int       `json:"completedAssets"`
	TotalAssets      int       `json:"totalAssets"`
	FailedAssets     int       `json:"failedAssets"`
	InProgressAssets int       `json:"inProgressAssets"`
}

type DagExecutionResponseDTO struct {
	TaskId       string             `json:"taskId"`
	Status       DagExecutionStatus `json:"status"`
	Tasks        []TaskStatusDTO    `json:"tasks"`
	LastTaskName string             `json:"lastTaskName"`
}

type DagRunRequestDTO struct {
	TaskId string                 `json:"taskId"`
	Data   map[string]interface{} `json:"data"`
}

type TaskSummaryDTO struct {
	TaskId          string             `json:"taskId"`
	Status          DagExecutionStatus `json:"status"`
	StartTime       *int64             `json:"startTime,omitempty"`
	EndTime         *int64             `json:"endTime,omitempty"`
	TotalAssets     int                `json:"totalAssets"`
	CompletedAssets int                `json:"completedAssets"`
	FailedAssets    int                `json:"failedAssets"`
	InProgressAssets int               `json:"inProgressAssets"`
}

type TaskListResponseDTO struct {
	Tasks []TaskSummaryDTO `json:"tasks"`
	Total int              `json:"total"`
}

type AssetExecuteRequestDTO struct {
	TaskId string `json:"taskId"`
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
}

type AssetDataResponseDTO struct {
	AssetName    string      `json:"assetName"`
	HasData      bool        `json:"hasData"`
	DataType     string      `json:"dataType"` // "dataframe", "map", "string", etc.
	IsDataFramed bool        `json:"isDataFramed"`
	Data         interface{} `json:"data,omitempty"`
	RowCount     int         `json:"rowCount,omitempty"`     // For dataframes
	ColumnCount  int         `json:"columnCount,omitempty"`  // For dataframes
	Columns      []string    `json:"columns,omitempty"`      // For dataframes
	Error        string      `json:"error,omitempty"`
}
