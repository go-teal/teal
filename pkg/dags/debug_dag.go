package dags

import (
	"sync"
	"time"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/processing"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// NodeState represents the execution state of a node in the debug DAG
type NodeState string

const (
	NodeStateInitial     NodeState = "INITIAL"
	NodeStateInProgress  NodeState = "IN_PROGRESS"
	NodeStateTesting     NodeState = "TESTING"
	NodeStateFailed      NodeState = "FAILED"
	NodeStateSuccess     NodeState = "SUCCESS"
	NodeStateTestsFailed NodeState = "TESTS_FAILED" // Asset succeeded but tests failed
)

// DagAssetDebugService represents a node in the debug DAG with pointer-based connections
type DagAssetDebugService struct {
	Name                  string
	Asset                 processing.Asset
	Upstreams             []*DagAssetDebugService // Pointers to upstream assets
	Downstreams           []*DagAssetDebugService // Pointers to downstream assets
	State                 NodeState
	TestsPassed           int
	TestsFailed           int
	Tests                 map[string]processing.ModelTesting
	TestResults           []processing.TestResult // Store test execution results
	LastError             error
	LastResult            interface{}
	LastExecutionDuration int64      // Duration in milliseconds
	LastTestsDuration     int64      // Duration of tests execution in milliseconds
	StartTime             *time.Time // Start time of execution
	EndTime               *time.Time // End time of execution
}

// TestExecutionResult stores the result of an individual test execution
type TestExecutionResult struct {
	DataFrame  interface{} // DataFrame containing violation rows
	Status     string
	RowCount   int
	ExecutedAt time.Time
}

// DebugDag is a debug-enabled DAG implementation using pointers instead of channels
type DebugDag struct {
	InnerDag         DAG // Keep for compatibility but will be nil
	DagInstanceName  string
	DagInstanceUUID  string
	DagGraph         [][]string
	AssetsMap        map[string]processing.Asset
	TestsMap         map[string]processing.ModelTesting
	Config           *configs.Config
	NodeMap          map[string]*DagAssetDebugService           // Map of asset name to debug service node
	RootNodes        []*DagAssetDebugService                    // Nodes with no upstreams
	LeafNodes        []*DagAssetDebugService                    // Nodes with no downstreams
	RootTestResults  []processing.TestResult                    // Results from root tests
	TaskUUIDMap      map[string]string                          // Map of taskId to taskUUID
	TestExecutionMap map[string]map[string]*TestExecutionResult // Map of taskId -> testName -> result
	isConnected      bool                                       // Track database connection status

	// mu protects short read/write windows on shared state: NodeMap entries
	// (per-node mutable fields), TaskUUIDMap, TestExecutionMap, RootTestResults,
	// and isConnected. It MUST NOT be held across DB-backed calls
	// (Asset.Execute, RunTests, ModelTesting.Execute), otherwise status
	// polling from the UI is blocked for the duration of the DAG run.
	mu sync.RWMutex

	// execMu serializes concurrent Push calls (one DAG execution at a time).
	// Separated from mu so that long-running execution does not block readers.
	execMu sync.Mutex
}

// InitDebugDag creates a new DebugDag with pointer-based structure
func InitDebugDag(dagGraph [][]string,
	assetsMap map[string]processing.Asset,
	testsMap map[string]processing.ModelTesting,
	config *configs.Config,
	name string) *DebugDag {

	dag := &DebugDag{
		InnerDag:         nil, // No inner DAG for debug mode
		DagInstanceName:  name,
		DagInstanceUUID:  uuid.New().String(),
		DagGraph:         dagGraph,
		AssetsMap:        assetsMap,
		TestsMap:         testsMap,
		Config:           config,
		NodeMap:          make(map[string]*DagAssetDebugService),
		RootNodes:        make([]*DagAssetDebugService, 0),
		LeafNodes:        make([]*DagAssetDebugService, 0),
		TaskUUIDMap:      make(map[string]string),
		TestExecutionMap: make(map[string]map[string]*TestExecutionResult),
	}

	dag.build()
	return dag
}

// build constructs the pointer-based graph structure
func (d *DebugDag) build() {
	// First pass: Create all nodes
	for _, taskGroup := range d.DagGraph {
		for _, assetName := range taskGroup {
			asset, exists := d.AssetsMap[assetName]
			if !exists {
				log.Warn().Str("assetName", assetName).Msg("Asset not found for task")
				continue
			}

			// Create debug service node
			node := &DagAssetDebugService{
				Name:        assetName,
				Asset:       asset,
				Upstreams:   make([]*DagAssetDebugService, 0),
				Downstreams: make([]*DagAssetDebugService, 0),
				State:       NodeStateInitial,
				Tests:       make(map[string]processing.ModelTesting),
			}

			// Add tests for this node from model profile
			if d.TestsMap != nil {
				descriptor := asset.GetDescriptor()
				switch desc := descriptor.(type) {
				case *models.SQLModelDescriptor:
					if desc.ModelProfile != nil && desc.ModelProfile.Tests != nil {
						for _, testProfile := range desc.ModelProfile.Tests {
							if test, exists := d.TestsMap[testProfile.Name]; exists {
								node.Tests[testProfile.Name] = test
							}
						}
					}
				case *models.RawModelDescriptor:
					if desc.ModelProfile != nil && desc.ModelProfile.Tests != nil {
						for _, testProfile := range desc.ModelProfile.Tests {
							if test, exists := d.TestsMap[testProfile.Name]; exists {
								node.Tests[testProfile.Name] = test
							}
						}
					}
				}
			}

			d.NodeMap[assetName] = node
		}
	}

	// Second pass: Connect nodes with pointers based on upstream/downstream relationships
	for assetName, node := range d.NodeMap {
		asset := d.AssetsMap[assetName]

		// Connect to upstream nodes
		for _, upstreamName := range asset.GetUpstreams() {
			if upstreamNode, exists := d.NodeMap[upstreamName]; exists {
				node.Upstreams = append(node.Upstreams, upstreamNode)
				// Also add this node as downstream to the upstream node
				upstreamNode.Downstreams = append(upstreamNode.Downstreams, node)
			} else {
				log.Warn().
					Str("assetName", assetName).
					Str("upstream", upstreamName).
					Msg("Upstream node not found")
			}
		}

		// Identify root nodes (no upstreams)
		if len(node.Upstreams) == 0 {
			d.RootNodes = append(d.RootNodes, node)
		}

		// Identify leaf nodes (no downstreams explicitly set in asset)
		if len(asset.GetDownstreams()) == 0 {
			d.LeafNodes = append(d.LeafNodes, node)
		}
	}

	log.Info().
		Int("totalNodes", len(d.NodeMap)).
		Int("rootNodes", len(d.RootNodes)).
		Int("leafNodes", len(d.LeafNodes)).
		Msg("Debug DAG built with pointer structure")
}

// Run implements DAG.Run - Debug DAG doesn't actually execute
func (d *DebugDag) Run() *sync.WaitGroup {
	var wg sync.WaitGroup
	// Debug DAG is for visualization/debugging, not execution
	log.Debug().Msg("DebugDag.Run() called - no execution in debug mode")
	return &wg
}

// Push implements DAG.Push - Executes assets sequentially according to dagGraph
func (d *DebugDag) Push(taskId string, data interface{}, resultChan chan map[string]interface{}) chan map[string]interface{} {
	taskUUID := uuid.New().String()
	log.Info().Str("taskId", taskId).Str("taskUUID", taskUUID).Msg("DebugDag.Push() starting sequential execution")

	// Store the taskUUID mapping
	d.mu.Lock()
	d.TaskUUIDMap[taskId] = taskUUID
	d.mu.Unlock()

	// Execute in a goroutine to not block.
	// execMu serializes concurrent Push calls so per-node fields are not
	// written by two executions at once. mu is taken briefly around each
	// state mutation so HTTP readers (GetNodeStates, GetNode, etc.) can
	// observe progress without waiting for the whole DAG run.
	go func() {
		d.execMu.Lock()
		defer d.execMu.Unlock()

		// Reset all node states and results.
		d.mu.Lock()
		for _, node := range d.NodeMap {
			node.State = NodeStateInitial
			node.LastResult = nil
			node.LastError = nil
			node.LastExecutionDuration = 0
			node.LastTestsDuration = 0
			node.TestsPassed = 0
			node.TestsFailed = 0
			node.TestResults = nil
			node.StartTime = nil
			node.EndTime = nil
		}
		if data != nil {
			for _, rootNode := range d.RootNodes {
				rootNode.LastResult = data
			}
		}
		d.mu.Unlock()

		// Execute assets according to dagGraph order (level by level)
		for levelIdx, taskGroup := range d.DagGraph {
			log.Info().Str("taskId", taskId).Str("taskUUID", taskUUID).Int("level", levelIdx).Int("tasks", len(taskGroup)).Msg("Executing DAG level")

			for _, assetName := range taskGroup {
				d.mu.RLock()
				node, exists := d.NodeMap[assetName]
				d.mu.RUnlock()
				if !exists {
					log.Error().Caller().Str("taskId", taskId).Str("taskUUID", taskUUID).Str("assetName", assetName).Msg("Asset not found in NodeMap")
					continue
				}

				// Snapshot upstream results into a local input map.
				d.mu.RLock()
				inputData := make(map[string]interface{})
				for _, upstream := range node.Upstreams {
					if upstream.LastResult != nil {
						inputData[upstream.Name] = upstream.LastResult
					}
				}
				d.mu.RUnlock()

				// If this is a root node with initial data, use it
				if len(node.Upstreams) == 0 && data != nil {
					inputData["__input__"] = data
				}

				log.Info().Str("taskId", taskId).Str("taskUUID", taskUUID).Str("assetName", assetName).Msg("Executing asset")
				startTime := time.Now()

				d.mu.Lock()
				node.State = NodeStateInProgress
				node.StartTime = &startTime
				d.mu.Unlock()

				ctx := &processing.TaskContext{
					TaskID:       taskId,
					TaskUUID:     taskUUID,
					InstanceName: d.DagInstanceName,
					InstanceUUID: d.DagInstanceUUID,
					Input:        inputData,
				}

				// Asset.Execute may run DB queries — do NOT hold d.mu here.
				result, err := node.Asset.Execute(ctx)
				endTime := time.Now()
				execDuration := endTime.Sub(startTime).Milliseconds()

				d.mu.Lock()
				node.EndTime = &endTime
				node.LastExecutionDuration = execDuration
				if err != nil {
					node.State = NodeStateFailed
					node.LastError = err
				} else {
					node.LastResult = result
					node.State = NodeStateSuccess
				}
				d.mu.Unlock()

				if err != nil {
					log.Error().Caller().
						Str("taskId", taskId).
						Str("assetName", assetName).
						Int64("durationMs", execDuration).
						Err(err).
						Msg("Asset execution failed")
					continue
				}

				log.Info().
					Str("taskId", taskId).
					Str("assetName", assetName).
					Int64("durationMs", execDuration).
					Msg("Asset executed successfully")

				// Run tests if configured. RunTests may run DB queries —
				// do NOT hold d.mu across it.
				if len(node.Tests) > 0 && d.TestsMap != nil {
					d.mu.Lock()
					node.State = NodeStateTesting
					node.TestsPassed = 0
					node.TestsFailed = 0
					d.mu.Unlock()

					log.Info().Str("taskId", taskId).Str("taskUUID", taskUUID).Str("assetName", assetName).Int("tests", len(node.Tests)).Msg("Running tests")

					testStartTime := time.Now()
					testResults := node.Asset.RunTests(ctx, d.TestsMap)
					testEndTime := time.Now()
					testDuration := testEndTime.Sub(testStartTime).Milliseconds()

					passed := 0
					failed := 0
					for _, testResult := range testResults {
						switch testResult.Status {
						case processing.TestStatusSuccess:
							passed++
							log.Info().
								Str("taskId", taskId).
								Str("taskUUID", taskUUID).
								Str("assetName", assetName).
								Str("testName", testResult.TestName).
								Int64("durationMs", testResult.DurationMs).
								Msg("Test passed")
						case processing.TestStatusFailed:
							failed++
							log.Warn().
								Str("taskId", taskId).
								Str("taskUUID", taskUUID).
								Str("assetName", assetName).
								Str("testName", testResult.TestName).
								Err(testResult.Error).
								Int64("durationMs", testResult.DurationMs).
								Msg("Test failed")
						case processing.TestStatusNotFound:
							log.Warn().
								Str("taskId", taskId).
								Str("taskUUID", taskUUID).
								Str("assetName", assetName).
								Str("testName", testResult.TestName).
								Str("message", testResult.Message).
								Msg("Test not found")
						}
					}

					d.mu.Lock()
					node.TestResults = testResults
					node.TestsPassed = passed
					node.TestsFailed = failed
					node.LastTestsDuration = testDuration
					if failed > 0 {
						node.State = NodeStateTestsFailed
					} else {
						node.State = NodeStateSuccess
					}
					d.mu.Unlock()

					if failed > 0 {
						var failedTestNames []string
						for _, tr := range testResults {
							if tr.Status == processing.TestStatusFailed {
								failedTestNames = append(failedTestNames, tr.TestName)
							}
						}
						log.Warn().
							Str("taskId", taskId).
							Str("assetName", assetName).
							Int("failed", failed).
							Int("passed", passed).
							Strs("failedTests", failedTestNames).
							Int64("testDurationMs", testDuration).
							Msg("Some tests failed")
					} else {
						log.Info().
							Str("taskId", taskId).
							Str("assetName", assetName).
							Int("passed", passed).
							Int64("testDurationMs", testDuration).
							Msg("All tests passed")
					}
				}
			}
		}

		// Collect results from leaf nodes
		finalResults := make(map[string]interface{})
		d.mu.RLock()
		for _, leafNode := range d.LeafNodes {
			if leafNode.State == NodeStateSuccess && leafNode.LastResult != nil {
				finalResults[leafNode.Name] = leafNode.LastResult
			}
		}
		d.mu.RUnlock()

		// Execute root tests after all DAG tasks are complete
		if d.TestsMap != nil {
			log.Info().Str("taskId", taskId).Str("taskUUID", taskUUID).Msg("Executing root tests")

			d.mu.Lock()
			d.RootTestResults = d.RootTestResults[:0]
			d.mu.Unlock()

			for testName, testCase := range d.TestsMap {
				// Only run tests with "root." prefix
				if len(testName) >= 5 && testName[:5] == "root." {
					rootCtx := &processing.TaskContext{
						TaskID:       taskId,
						TaskUUID:     taskUUID,
						InstanceName: d.DagInstanceName,
						InstanceUUID: d.DagInstanceUUID,
					}
					startTime := time.Now()
					// testCase.Execute runs DB queries — no lock held.
					status, executedTestName, err := testCase.Execute(rootCtx)
					duration := time.Since(startTime).Milliseconds()

					testResult := processing.TestResult{
						TestName:   executedTestName,
						DurationMs: duration,
					}

					if status {
						testResult.Status = processing.TestStatusSuccess
						testResult.Message = "Root test passed"
						log.Info().Str("taskId", taskId).Str("taskUUID", taskUUID).Str("testName", executedTestName).Int64("durationMs", duration).Msg("Root test passed")
					} else {
						testResult.Status = processing.TestStatusFailed
						testResult.Error = err
						testResult.Message = "Root test failed"
						log.Error().Str("taskId", taskId).Str("taskUUID", taskUUID).Str("testName", executedTestName).Int64("durationMs", duration).Err(err).Msg("Root test failed")
					}

					d.mu.Lock()
					d.RootTestResults = append(d.RootTestResults, testResult)
					d.mu.Unlock()
				}
			}
		}

		// Send results back through the channel
		select {
		case resultChan <- finalResults:
			log.Info().Str("taskId", taskId).Str("taskUUID", taskUUID).Msg("Results sent to result channel")
		default:
			log.Warn().Str("taskId", taskId).Str("taskUUID", taskUUID).Msg("Result channel not ready, results not sent")
		}
	}()

	return resultChan
}

// Stop implements DAG.Stop
func (d *DebugDag) Stop() {
	log.Debug().Msg("DebugDag stopped")
}

// GetTaskUUID returns the UUID for a given taskId
func (d *DebugDag) GetTaskUUID(taskId string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.TaskUUIDMap[taskId]
}

// GetOrCreateTaskUUID returns the existing UUID for a taskId, or creates a new one if it doesn't exist
func (d *DebugDag) GetOrCreateTaskUUID(taskId string) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if UUID already exists
	if taskUUID, exists := d.TaskUUIDMap[taskId]; exists {
		return taskUUID
	}

	// Generate new UUID and store it
	taskUUID := uuid.New().String()
	d.TaskUUIDMap[taskId] = taskUUID

	log.Debug().
		Str("taskId", taskId).
		Str("taskUUID", taskUUID).
		Msg("Generated new TaskUUID")

	return taskUUID
}

