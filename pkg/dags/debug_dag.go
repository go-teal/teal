package dags

import (
	"sync"
	"time"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/processing"
	"github.com/rs/zerolog/log"
)

// NodeState represents the execution state of a node in the debug DAG
type NodeState string

const (
	NodeStateInitial    NodeState = "INITIAL"
	NodeStateInProgress NodeState = "IN_PROGRESS"
	NodeStateTesting    NodeState = "TESTING"
	NodeStateFailed     NodeState = "FAILED"
	NodeStateSuccess    NodeState = "SUCCESS"
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

// DebugDag is a debug-enabled DAG implementation using pointers instead of channels
type DebugDag struct {
	InnerDag        DAG // Keep for compatibility but will be nil
	DagInstanceName string
	DagGraph        [][]string
	AssetsMap       map[string]processing.Asset
	TestsMap        map[string]processing.ModelTesting
	Config          *configs.Config
	NodeMap         map[string]*DagAssetDebugService // Map of asset name to debug service node
	RootNodes       []*DagAssetDebugService          // Nodes with no upstreams
	LeafNodes       []*DagAssetDebugService          // Nodes with no downstreams
	mu              sync.RWMutex                     // Mutex for thread-safe access
}

// InitDebugDag creates a new DebugDag with pointer-based structure
func InitDebugDag(dagGraph [][]string,
	assetsMap map[string]processing.Asset,
	testsMap map[string]processing.ModelTesting,
	config *configs.Config,
	name string) *DebugDag {

	dag := &DebugDag{
		InnerDag:        nil, // No inner DAG for debug mode
		DagInstanceName: name,
		DagGraph:        dagGraph,
		AssetsMap:       assetsMap,
		TestsMap:        testsMap,
		Config:          config,
		NodeMap:         make(map[string]*DagAssetDebugService),
		RootNodes:       make([]*DagAssetDebugService, 0),
		LeafNodes:       make([]*DagAssetDebugService, 0),
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
	log.Info().Str("taskId", taskId).Msg("DebugDag.Push() starting sequential execution")

	// Execute in a goroutine to not block
	go func() {
		core.GetInstance().ConnectAll()
		defer core.GetInstance().Shutdown()
		d.mu.Lock()
		defer d.mu.Unlock()

		// Reset all node states and results
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

		// Store initial input data if provided
		if data != nil {
			// Store initial data as a special "input" result that all root nodes can access
			for _, rootNode := range d.RootNodes {
				rootNode.LastResult = data
			}
		}

		// Execute assets according to dagGraph order (level by level)
		for levelIdx, taskGroup := range d.DagGraph {
			log.Info().Str("taskId", taskId).Int("level", levelIdx).Int("tasks", len(taskGroup)).Msg("Executing DAG level")

			for _, assetName := range taskGroup {
				node, exists := d.NodeMap[assetName]
				if !exists {
					log.Error().Caller().Str("taskId", taskId).Str("asset", assetName).Msg("Asset not found in NodeMap")
					continue
				}

				// Prepare input data from upstream results
				inputData := make(map[string]interface{})
				for _, upstream := range node.Upstreams {
					if upstream.LastResult != nil {
						inputData[upstream.Name] = upstream.LastResult
					}
				}

				// If this is a root node with initial data, use it
				if len(node.Upstreams) == 0 && data != nil {
					inputData["__input__"] = data
				}

				// Execute the asset
				log.Info().Str("taskId", taskId).Str("asset", assetName).Msg("Executing asset")
				node.State = NodeStateInProgress

				startTime := time.Now()
				node.StartTime = &startTime
				result, err := node.Asset.Execute(inputData)
				endTime := time.Now()
				node.EndTime = &endTime
				node.LastExecutionDuration = endTime.Sub(startTime).Milliseconds()

				if err != nil {
					node.State = NodeStateFailed
					node.LastError = err
					log.Error().Caller().
						Str("taskId", taskId).
						Str("asset", assetName).
						Int64("durationMs", node.LastExecutionDuration).
						Err(err).
						Msg("Asset execution failed")
					continue
				}

				// Store the result in the node
				node.LastResult = result
				node.State = NodeStateSuccess
				log.Info().
					Str("taskId", taskId).
					Str("asset", assetName).
					Int64("durationMs", node.LastExecutionDuration).
					Msg("Asset executed successfully")

				// Run tests using the asset's RunTests method if tests are configured
				if len(node.Tests) > 0 && d.TestsMap != nil {
					node.State = NodeStateTesting
					node.TestsPassed = 0
					node.TestsFailed = 0

					log.Info().Str("taskId", taskId).Str("asset", assetName).Int("tests", len(node.Tests)).Msg("Running tests")

					// Track test execution time
					testStartTime := time.Now()

					// Execute tests and get results
					testResults := node.Asset.RunTests(d.TestsMap)

					// Store test results for later retrieval
					node.TestResults = testResults

					// Process test results
					for _, testResult := range testResults {
						if testResult.Status == processing.TestStatusSuccess {
							node.TestsPassed++
							log.Info().
								Str("taskId", taskId).
								Str("asset", assetName).
								Str("testName", testResult.TestName).
								Int64("durationMs", testResult.DurationMs).
								Msg("Test passed")
						} else if testResult.Status == processing.TestStatusFailed {
							node.TestsFailed++
							log.Warn().
								Str("taskId", taskId).
								Str("asset", assetName).
								Str("testName", testResult.TestName).
								Err(testResult.Error).
								Int64("durationMs", testResult.DurationMs).
								Msg("Test failed")
						} else if testResult.Status == processing.TestStatusNotFound {
							log.Warn().
								Str("taskId", taskId).
								Str("asset", assetName).
								Str("testName", testResult.TestName).
								Str("message", testResult.Message).
								Msg("Test not found")
						}
					}

					// Calculate total test duration
					testEndTime := time.Now()
					node.LastTestsDuration = testEndTime.Sub(testStartTime).Milliseconds()

					// Update final state based on test results
					if node.TestsFailed > 0 {
						node.State = NodeStateFailed
						log.Warn().
							Str("taskId", taskId).
							Str("asset", assetName).
							Int("failed", node.TestsFailed).
							Int("passed", node.TestsPassed).
							Int64("testDurationMs", node.LastTestsDuration).
							Msg("Some tests failed")
					} else {
						node.State = NodeStateSuccess
						log.Info().
							Str("taskId", taskId).
							Str("asset", assetName).
							Int("passed", node.TestsPassed).
							Int64("testDurationMs", node.LastTestsDuration).
							Msg("All tests passed")
					}
				}
			}
		}

		// Collect results from leaf nodes
		finalResults := make(map[string]interface{})
		for _, leafNode := range d.LeafNodes {
			if leafNode.State == NodeStateSuccess && leafNode.LastResult != nil {
				finalResults[leafNode.Name] = leafNode.LastResult
			}
		}

		// Send results back through the channel
		select {
		case resultChan <- finalResults:
			log.Info().Str("taskId", taskId).Msg("Results sent to result channel")
		default:
			log.Warn().Str("taskId", taskId).Msg("Result channel not ready, results not sent")
		}
	}()

	return resultChan
}

// Stop implements DAG.Stop
func (d *DebugDag) Stop() {
	log.Debug().Msg("DebugDag stopped")
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
