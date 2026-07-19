package debugging

import (
	"fmt"
	"time"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/dags"
	"github.com/go-teal/teal/pkg/models"
	"github.com/rs/zerolog/log"
)

// relationMaintenanceOp describes a destructive DDL maintenance operation on the
// physical relation an asset materializes (drop / truncate).
//
// Each operation declares only its policy — which DDL to run for a given
// materialization — while runRelationMaintenance owns the shared orchestration
// (validation, concurrency locking, transaction, node reset). Adding a new
// maintenance operation is a matter of adding a resolver, not touching the
// orchestration: Open/Closed. The DDL itself is domain knowledge already generated
// into the model descriptor (DropTableSQL/DropViewSQL/TruncateTableSQL), so the
// driver contract stays untouched — the service composes existing primitives
// (Begin/Exec/Commit/CheckTableExists) rather than growing the DBDriver interface.
type relationMaintenanceOp struct {
	// name is a short verb used in logs and user-facing messages ("drop", "truncate").
	name string
	// resolveSQL returns the DDL to run for the given model, or ok=false when the
	// operation does not apply to that materialization (e.g. truncate on a view,
	// or any maintenance on a custom/raw asset).
	resolveSQL func(*models.SQLModelDescriptor) (sql string, ok bool)
}

// resolveDropSQL is the drop policy: DROP TABLE for table/incremental models,
// DROP VIEW for view models, not applicable to custom/raw. Package-level and pure
// so the materialization branching can be unit-tested without a DB or DAG.
func resolveDropSQL(d *models.SQLModelDescriptor) (string, bool) {
	switch d.ModelProfile.Materialization {
	case configs.MAT_TABLE, configs.MAT_INCREMENTAL:
		return d.DropTableSQL, true
	case configs.MAT_VIEW:
		return d.DropViewSQL, true
	default:
		return "", false
	}
}

// resolveTruncateSQL is the truncate policy: only table/incremental models can be
// truncated (a view has no rows of its own; custom/raw own no managed table).
func resolveTruncateSQL(d *models.SQLModelDescriptor) (string, bool) {
	switch d.ModelProfile.Materialization {
	case configs.MAT_TABLE, configs.MAT_INCREMENTAL:
		return d.TruncateTableSQL, true
	default:
		return "", false
	}
}

// DropAssetPersistedData drops the physical relation an asset materializes:
// DROP TABLE for table/incremental models, DROP VIEW for view models. It is
// idempotent — reported as success when the relation is already absent — and is
// not applicable to custom or raw assets, which teal does not own as a droppable
// relation.
func (s *DebuggingService) DropAssetPersistedData(assetName, taskId string) <-chan AssetExecuteResponseDTO {
	return s.runRelationMaintenance(assetName, taskId, relationMaintenanceOp{name: "drop", resolveSQL: resolveDropSQL})
}

// TruncateAssetTable empties the table an asset materializes (table/incremental
// only), leaving the table structure in place. It is not applicable to view,
// custom or raw assets.
func (s *DebuggingService) TruncateAssetTable(assetName, taskId string) <-chan AssetExecuteResponseDTO {
	return s.runRelationMaintenance(assetName, taskId, relationMaintenanceOp{name: "truncate", resolveSQL: resolveTruncateSQL})
}

