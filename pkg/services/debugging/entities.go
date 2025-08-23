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