// GetNode returns a debug service node by name
func (d *DebugDag) GetNode(name string) *DagAssetDebugService {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.NodeMap[name]
}

// UpdateNodeState updates the state of a node
func (d *DebugDag) UpdateNodeState(name string, state NodeState) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if node, exists := d.NodeMap[name]; exists {
		node.State = state
		log.Debug().
			Str("node", name).
			Str("state", string(state)).
			Msg("Node state updated")
	}
}

// UpdateNodeTestResults updates test results for a node
func (d *DebugDag) UpdateNodeTestResults(name string, passed, failed int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if node, exists := d.NodeMap[name]; exists {
		node.TestsPassed = passed
		node.TestsFailed = failed
		log.Debug().
			Str("node", name).
			Int("passed", passed).
			Int("failed", failed).
			Msg("Node test results updated")
	}
}

// GetNodeStates returns current state of all nodes in the DAG
func (d *DebugDag) GetNodeStates() map[string]NodeState {
	d.mu.RLock()
	defer d.mu.RUnlock()

	states := make(map[string]NodeState)
	for name, node := range d.NodeMap {
		states[name] = node.State
	}
	return states
}

// GetNodeResults returns the execution results of all nodes
func (d *DebugDag) GetNodeResults() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	results := make(map[string]interface{})
	for name, node := range d.NodeMap {
		if node.LastResult != nil {
			results[name] = node.LastResult
		}
	}
	return results
}

