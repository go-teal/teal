package debugging

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-teal/gota/dataframe"
	"github.com/go-teal/teal/pkg/dags"
	"github.com/go-teal/teal/pkg/models"
)

type DebuggingService struct {
	dag         *dags.DebugDag
	taskHistory map[string]DagExecutionResponseDTO // Store execution results by taskId
	mu          sync.RWMutex                       // Mutex for thread-safe access to taskHistory
}

func NewDebuggingService(dag *dags.DebugDag) *DebuggingService {
	return &DebuggingService{
		dag:         dag,
		taskHistory: make(map[string]DagExecutionResponseDTO),
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
			node.State = NodeState(debugNode.State) // Convert dags.NodeState to debugging.NodeState
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

func (s *DebuggingService) GetDagExecutionStatus(taskId string) DagExecutionResponseDTO {
	if s.dag == nil {
		return DagExecutionResponseDTO{
			TaskId:      taskId,
			Status:      DagExecutionStatusNotStarted,
			NodesStatus: []NodeStatusDTO{},
		}
	}

	// Use GetNodeStates to get thread-safe access to node states
	nodeStates := s.dag.GetNodeStates()

	totalAssets := len(nodeStates)
	completedAssets := 0
	failedAssets := 0
	inProgressAssets := 0
	tasks := make([]NodeStatusDTO, 0)
	lastTaskName := ""
	order := 0

	// First, count overall stats
	for _, nodeState := range nodeStates {
		switch nodeState {
		case dags.NodeStateSuccess:
			completedAssets++
		case dags.NodeStateFailed:
			failedAssets++
		case dags.NodeStateInProgress, dags.NodeStateTesting:
			inProgressAssets++
		}
	}

	// Build task status list from DAG graph order
	for _, taskGroup := range s.dag.DagGraph {
		for _, assetName := range taskGroup {
			if nodeState, exists := nodeStates[assetName]; exists {
				taskStatus := NodeStatusDTO{
					Name:  assetName,
					State: NodeState(nodeState),
					Order: order,
				}

				// Get additional details from the node if available
				if node := s.dag.GetNode(assetName); node != nil {
					taskStatus.ExecutionTimeMs = node.LastExecutionDuration
					if node.LastError != nil {
						taskStatus.Message = node.LastError.Error()
					}
					// Add start and end times
					if node.StartTime != nil {
						startTimeMs := node.StartTime.UnixMilli()
						taskStatus.StartTime = &startTimeMs
					}
					if node.EndTime != nil {
						endTimeMs := node.EndTime.UnixMilli()
						taskStatus.EndTime = &endTimeMs
					}
					// Add test counts
					taskStatus.TotalTests = len(node.Tests)
					taskStatus.PassedTests = node.TestsPassed
					taskStatus.FailedTests = node.TestsFailed

					// Add test results
					if len(node.TestResults) > 0 {
						taskStatus.TestResults = make([]TestResultDTO, len(node.TestResults))
						for i, tr := range node.TestResults {
							taskStatus.TestResults[i] = TestResultDTO{
								TestName:   tr.TestName,
								Status:     string(tr.Status),
								DurationMs: tr.DurationMs,
							}
							// Convert error to string for JSON serialization
							if tr.Error != nil {
								taskStatus.TestResults[i].ErrorMsg = tr.Error.Error()
							}
						}
					}
				}

				tasks = append(tasks, taskStatus)
				order++

				// Track last task name
				if nodeState != dags.NodeStateInitial {
					lastTaskName = assetName
				}
			}
		}
	}

	// Determine overall DAG status
	var status DagExecutionStatus
	if completedAssets == 0 && failedAssets == 0 && inProgressAssets == 0 {
		status = DagExecutionStatusNotStarted
	} else if inProgressAssets > 0 {
		status = DagExecutionStatusInProgress
	} else if failedAssets > 0 {
		status = DagExecutionStatusFailed
	} else if completedAssets == totalAssets {
		status = DagExecutionStatusSuccess
	} else {
		status = DagExecutionStatusInProgress // Partial completion
	}

	return DagExecutionResponseDTO{
		TaskId:           taskId,
		Status:           status,
		NodesStatus:      tasks,
		LastTaskName:     lastTaskName,
		CompletedAssets:  completedAssets,
		TotalAssets:      totalAssets,
		FailedAssets:     failedAssets,
		InProgressAssets: inProgressAssets,
	}
}

func (s *DebuggingService) ExecuteDag(taskId string, data map[string]interface{}) <-chan DagExecutionResponseDTO {
	responseChan := make(chan DagExecutionResponseDTO, 1)

	if s.dag == nil {
		response := DagExecutionResponseDTO{
			TaskId:      taskId,
			Status:      DagExecutionStatusFailed,
			NodesStatus: []NodeStatusDTO{},
		}
		// Store failed result in history
		s.mu.Lock()
		s.taskHistory[taskId] = response
		s.mu.Unlock()

		responseChan <- response
		close(responseChan)
		return responseChan
	}

	// Store initial status
	s.mu.Lock()
	s.taskHistory[taskId] = DagExecutionResponseDTO{
		TaskId:      taskId,
		Status:      DagExecutionStatusInProgress,
		NodesStatus: []NodeStatusDTO{},
	}
	s.mu.Unlock()

	// Create a channel for DAG results
	dagResultChan := make(chan map[string]interface{}, 1)

	// Start the DAG execution
	s.dag.Push(taskId, data, dagResultChan)

	// Monitor execution in a goroutine
	go func() {
		defer close(responseChan)

		// Periodically update task history during execution
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-dagResultChan:
				// DAG completed, get final status
				finalStatus := s.GetDagExecutionStatus(taskId)

				// Store final result in history
				s.mu.Lock()
				s.taskHistory[taskId] = finalStatus
				s.mu.Unlock()

				responseChan <- finalStatus
				return

			case <-ticker.C:
				// Update task history with current progress
				currentStatus := s.GetDagExecutionStatus(taskId)
				s.mu.Lock()
				s.taskHistory[taskId] = currentStatus
				s.mu.Unlock()

			case <-time.After(10 * time.Second):
				// Timeout - return current status as PENDING
				status := s.GetDagExecutionStatus(taskId)
				if status.Status == DagExecutionStatusInProgress {
					status.Status = DagExecutionStatusPending
				}

				// Store timeout result in history
				s.mu.Lock()
				s.taskHistory[taskId] = status
				s.mu.Unlock()

				responseChan <- status
				return
			}
		}
	}()

	return responseChan
}

