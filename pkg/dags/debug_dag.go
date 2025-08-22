package dags

import (
	"sync"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

// DebugDag is a debug-enabled DAG implementation that wraps ChannelDag
type DebugDag struct {
	InnerDag         DAG
	DagInstanceName  string
	DagGraph         [][]string
	AssetsMap        map[string]processing.Asset
	TestsMap         map[string]processing.ModelTesting
	Config           *configs.Config
}

// InitDebugDag creates a new DebugDag with the same parameters as InitChannelDagWithTests
func InitDebugDag(dagGraph [][]string,
	assetsMap map[string]processing.Asset,
	testsMap map[string]processing.ModelTesting,
	config *configs.Config,
	name string) *DebugDag {
	
	// Create the inner DAG using InitChannelDagWithTests
	innerDag := InitChannelDagWithTests(dagGraph, assetsMap, testsMap, config, name)
	
	return &DebugDag{
		InnerDag:        innerDag,
		DagInstanceName: name,
		DagGraph:        dagGraph,
		AssetsMap:       assetsMap,
		TestsMap:        testsMap,
		Config:          config,
	}
}

// Run implements DAG.Run
func (d *DebugDag) Run() *sync.WaitGroup {
	// TODO: Add debug functionality
	return d.InnerDag.Run()
}

// Push implements DAG.Push
func (d *DebugDag) Push(taskName string, data interface{}, resultChan chan map[string]interface{}) chan map[string]interface{} {
	// TODO: Add debug functionality
	return d.InnerDag.Push(taskName, data, resultChan)
}

// Stop implements DAG.Stop
func (d *DebugDag) Stop() {
	// TODO: Add debug functionality
	d.InnerDag.Stop()
}