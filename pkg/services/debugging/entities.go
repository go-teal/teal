package debugging

type MaterializationType string

const (
	MaterializationTable       MaterializationType = "table"
	MaterializationIncremental MaterializationType = "incremental"
	MaterializationView        MaterializationType = "view"
	MaterializationCustom      MaterializationType = "custom"
	MaterializationRaw         MaterializationType = "raw"
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
}