// Connect establishes all database connections
func (d *DebugDag) Connect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	log.Info().Msg("Connecting to all databases")
	core.GetInstance().ConnectAll()
	d.isConnected = true
	log.Info().Msg("All database connections established")
	return nil
}

// Disconnect closes all database connections
func (d *DebugDag) Disconnect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	log.Info().Msg("Disconnecting from all databases")
	core.GetInstance().Shutdown()
	d.isConnected = false
	log.Info().Msg("All database connections closed")
	return nil
}

// IsConnected returns the current connection status
func (d *DebugDag) IsConnected() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.isConnected
}

// StoreTestResult stores a test execution result
func (d *DebugDag) StoreTestResult(taskId, testName string, result *TestExecutionResult) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.TestExecutionMap[taskId] == nil {
		d.TestExecutionMap[taskId] = make(map[string]*TestExecutionResult)
	}
	d.TestExecutionMap[taskId][testName] = result

	log.Debug().
		Str("taskId", taskId).
		Str("testName", testName).
		Str("status", result.Status).
		Int("rowCount", result.RowCount).
		Msg("Test result stored")
}

// GetTestResult retrieves a test execution result
func (d *DebugDag) GetTestResult(taskId, testName string) *TestExecutionResult {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if taskTests, exists := d.TestExecutionMap[taskId]; exists {
		return taskTests[testName]
	}
	return nil
}

