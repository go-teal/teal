package debugging

import (
	"encoding/base64"
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

func (s *DebuggingService) ExecuteAsset(assetName string, taskId string, taskUUID string) <-chan AssetExecuteResponseDTO {
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

			// Store initial in-progress state
			s.storeAssetExecutionMetadata(taskId, assetName, response)

			// Execute the asset
			// Create TaskContext
			ctx := &processing.TaskContext{
				TaskID:       taskId,
				TaskUUID:     taskUUID,
				InstanceName: s.dag.DagInstanceName,
				InstanceUUID: s.dag.DagInstanceUUID,
			}
			result, err := node.Asset.Execute(ctx, inputData)

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
				Error:         "Execution still in progress after 10 seconds",
			}
			responseChan <- timeoutResponse
			// Execution continues in background and will update metadata when done
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
func (s *DebuggingService) GetAssetData(taskId, assetName string) AssetExecuteResponseDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Default response if not found
	response := AssetExecuteResponseDTO{
		TaskId:    taskId,
		AssetName: assetName,
		Status:    NodeStateInitial,
	}

	// First check if we have execution metadata for this task
	if taskAssets, exists := s.assetHistory[taskId]; exists {
		if assetMetadata, found := taskAssets[assetName]; found {
			// Start with stored metadata
			response = assetMetadata

			// Now fetch the actual data from the node if execution completed
			if s.dag != nil && (response.Status == NodeStateSuccess || response.Status == NodeStateFailed) {
				if node := s.dag.GetNode(assetName); node != nil {
					// Add the actual result from node's LastResult (for downstream access)
					if node.LastResult != nil && response.Status == NodeStateSuccess {
						response.Result = node.LastResult
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
				response.Result = node.LastResult
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

// GetAssetDataLegacy is the old implementation for backward compatibility
func (s *DebuggingService) GetAssetDataLegacy(assetName string) AssetDataResponseDTO {
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