// runRelationMaintenance is the shared orchestration for destructive relation
// operations. It validates the request against the current DAG, resolves the DDL
// via the operation's policy, and runs it against the target database under the
// connection's concurrency lock inside a single transaction — mirroring the
// runtime asset path (DuckDB is single-writer; Postgres relies on MVCC). On
// success it clears any cached preview on the node so the dashboard stops showing
// data that no longer exists.
func (s *DebuggingService) runRelationMaintenance(assetName, taskId string, op relationMaintenanceOp) <-chan AssetExecuteResponseDTO {
	responseChan := make(chan AssetExecuteResponseDTO, 1)

	go func() {
		defer close(responseChan)

		startTime := time.Now()
		startMs := startTime.UnixMilli()
		response := AssetExecuteResponseDTO{
			AssetName: assetName,
			TaskId:    taskId,
			Status:    NodeStateFailed,
			StartTime: &startMs,
		}
		fail := func(msg string) {
			response.Status = NodeStateFailed
			response.Error = msg
			endMs := time.Now().UnixMilli()
			response.EndTime = &endMs
			responseChan <- response
		}

		// --- Validate the request against the current DAG state ---
		s.mu.RLock()
		if s.dag == nil {
			s.mu.RUnlock()
			fail("DAG not initialized")
			return
		}
		if !s.dag.IsConnected() {
			s.mu.RUnlock()
			fail("Database connections not established. Please connect to databases first using POST /api/dag/connect")
			return
		}
		node := s.dag.GetNode(assetName)
		if node == nil {
			s.mu.RUnlock()
			fail("Asset not found in DAG")
			return
		}
		descriptor := node.Asset.GetDescriptor()
		s.mu.RUnlock()

		// Only SQL models own a droppable/truncatable relation; raw (Go) assets do not.
		sqlModelDesc, ok := descriptor.(*models.SQLModelDescriptor)
		if !ok {
			fail(fmt.Sprintf("%s is not applicable: asset does not materialize a SQL relation", op.name))
			return
		}

		sqlDDL, applicable := op.resolveSQL(sqlModelDesc)
		if !applicable {
			fail(fmt.Sprintf("%s is not applicable to '%s' materialization", op.name, sqlModelDesc.ModelProfile.Materialization))
			return
		}

		dbConnection := core.GetInstance().GetDBConnection(sqlModelDesc.ModelProfile.Connection)
		if dbConnection == nil {
			fail(fmt.Sprintf("Connection '%s' not found", sqlModelDesc.ModelProfile.Connection))
			return
		}

		// --- Execute the DDL against production data ---
		// Serialize with the connection's concurrency lock (a real mutex for DuckDB's
		// single-writer model; a no-op for Postgres) and wrap in one transaction.
		dbConnection.ConcurrencyLock()
		defer dbConnection.ConcurrencyUnlock()

		tx, err := dbConnection.Begin()
		if err != nil {
			fail(fmt.Sprintf("Failed to begin transaction: %v", err))
			return
		}

		// Idempotency: nothing to do if the relation is already gone. CheckTableExists
		// covers views too, since information_schema.tables lists them.
		if !dbConnection.CheckTableExists(tx, sqlModelDesc.Name) {
			if err := dbConnection.Commit(tx); err != nil {
				fail(fmt.Sprintf("Failed to commit: %v", err))
				return
			}
			s.resetNodeMaterialization(node, startTime)
			log.Info().Str("taskId", taskId).Str("assetName", assetName).
				Msgf("Relation already absent — %s is a no-op", op.name)
			s.finishMaintenance(&response, startTime, taskId, assetName)
			responseChan <- response
			return
		}

		if err := dbConnection.Exec(tx, sqlDDL); err != nil {
			dbConnection.Rollback(tx)
			log.Error().Caller().Str("taskId", taskId).Str("assetName", assetName).
				Str("sql", sqlDDL).Err(err).Msgf("Failed to %s relation", op.name)
			fail(fmt.Sprintf("Failed to %s relation: %v", op.name, err))
			return
		}

		if err := dbConnection.Commit(tx); err != nil {
			fail(fmt.Sprintf("Failed to commit %s: %v", op.name, err))
			return
		}

		// The persisted relation changed — drop any cached preview so the UI reflects it.
		s.resetNodeMaterialization(node, startTime)

		log.Info().Str("taskId", taskId).Str("assetName", assetName).
			Str("sql", sqlDDL).Msgf("Relation %s executed successfully", op.name)
		s.finishMaintenance(&response, startTime, taskId, assetName)
		responseChan <- response
	}()

	return responseChan
}

// finishMaintenance stamps a successful maintenance response and persists it to the
// asset history so GET /api/dag/asset/:name/data reflects the new state.
func (s *DebuggingService) finishMaintenance(response *AssetExecuteResponseDTO, startTime time.Time, taskId, assetName string) {
	response.Status = NodeStateSuccess
	response.Error = ""
	endMs := time.Now().UnixMilli()
	response.EndTime = &endMs
	response.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	s.storeAssetExecutionMetadata(taskId, assetName, *response)
}

// resetNodeMaterialization clears a node's cached result after its persisted relation
// has been dropped or emptied, so the dashboard stops showing stale data.
func (s *DebuggingService) resetNodeMaterialization(node *dags.DagAssetDebugService, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	node.LastResult = nil
	node.LastError = nil
	node.State = dags.NodeStateInitial
	node.EndTime = &at
}
