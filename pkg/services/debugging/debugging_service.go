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
			Name:        name,
			Upstreams:   asset.GetUpstreams(),
			Downstreams: asset.GetDownstreams(),
		}

		descriptor := asset.GetDescriptor()

		switch desc := descriptor.(type) {
		case *models.SQLModelDescriptor:
			node.SQLSelectQuery = strings.TrimSpace(desc.RawSQL)
			node.SQLCompiledQuery = strings.TrimSpace(desc.InsertSQL)

			node.Materialization = MaterializationType(desc.ModelProfile.Materialization)

			node.ConnectionName = desc.ModelProfile.Connection

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