// NodeRuntimeState is a thread-safe snapshot of a node's mutable execution
// state. Returned by NodeRuntimeStateFor; callers may read the fields
// without holding any lock.
type NodeRuntimeState struct {
	State                 NodeState
	TestsPassed           int
	TestsFailed           int
	LastExecutionDuration int64
	LastTestsDuration     int64
	TestResults           []processing.TestResult
}

// NodeRuntimeStateFor returns a copy of the named node's mutable fields,
// taken under d.mu. Returns ok=false if the node does not exist.
func (d *DebugDag) NodeRuntimeStateFor(name string) (NodeRuntimeState, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	n, ok := d.NodeMap[name]
	if !ok {
		return NodeRuntimeState{}, false
	}
	snap := NodeRuntimeState{
		State:                 n.State,
		TestsPassed:           n.TestsPassed,
		TestsFailed:           n.TestsFailed,
		LastExecutionDuration: n.LastExecutionDuration,
		LastTestsDuration:     n.LastTestsDuration,
	}
	if len(n.TestResults) > 0 {
		snap.TestResults = make([]processing.TestResult, len(n.TestResults))
		copy(snap.TestResults, n.TestResults)
	}
	return snap, true
}

// GetRootTestResults returns a copy of the most recent root-test results,
// taken under d.mu.
func (d *DebugDag) GetRootTestResults() []processing.TestResult {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if len(d.RootTestResults) == 0 {
		return nil
	}
	out := make([]processing.TestResult, len(d.RootTestResults))
	copy(out, d.RootTestResults)
	return out
}

// ResetAllNodes clears per-node execution state and root-test results under
// d.mu. Callers that previously mutated NodeMap entries directly should use
// this instead.
func (d *DebugDag) ResetAllNodes() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, node := range d.NodeMap {
		node.State = NodeStateInitial
		node.LastResult = nil
		node.LastError = nil
		node.LastExecutionDuration = 0
		node.LastTestsDuration = 0
		node.TestsPassed = 0
		node.TestsFailed = 0
		node.TestResults = nil
		node.StartTime = nil
		node.EndTime = nil
	}
	d.RootTestResults = nil
}