func (s *DebuggingService) GetTaskStatus(taskId string) DagExecutionResponseDTO {
	// First check if we have a stored history for this taskId
	s.mu.RLock()
	if history, exists := s.taskHistory[taskId]; exists {
		s.mu.RUnlock()
		return history
	}
	s.mu.RUnlock()

	// If no history exists, return current DAG state with the taskId
	return s.GetDagExecutionStatus(taskId)
}

func (s *DebuggingService) ResetDagState() error {
	// Clear task history map
	s.mu.Lock()
	s.taskHistory = make(map[string]DagExecutionResponseDTO)
	s.mu.Unlock()

	if s.dag == nil {
		return nil // Nothing else to reset
	}

	// Reset all node states and clear execution history
	for _, node := range s.dag.NodeMap {
		node.State = dags.NodeStateInitial

		// Clear execution results - this handles both regular results and dataframes
		// If the asset is_data_framed, LastResult would be a *dataframe.DataFrame
		// Setting it to nil properly clears the pointer and allows GC to free memory
		node.LastResult = nil
		node.LastError = nil

		// Reset execution metrics
		node.LastExecutionDuration = 0
		node.LastTestsDuration = 0
		node.TestsPassed = 0
		node.TestsFailed = 0
		node.TestResults = nil // Clear test results
		node.StartTime = nil
		node.EndTime = nil

		// Note: The DAG structure (Upstreams, Downstreams pointers) remains intact
		// Only the data in LastResult is cleared, which will be recreated on next execution
	}

	return nil
}

func (s *DebuggingService) GetAllTasks() TaskListResponseDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]TaskSummaryDTO, 0, len(s.taskHistory))

	// Convert task history to summary DTOs
	for taskId, execution := range s.taskHistory {
		summary := TaskSummaryDTO{
			TaskId:           taskId,
			Status:           execution.Status,
			TotalAssets:      execution.TotalAssets,
			CompletedAssets:  execution.CompletedAssets,
			FailedAssets:     execution.FailedAssets,
			InProgressAssets: execution.InProgressAssets,
		}

		// Find the earliest start time and latest end time from all tasks
		var earliestStart *int64
		var latestEnd *int64

		for _, task := range execution.NodesStatus {
			if task.StartTime != nil {
				if earliestStart == nil || *task.StartTime < *earliestStart {
					earliestStart = task.StartTime
				}
			}
			if task.EndTime != nil {
				if latestEnd == nil || *task.EndTime > *latestEnd {
					latestEnd = task.EndTime
				}
			}
		}

		summary.StartTime = earliestStart
		summary.EndTime = latestEnd

		tasks = append(tasks, summary)
	}

	// Sort tasks by start time (earliest first)
	sort.Slice(tasks, func(i, j int) bool {
		// Handle nil start times (put them at the end)
		if tasks[i].StartTime == nil && tasks[j].StartTime == nil {
			return false
		}
		if tasks[i].StartTime == nil {
			return false
		}
		if tasks[j].StartTime == nil {
			return true
		}
		return *tasks[i].StartTime < *tasks[j].StartTime
	})

	return TaskListResponseDTO{
		Tasks: tasks,
		Total: len(tasks),
	}
}

