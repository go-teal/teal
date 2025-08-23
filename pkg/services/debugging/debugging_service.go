package debugging

import (
	"strings"

	"github.com/go-teal/teal/pkg/dags"
	"github.com/go-teal/teal/pkg/models"
)

type DebuggingService struct {
	dag *dags.DebugDag
}

func NewDebuggingService(dag *dags.DebugDag) *DebuggingService {
	return &DebuggingService{
		dag: dag,
	}
}

func (s *DebuggingService) GetDagNodes() []DagNodeDTO {
	if s.dag == nil {
		return []DagNodeDTO{}
	}

	nodes := make([]DagNodeDTO, 0)

	for name, asset := range s.dag.AssetsMap {
		node := DagNodeDTO{
			Name:            name,
			Upstreams:       asset.GetUpstreams(),
			Downstreams:     asset.GetDownstreams(),
			State:           NodeStateInitial,
			TotalTests:      0,
			SuccessfulTests: 0,
		}

		// Get actual runtime state from NodeMap if available
		if debugNode, exists := s.dag.NodeMap[name]; exists {
			node.State = NodeState(debugNode.State)  // Convert dags.NodeState to debugging.NodeState
			node.SuccessfulTests = debugNode.TestsPassed
			node.LastExecutionDuration = debugNode.LastExecutionDuration
			node.LastTestsDuration = debugNode.LastTestsDuration
			// TotalTests will be set below from model profile
		}

		descriptor := asset.GetDescriptor()

		switch desc := descriptor.(type) {
		case *models.SQLModelDescriptor:
			node.SQLSelectQuery = strings.TrimSpace(desc.RawSQL)
			
			// Use appropriate compiled query based on materialization type
			node.Materialization = MaterializationType(desc.ModelProfile.Materialization)
			switch node.Materialization {
			case MaterializationView:
				node.SQLCompiledQuery = strings.TrimSpace(desc.CreateViewSQL)
			case MaterializationTable, MaterializationIncremental:
				node.SQLCompiledQuery = strings.TrimSpace(desc.InsertSQL)
			case MaterializationCustom:
				// Custom materialization uses RawSQL directly
				node.SQLCompiledQuery = strings.TrimSpace(desc.RawSQL)
			default:
				node.SQLCompiledQuery = strings.TrimSpace(desc.InsertSQL)
			}

			node.ConnectionName = desc.ModelProfile.Connection
			node.IsDataFramed = desc.ModelProfile.IsDataFramed
			node.PersistInputs = desc.ModelProfile.PersistInputs

			// Add tests from model profile
			if desc.ModelProfile.Tests != nil {
				node.Tests = make([]string, 0, len(desc.ModelProfile.Tests))
				for _, test := range desc.ModelProfile.Tests {
					node.Tests = append(node.Tests, test.Name)
				}
				node.TotalTests = len(desc.ModelProfile.Tests)
				// SuccessfulTests is already set from debugNode.TestsPassed above if available
			}

			// Find connection type from config
			if s.dag.Config != nil {
				for _, conn := range s.dag.Config.Connections {
					if conn.Name == desc.ModelProfile.Connection {
						node.ConnectionType = conn.Type
						break
					}
				}
			}

		case *models.RawModelDescriptor:
			node.Materialization = MaterializationRaw
			node.ConnectionName = desc.ModelProfile.Connection
			node.IsDataFramed = desc.ModelProfile.IsDataFramed
			node.PersistInputs = desc.ModelProfile.PersistInputs

			// Add tests from model profile
			if desc.ModelProfile.Tests != nil {
				node.Tests = make([]string, 0, len(desc.ModelProfile.Tests))
				for _, test := range desc.ModelProfile.Tests {
					node.Tests = append(node.Tests, test.Name)
				}
				node.TotalTests = len(desc.ModelProfile.Tests)
				// SuccessfulTests is already set from debugNode.TestsPassed above if available
			}

			// Find connection type from config
			if s.dag.Config != nil {
				for _, conn := range s.dag.Config.Connections {
					if conn.Name == desc.ModelProfile.Connection {
						node.ConnectionType = conn.Type
						break
					}
				}
			}
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func (s *DebuggingService) GetDagInstanceName() string {
	if s.dag == nil {
		return ""
	}
	return s.dag.DagInstanceName
}

func (s *DebuggingService) GetTestProfiles() []TestProfileDTO {
	if s.dag == nil || s.dag.TestsMap == nil {
		return []TestProfileDTO{}
	}

	tests := make([]TestProfileDTO, 0)

	for name, test := range s.dag.TestsMap {
		testDTO := TestProfileDTO{
			Name:   name,
			Status: TestStatusInitial,
		}

		// Get the descriptor to access SQL and connection info
		descriptor := test.GetDescriptor()
		if desc, ok := descriptor.(*models.SQLModelTestDescriptor); ok {
			testDTO.SQL = strings.TrimSpace(desc.RawSQL)

			if desc.TestProfile != nil {
				testDTO.ConnectionName = desc.TestProfile.Connection

				// Find connection type from config
				if s.dag.Config != nil {
					for _, conn := range s.dag.Config.Connections {
						if conn.Name == desc.TestProfile.Connection {
							testDTO.ConnectionType = conn.Type
							break
						}
					}
				}
			}
		}

		tests = append(tests, testDTO)
	}

	return tests
}
