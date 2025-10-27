package debugging

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/gota/dataframe"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/dags"
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/processing"
)

type DebuggingService struct {
	dag          *dags.DebugDag
	taskHistory  map[string]DagExecutionResponseDTO            // Store execution results by taskId
	testHistory  map[string][]TestProfileDTO                   // Store test results by taskId
	assetHistory map[string]map[string]AssetExecuteResponseDTO // Store asset execution results by taskId -> assetName
	mu           sync.RWMutex                                  // Mutex for thread-safe access to taskHistory
}

func NewDebuggingService(dag *dags.DebugDag) *DebuggingService {
	return &DebuggingService{
		dag:          dag,
		taskHistory:  make(map[string]DagExecutionResponseDTO),
		testHistory:  make(map[string][]TestProfileDTO),
		assetHistory: make(map[string]map[string]AssetExecuteResponseDTO),
	}
}

func (s *DebuggingService) GetDagNodes() []DagNodeDTO {
	if s.dag == nil {
		return []DagNodeDTO{}
	}

	nodes := make([]DagNodeDTO, 0)

	// Traverse nodes following the order from DagGraph
	for taskGroupIndex, taskGroup := range s.dag.DagGraph {
		for _, name := range taskGroup {
			asset, exists := s.dag.AssetsMap[name]
			if !exists {
				continue
			}

			node := DagNodeDTO{
				Name:            name,
				Upstreams:       asset.GetUpstreams(),
				Downstreams:     asset.GetDownstreams(),
				State:           NodeStateInitial,
				TotalTests:      0,
				SuccessfulTests: 0,
				TaskGroupIndex:  taskGroupIndex,
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
				// Decode base64 encoded description if present
				if desc.ModelProfile.Description != "" {
					decoded, err := base64.StdEncoding.DecodeString(desc.ModelProfile.Description)
					if err == nil {
						node.Description = string(decoded)
					} else {
						// Fall back to raw description if decode fails
						node.Description = desc.ModelProfile.Description
					}
				}

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
				// Decode base64 encoded description if present
				if desc.ModelProfile.Description != "" {
					decoded, err := base64.StdEncoding.DecodeString(desc.ModelProfile.Description)
					if err == nil {
						node.Description = string(decoded)
					} else {
						// Fall back to raw description if decode fails
						node.Description = desc.ModelProfile.Description
					}
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
			}

			nodes = append(nodes, node)
		}
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
			Name: name,
		}

		// Get the descriptor to access SQL and connection info
		descriptor := test.GetDescriptor()
		if desc, ok := descriptor.(*models.SQLModelTestDescriptor); ok {
			testDTO.SQL = strings.TrimSpace(desc.RawSQL)

			if desc.TestProfile != nil {
				// Decode base64 encoded description if present
				if desc.TestProfile.Description != "" {
					decoded, err := base64.StdEncoding.DecodeString(desc.TestProfile.Description)
					if err == nil {
						testDTO.Description = string(decoded)
					} else {
						// Fall back to raw description if decode fails
						testDTO.Description = desc.TestProfile.Description
					}
				}
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

func (s *DebuggingService) GetTestResultsForTask(taskId string) []TestProfileDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if testResults, exists := s.testHistory[taskId]; exists {
		// Return a copy to avoid race conditions
		result := make([]TestProfileDTO, len(testResults))
		copy(result, testResults)
		return result
	}

	// If no test results for this taskId, return empty list
	return []TestProfileDTO{}
}

func (s *DebuggingService) storeTestResultsForTask(taskId string) {
	if s.dag == nil || s.dag.TestsMap == nil {
		return
	}

	testResults := make([]TestProfileDTO, 0)

	// Iterate through all tests and get their execution status from nodes
	for testName, testAsset := range s.dag.TestsMap {
		testDTO := TestProfileDTO{
			Name: testName,
			SQL:  "", // Will be populated from descriptor
		}

		// Get test descriptor for connection info and SQL
		if descriptor := testAsset.GetDescriptor(); descriptor != nil {
			if desc, ok := descriptor.(*models.SQLModelTestDescriptor); ok {
				testDTO.SQL = desc.RawSQL
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
		}

		testResults = append(testResults, testDTO)
	}

	// Store test results for this taskId
	s.mu.Lock()
	s.testHistory[taskId] = testResults
	s.mu.Unlock()
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
	testsFailedAssets := 0 // Track assets with failed tests
	tasks := make([]NodeStatusDTO, 0)
	lastTaskName := ""
	order := 0

	// First, count overall stats
	for _, nodeState := range nodeStates {
		switch nodeState {
		case dags.NodeStateSuccess:
			completedAssets++
		case dags.NodeStateTestsFailed:
			completedAssets++   // Count as completed even if tests failed
			testsFailedAssets++ // Track that tests failed
		case dags.NodeStateFailed:
			failedAssets++ // Only count actual execution failures
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
					// Only add test counts if this node has tests
					if len(node.Tests) > 0 {
						taskStatus.TotalTests = len(node.Tests)
						taskStatus.PassedTests = node.TestsPassed
						taskStatus.FailedTests = node.TestsFailed
					}

					// Add test results if any
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
		status = DagExecutionStatusFailed // Asset execution failed
	} else if completedAssets == totalAssets {
		// All assets completed, check if any tests failed
		if testsFailedAssets > 0 {
			status = DagExecutionStatusTestsFailed // Tests failed but all assets executed
		} else {
			status = DagExecutionStatusSuccess // Everything passed
		}
	} else {
		status = DagExecutionStatusInProgress // Partial completion
	}

	// Convert root test results if available
	var rootTestResults []TestResultDTO
	if s.dag != nil && s.dag.RootTestResults != nil {
		for _, tr := range s.dag.RootTestResults {
			rootTestResults = append(rootTestResults, TestResultDTO{
				TestName: tr.TestName,
				Status:   string(tr.Status),
				ErrorMsg: func() string {
					if tr.Error != nil {
						return tr.Error.Error()
					} else {
						return ""
					}
				}(),
				DurationMs: tr.DurationMs,
			})
		}
	}

	// Get TaskUUID if available
	var taskUUID string
	if s.dag != nil {
		taskUUID = s.dag.GetTaskUUID(taskId)
	}

	return DagExecutionResponseDTO{
		TaskId:           taskId,
		TaskUUID:         taskUUID,
		Status:           status,
		NodesStatus:      tasks,
		LastTaskName:     lastTaskName,
		CompletedAssets:  completedAssets,
		TotalAssets:      totalAssets,
		FailedAssets:     failedAssets,
		InProgressAssets: inProgressAssets,
		RootTestResults:  rootTestResults,
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

	// Check if DAG is connected to databases
	if !s.dag.IsConnected() {
		response := DagExecutionResponseDTO{
			TaskId:      taskId,
			Status:      DagExecutionStatusFailed,
			NodesStatus: []NodeStatusDTO{{
				Name:    "connection_check",
				State:   NodeStateFailed,
				Message: "Database connections not established. Please connect to databases first using POST /api/dag/connect",
			}},
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

		// Create timeout timer
		timeout := time.NewTimer(10 * time.Second)
		defer timeout.Stop()

		for {
			select {
			case <-dagResultChan:
				// DAG completed, get final status
				finalStatus := s.GetDagExecutionStatus(taskId)

				// Store test results for this task
				s.storeTestResultsForTask(taskId)

				// Store final result in history
				s.mu.Lock()
				s.taskHistory[taskId] = finalStatus
				s.mu.Unlock()

				// Send response only if within timeout window
				select {
				case responseChan <- finalStatus:
					// Response sent successfully
				default:
					// Channel already closed or timeout occurred, but execution completed
					// The result is still stored in history for later retrieval
				}
				return

			case <-ticker.C:
				// Update task history with current progress
				currentStatus := s.GetDagExecutionStatus(taskId)
				s.mu.Lock()
				s.taskHistory[taskId] = currentStatus
				s.mu.Unlock()

			case <-timeout.C:
				// Timeout - return current status as PENDING but let execution continue
				status := s.GetDagExecutionStatus(taskId)
				if status.Status == DagExecutionStatusInProgress {
					status.Status = DagExecutionStatusPending
				}

				// Store timeout result in history
				s.mu.Lock()
				s.taskHistory[taskId] = status
				s.mu.Unlock()

				responseChan <- status

				// Start background goroutine to continue monitoring execution
				go s.monitorDagExecutionInBackground(taskId, dagResultChan)
				return
			}
		}
	}()

	return responseChan
}

// monitorDagExecutionInBackground continues monitoring DAG execution after a timeout
func (s *DebuggingService) monitorDagExecutionInBackground(taskId string, dagResultChan <-chan map[string]interface{}) {
	// Continue updating task history periodically
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-dagResultChan:
			// DAG completed in background, update final status
			finalStatus := s.GetDagExecutionStatus(taskId)

			// Store test results for this task
			s.storeTestResultsForTask(taskId)

			// Update status from PENDING to actual completion status
			s.mu.Lock()
			s.taskHistory[taskId] = finalStatus
			s.mu.Unlock()

			return // Exit background monitoring

		case <-ticker.C:
			// Continue updating task history with current progress
			currentStatus := s.GetDagExecutionStatus(taskId)

			// Keep status as IN_PROGRESS (not PENDING) since we're still executing
			if currentStatus.Status != DagExecutionStatusFailed &&
				currentStatus.Status != DagExecutionStatusSuccess &&
				currentStatus.Status != DagExecutionStatusTestsFailed {
				currentStatus.Status = DagExecutionStatusInProgress
			}

			s.mu.Lock()
			s.taskHistory[taskId] = currentStatus
			s.mu.Unlock()
		}
	}
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
	// Clear task history map and asset history
	s.mu.Lock()
	s.taskHistory = make(map[string]DagExecutionResponseDTO)
	s.assetHistory = make(map[string]map[string]AssetExecuteResponseDTO)
	s.testHistory = make(map[string][]TestProfileDTO)
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

	// Clear root test results
	s.dag.RootTestResults = nil

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
			TaskUUID:         execution.TaskUUID,
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

func (s *DebuggingService) MutateAsset(assetName string, taskId string, taskUUID string) <-chan AssetExecuteResponseDTO {
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

		// Check if DAG is connected to databases
		if !s.dag.IsConnected() {
			response.Status = NodeStateFailed
			response.Error = "Database connections not established. Please connect to databases first using POST /api/dag/connect"
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

			// Store initial in-progress state
			s.storeAssetExecutionMetadata(taskId, assetName, response)

			// Execute the asset
			// Create TaskContext
			ctx := &processing.TaskContext{
				TaskID:       taskId,
				TaskUUID:     taskUUID,
				InstanceName: s.dag.DagInstanceName,
				InstanceUUID: s.dag.DagInstanceUUID,
				Input:        inputData,
			}
			result, err := node.Asset.Execute(ctx)

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
				node.LastResult = result // Store in node for downstream access
				node.LastError = nil
				response.Status = NodeStateSuccess
				// Don't store result in response to avoid memory issues
				// Result will be accessed from node.LastResult when needed
			}

			// Update stored metadata after completion (without the actual data)
			s.storeAssetExecutionMetadata(taskId, assetName, response)

			close(execDone)
		}()

		// Wait for completion or timeout
		select {
		case <-execDone:
			// Execution completed within timeout
			responseChan <- response
		case <-time.After(10 * time.Second):
			// Timeout - return IN_PROGRESS status but let execution continue
			timeoutResponse := AssetExecuteResponseDTO{
				TaskId:        taskId,
				AssetName:     assetName,
				Status:        NodeStateInProgress,
				StartTime:     response.StartTime,
				UpstreamsUsed: response.UpstreamsUsed,
				Error:         "Execution timeout (still processing in background)",
			}

			// Store intermediate metadata so GET /api/dag/asset/:name/data can retrieve status
			// The execution continues in background and will update this when complete
			s.mu.Lock()
			s.storeAssetExecutionMetadata(taskId, assetName, timeoutResponse)
			s.mu.Unlock()

			responseChan <- timeoutResponse
			// Execution continues in background and will update metadata when done (line 785)
		}
	}()

	return responseChan
}

// storeAssetExecutionMetadata stores the asset execution metadata (without data) for later retrieval
func (s *DebuggingService) storeAssetExecutionMetadata(taskId, assetName string, response AssetExecuteResponseDTO) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.assetHistory[taskId] == nil {
		s.assetHistory[taskId] = make(map[string]AssetExecuteResponseDTO)
	}

	// Store metadata only, not the actual result data
	metadataOnly := AssetExecuteResponseDTO{
		TaskId:          response.TaskId,
		AssetName:       response.AssetName,
		Status:          response.Status,
		StartTime:       response.StartTime,
		EndTime:         response.EndTime,
		ExecutionTimeMs: response.ExecutionTimeMs,
		Error:           response.Error,
		UpstreamsUsed:   response.UpstreamsUsed,
		// Result is intentionally omitted to avoid storing large data
	}

	s.assetHistory[taskId][assetName] = metadataOnly
}

// GetAssetData retrieves asset execution data for a specific task
func (s *DebuggingService) GetAssetData(taskId, assetName string, offset, limit int) AssetExecuteResponseDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Default response if not found
	response := AssetExecuteResponseDTO{
		TaskId:    taskId,
		AssetName: assetName,
		Status:    NodeStateInitial,
		// Offset:    offset,
		// Limit:     limit,
	}

	// First check if we have execution metadata for this task
	if taskAssets, exists := s.assetHistory[taskId]; exists {
		if assetMetadata, found := taskAssets[assetName]; found {
			// Start with stored metadata
			response = assetMetadata
			// Update pagination parameters
			// response.Offset = offset
			// response.Limit = limit

			// Now fetch the actual data from the node if execution completed
			if s.dag != nil && (response.Status == NodeStateSuccess || response.Status == NodeStateFailed) {
				if node := s.dag.GetNode(assetName); node != nil {
					// Add the actual result from node's LastResult (for downstream access)
					if node.LastResult != nil && response.Status == NodeStateSuccess {
						totalRecords := 0
						response.Result, totalRecords = s.serializeResultWithPagination(node.LastResult, offset, limit)
						response.TotalRecords = totalRecords
					}

					// Update status from current node state in case it changed
					response.Status = NodeState(node.State)

					// Update timing if execution completed after timeout
					if node.EndTime != nil && response.EndTime == nil {
						endTimeMs := node.EndTime.UnixMilli()
						response.EndTime = &endTimeMs
						if node.StartTime != nil {
							response.ExecutionTimeMs = node.EndTime.Sub(*node.StartTime).Milliseconds()
						}
					}
				}
			}

			return response
		}
	}

	// If no stored execution metadata, check current node state
	if s.dag != nil {
		if node := s.dag.GetNode(assetName); node != nil {
			// Map node state to response
			response.Status = NodeState(node.State)

			// Add timing information if available
			if node.StartTime != nil {
				startTimeMs := node.StartTime.UnixMilli()
				response.StartTime = &startTimeMs
			}
			if node.EndTime != nil {
				endTimeMs := node.EndTime.UnixMilli()
				response.EndTime = &endTimeMs
			}
			response.ExecutionTimeMs = node.LastExecutionDuration

			// Add result or error from node
			if node.LastResult != nil {
				totalRecords := 0
				response.Result, totalRecords = s.serializeResultWithPagination(node.LastResult, offset, limit)
				response.TotalRecords = totalRecords
			}
			if node.LastError != nil {
				response.Error = node.LastError.Error()
			}

			// Add upstreams information
			upstreams := []string{}
			for _, upstream := range node.Upstreams {
				upstreams = append(upstreams, upstream.Name)
			}
			response.UpstreamsUsed = upstreams
		} else {
			response.Error = "Asset not found in DAG"
			response.Status = NodeStateFailed
		}
	} else {
		response.Error = "DAG not initialized"
		response.Status = NodeStateFailed
	}

	return response
}

// serializeResultWithPagination converts the result to a JSON-serializable format with pagination support
// Returns the serialized result and total record count
// offset: starting index (0-based), limit: max records to return (0 = no limit)
func (s *DebuggingService) serializeResultWithPagination(result interface{}, offset, limit int) (interface{}, int) {
	if result == nil {
		return nil, 0
	}

	// Check if result is a DataFrame
	if df, ok := result.(*dataframe.DataFrame); ok {
		totalRecords := df.Nrow()

		// Calculate slice boundaries
		start := offset
		end := totalRecords

		if start >= totalRecords {
			// Offset beyond data, return empty slice
			return []map[string]interface{}{}, totalRecords
		}

		if limit > 0 {
			end = start + limit
			if end > totalRecords {
				end = totalRecords
			}
		}

		// Convert dataframe to JSON-serializable format with pagination
		records := make([]map[string]interface{}, 0, end-start)
		for i := start; i < end; i++ {
			record := make(map[string]interface{})
			for _, colName := range df.Names() {
				col := df.Col(colName)
				if col.Err != nil {
					continue
				}
				record[colName] = col.Elem(i).Val()
			}
			records = append(records, record)
		}
		return records, totalRecords
	}

	// For arrays, apply pagination
	if arr, ok := result.([]interface{}); ok {
		totalRecords := len(arr)

		start := offset
		end := totalRecords

		if start >= totalRecords {
			return []interface{}{}, totalRecords
		}

		if limit > 0 {
			end = start + limit
			if end > totalRecords {
				end = totalRecords
			}
		}

		return arr[start:end], totalRecords
	}

	// For all other types, return as-is (maps, primitives)
	// Total records = 1 for non-array types
	return result, 1
}

// ExecuteAssetSelect executes the asset's SQL query using ToDataFrame and saves the result to the DAG node
// The SQL template is rendered first, executing all template functions (like Ref)
// Returns a response with taskId that can be used to retrieve the data via GetAssetData endpoint
func (s *DebuggingService) ExecuteAssetSelect(assetName, taskId string) <-chan AssetExecuteResponseDTO {
	responseChan := make(chan AssetExecuteResponseDTO, 1)

	go func() {
		defer close(responseChan)

		response := AssetExecuteResponseDTO{
			AssetName: assetName,
			TaskId:    taskId,
			Status:    NodeStateFailed,
		}

		// Create a channel for the actual execution
		executionChan := make(chan AssetExecuteResponseDTO, 1)

		go func() {
			startTime := time.Now()
			startTimeMs := startTime.UnixMilli()
			execResponse := AssetExecuteResponseDTO{
				AssetName:  assetName,
				TaskId:     taskId,
				Status:     NodeStateFailed,
				StartTime:  &startTimeMs,
			}

			// Validate and get node (with lock)
			s.mu.RLock()
			if s.dag == nil {
				s.mu.RUnlock()
				execResponse.Error = "DAG not initialized"
				executionChan <- execResponse
				return
			}

			// Check if DAG is connected to databases
			if !s.dag.IsConnected() {
				s.mu.RUnlock()
				execResponse.Error = "Database connections not established. Please connect to databases first using POST /api/dag/connect"
				executionChan <- execResponse
				return
			}

			node := s.dag.GetNode(assetName)
			if node == nil {
				s.mu.RUnlock()
				execResponse.Error = "Asset not found in DAG"
				executionChan <- execResponse
				return
			}

			descriptor := node.Asset.GetDescriptor()
			s.mu.RUnlock()

			// Validate asset type and get connection (no lock needed)
			sqlModelDesc, ok := descriptor.(*models.SQLModelDescriptor)
			if !ok {
				execResponse.Error = "Asset is not a SQL model (only SQL assets support select)"
				executionChan <- execResponse
				return
			}

			dbConnection := core.GetInstance().GetDBConnection(sqlModelDesc.ModelProfile.Connection)
			if dbConnection == nil {
				execResponse.Error = fmt.Sprintf("Connection '%s' not found", sqlModelDesc.ModelProfile.Connection)
				executionChan <- execResponse
				return
			}

			// Check if table exists for incremental materialization (no lock needed - DB operation)
			isIncremental := false
			if sqlModelDesc.ModelProfile.Materialization == configs.MAT_INCREMENTAL {
				tx, err := dbConnection.Begin()
				if err == nil {
					isIncremental = dbConnection.CheckTableExists(tx, sqlModelDesc.Name)
					dbConnection.Commit(tx)
				}
			}

			// Store initial in-progress state
			s.storeAssetExecutionMetadata(taskId, assetName, execResponse)

			// Update node state to in-progress (with lock)
			s.mu.Lock()
			node.State = dags.NodeStateInProgress
			node.StartTime = &startTime
			s.mu.Unlock()

			// Render and execute SQL (no lock needed - long operation)
			templateFuncs := pongo2.Context{
				"IsIncremental": func() bool {
					return isIncremental
				},
			}

			sqlTemplate, err := pongo2.FromString(sqlModelDesc.RawSQL)
			if err != nil {
				execResponse.Error = fmt.Sprintf("Failed to parse SQL template: %v", err)
				s.storeAssetExecutionMetadata(taskId, assetName, execResponse)
				executionChan <- execResponse
				return
			}

			context := processing.MergePongo2Context(
				processing.FromConnectionContext(dbConnection, nil, sqlModelDesc.Name, templateFuncs),
			)
			renderedSQL, err := sqlTemplate.Execute(context)
			if err != nil {
				execResponse.Error = fmt.Sprintf("Failed to render SQL template: %v", err)
				s.storeAssetExecutionMetadata(taskId, assetName, execResponse)
				executionChan <- execResponse
				return
			}

			// Execute the SQL query (long operation - no lock)
			df, err := dbConnection.ToDataFrame(renderedSQL)

			endTime := time.Now()
			endTimeMs := endTime.UnixMilli()
			execResponse.EndTime = &endTimeMs
			execResponse.ExecutionTimeMs = endTime.Sub(startTime).Milliseconds()

			if err != nil {
				execResponse.Error = fmt.Sprintf("Failed to execute query: %v", err)
				s.storeAssetExecutionMetadata(taskId, assetName, execResponse)
				executionChan <- execResponse
				return
			}

			if df == nil {
				execResponse.Error = "Query returned nil DataFrame"
				s.storeAssetExecutionMetadata(taskId, assetName, execResponse)
				executionChan <- execResponse
				return
			}

			// Update node with results (with lock)
			s.mu.Lock()
			node.LastResult = df
			node.State = dags.NodeStateSuccess
			node.EndTime = &endTime
			node.LastExecutionDuration = execResponse.ExecutionTimeMs
			s.mu.Unlock()

			// Build successful response
			execResponse.Status = NodeStateSuccess

			// Store metadata in asset history (function handles its own locking)
			s.storeAssetExecutionMetadata(taskId, assetName, execResponse)

			executionChan <- execResponse
		}()

		// Wait for execution with 10-second timeout
		select {
		case execResult := <-executionChan:
			response = execResult
		case <-time.After(10 * time.Second):
			response.Status = NodeStateInProgress
			response.Error = "Execution timeout (still processing in background)"

			// Store intermediate metadata so GET /api/dag/asset/:name/data can retrieve status
			// The execution continues in background and will update this when complete
			s.mu.Lock()
			s.storeAssetExecutionMetadata(taskId, assetName, response)
			s.mu.Unlock()

			// Continue listening for completion in background (don't block response)
			go func() {
				// Wait for the actual execution to complete
				if execResult := <-executionChan; execResult.Status == NodeStateSuccess {
					// Update stored metadata with final result when execution completes
					s.mu.Lock()
					s.storeAssetExecutionMetadata(taskId, assetName, execResult)
					s.mu.Unlock()
				}
			}()
		}

		responseChan <- response
	}()

	return responseChan
}

// Connect establishes connections to all databases
func (s *DebuggingService) Connect() error {
	if s.dag == nil {
		return fmt.Errorf("DAG not initialized")
	}
	return s.dag.Connect()
}

// Disconnect closes all database connections
func (s *DebuggingService) Disconnect() error {
	if s.dag == nil {
		return fmt.Errorf("DAG not initialized")
	}
	return s.dag.Disconnect()
}

// GetConnectionStatus returns the connection status with configuration details for each connection
func (s *DebuggingService) GetConnectionStatus() ConnectionStatusResponseDTO {
	response := ConnectionStatusResponseDTO{
		IsConnected: false,
		Connections: []ConnectionConfigDTO{},
	}

	if s.dag == nil {
		return response
	}

	// Get connection status from DAG
	response.IsConnected = s.dag.IsConnected()

	// Build connections list from config
	if s.dag.Config != nil && s.dag.Config.Connections != nil {
		for _, conn := range s.dag.Config.Connections {
			connDTO := ConnectionConfigDTO{
				Name: conn.Name,
				Type: conn.Type,
			}

			// Add configuration details if available
			if conn.Config != nil {
				connDTO.Host = conn.Config.Host
				connDTO.Port = conn.Config.Port
				connDTO.Database = conn.Config.Database
				connDTO.User = conn.Config.User
				connDTO.Path = conn.Config.Path
				connDTO.Extensions = conn.Config.Extensions
			}

			response.Connections = append(response.Connections, connDTO)
		}
	}

	return response
}