func (s *DebuggingService) ExecuteAsset(assetName string, taskId string) <-chan AssetExecuteResponseDTO {
	responseChan := make(chan AssetExecuteResponseDTO, 1)

	go func() {
		defer close(responseChan)

		response := AssetExecuteResponseDTO{
			TaskId:    taskId,
			AssetName: assetName,
			Status:    NodeStateInitial,
		}

		if s.dag == nil {
			response.Status = NodeStateFailed
			response.Error = "DAG not initialized"
			responseChan <- response
			return
		}

		// Get the node from the DAG
		node := s.dag.GetNode(assetName)
		if node == nil {
			response.Status = NodeStateFailed
			response.Error = "Asset not found in DAG"
			responseChan <- response
			return
		}

		// Collect upstream data from the current DAG state
		inputData := make(map[string]interface{})
		upstreamsUsed := []string{}

		for _, upstream := range node.Upstreams {
			if upstream.LastResult != nil {
				inputData[upstream.Name] = upstream.LastResult
				upstreamsUsed = append(upstreamsUsed, upstream.Name)
			}
		}

		response.UpstreamsUsed = upstreamsUsed

		// Create a channel for execution result
		execDone := make(chan struct{})

		// Execute the asset in a goroutine
		go func() {
			startTime := time.Now()
			startTimeMs := startTime.UnixMilli()
			response.StartTime = &startTimeMs

			// Update node state
			node.State = dags.NodeStateInProgress
			node.StartTime = &startTime

			// Execute the asset
			result, err := node.Asset.Execute(inputData)

			endTime := time.Now()
			endTimeMs := endTime.UnixMilli()
			response.EndTime = &endTimeMs
			response.ExecutionTimeMs = endTime.Sub(startTime).Milliseconds()

			// Update node with execution results
			node.EndTime = &endTime
			node.LastExecutionDuration = response.ExecutionTimeMs

			if err != nil {
				node.State = dags.NodeStateFailed
				node.LastError = err
				response.Status = NodeStateFailed
				response.Error = err.Error()
			} else {
				node.State = dags.NodeStateSuccess
				node.LastResult = result
				node.LastError = nil
				response.Status = NodeStateSuccess
				response.Result = result
			}

			close(execDone)
		}()

		// Wait for completion or timeout
		select {
		case <-execDone:
			// Execution completed, send the response
			responseChan <- response
		case <-time.After(10 * time.Second):
			// Timeout - return IN_PROGRESS status
			response.Status = NodeStateInProgress
			response.Error = "Execution still in progress after 10 seconds"
			responseChan <- response
		}
	}()

	return responseChan
}

func (s *DebuggingService) GetAssetData(assetName string) AssetDataResponseDTO {
	response := AssetDataResponseDTO{
		AssetName: assetName,
		HasData:   false,
	}

	if s.dag == nil {
		response.Error = "DAG not initialized"
		return response
	}

	// Get the node from the DAG
	node := s.dag.GetNode(assetName)
	if node == nil {
		response.Error = "Asset not found in DAG"
		return response
	}

	// Check if node has data
	if node.LastResult == nil {
		response.Error = "No data available for this asset (not executed or reset)"
		return response
	}

	response.HasData = true

	// Check if the asset is data-framed by looking at its descriptor
	if node.Asset != nil {
		descriptor := node.Asset.GetDescriptor()
		switch desc := descriptor.(type) {
		case *models.SQLModelDescriptor:
			response.IsDataFramed = desc.ModelProfile.IsDataFramed
		case *models.RawModelDescriptor:
			response.IsDataFramed = desc.ModelProfile.IsDataFramed
		}
	}

	// Determine data type and serialize accordingly
	switch v := node.LastResult.(type) {
	case *dataframe.DataFrame:
		response.DataType = "dataframe"
		response.IsDataFramed = true

		// Get dataframe metadata
		response.RowCount = v.Nrow()
		response.ColumnCount = v.Ncol()
		response.Columns = v.Names()

		// Convert dataframe to JSON
		// Create a slice of maps for JSON representation
		records := make([]map[string]interface{}, 0, v.Nrow())
		for i := 0; i < v.Nrow(); i++ {
			record := make(map[string]interface{})
			for _, colName := range v.Names() {
				col := v.Col(colName)
				if col.Err != nil {
					continue
				}
				record[colName] = col.Elem(i).Val()
			}
			records = append(records, record)
		}
		response.Data = records

	case map[string]interface{}:
		response.DataType = "map"
		response.Data = v

	case []interface{}:
		response.DataType = "array"
		response.Data = v

	case string:
		response.DataType = "string"
		response.Data = v

	case int, int64, float64, bool:
		response.DataType = fmt.Sprintf("%T", v)
		response.Data = v

	default:
		// For any other type, try to serialize as JSON
		response.DataType = reflect.TypeOf(v).String()

		// Attempt to marshal to JSON and unmarshal back to interface{}
		// This ensures the data is JSON-serializable
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			response.Error = fmt.Sprintf("Failed to serialize data: %v", err)
			response.HasData = false
		} else {
			var jsonData interface{}
			if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
				response.Error = fmt.Sprintf("Failed to parse serialized data: %v", err)
				response.HasData = false
			} else {
				response.Data = jsonData
			}
		}
	}

	return response
}
