package dags

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
	"github.com/rs/zerolog/log"
)

const (
	STOP_TASK_ID = "STOP"
)

type DagRoutine struct {
	Name           string
	dag            *ChannelDag
	Asset          processing.Asset
	InputChannels  map[string]chan *TransitionTask
	OutPutChannels map[string]chan *TransitionTask
	config         *configs.Config
	testsMap       map[string]processing.ModelTesting
}

type ChannelDag struct {
	DagInstanceName      string
	dagGrpah             [][]string
	assetsMap            map[string]processing.Asset
	testsMap             map[string]processing.ModelTesting
	dagRoutineMap        map[string]*DagRoutine
	config               *configs.Config
	completeTasksResults map[string]*taskResult
	numberOfFinalTasks   int
}

type taskResult struct {
	results              map[string]interface{}
	resultChan           chan map[string]interface{}
	remainingTasksNumber *atomic.Int32
}

// Send implements DAG.

func InitChannelDag(dagGrpah [][]string, assetsMap map[string]processing.Asset, config *configs.Config, name string) DAG {
	dag := ChannelDag{
		DagInstanceName:      name,
		dagGrpah:             dagGrpah,
		assetsMap:            assetsMap,
		dagRoutineMap:        make(map[string]*DagRoutine, len(assetsMap)),
		completeTasksResults: make(map[string]*taskResult),
		config:               config,
	}
	dag.build()
	return &dag
}

func InitChannelDagWithTests(dagGrpah [][]string,
	assetsMap map[string]processing.Asset,
	testsMap map[string]processing.ModelTesting,
	config *configs.Config,
	name string) DAG {
	dag := ChannelDag{
		DagInstanceName:      name,
		dagGrpah:             dagGrpah,
		assetsMap:            assetsMap,
		testsMap:             testsMap,
		dagRoutineMap:        make(map[string]*DagRoutine, len(assetsMap)),
		completeTasksResults: make(map[string]*taskResult),
		config:               config,
	}
	dag.build()
	return &dag
}

// build implements DAG.
func (dag *ChannelDag) build() {
	dag.numberOfFinalTasks = 0
	for _, taskGroup := range dag.dagGrpah {
		for _, task := range taskGroup {
			if len(dag.assetsMap[task].GetDownstreams()) == 0 {
				dag.numberOfFinalTasks++
			}
			dag.dagRoutineMap[task] = initDagRoutine(dag, dag.assetsMap[task], dag.dagRoutineMap, dag.testsMap, dag.config)
		}
	}
}

// Run implements DAG.
func (dag *ChannelDag) Run() *sync.WaitGroup {
	var wg sync.WaitGroup
	for _, taskGroup := range dag.dagGrpah {
		for _, task := range taskGroup {
			wg.Add(1)
			go dag.dagRoutineMap[task].run(&wg)
		}
	}
	return &wg
}

type TransitionTask struct {
	TaskID       string
	Data         interface{}
	StopSignal   bool
	IngoreSignal bool
}

func initDagRoutine(dag *ChannelDag,
	asset processing.Asset,
	channelAssets map[string]*DagRoutine,
	testsMap map[string]processing.ModelTesting,
	config *configs.Config) *DagRoutine {
	channelAsset := DagRoutine{
		dag:      dag,
		Name:     asset.GetName(),
		Asset:    asset,
		config:   config,
		testsMap: testsMap,
	}

	downstreams := asset.GetDownstreams()
	if len(downstreams) != 0 {
		channelAsset.OutPutChannels = make(map[string]chan *TransitionTask, len(downstreams))
		for _, downstreamName := range asset.GetDownstreams() {
			channelAsset.OutPutChannels[downstreamName] = make(chan *TransitionTask, config.Cores)
		}
	}
	upstreams := asset.GetUpstreams()
	if len(upstreams) == 0 {
		channelAsset.InputChannels = map[string]chan *TransitionTask{
			"start": make(chan *TransitionTask, config.Cores),
		}
	} else {
		channelAsset.InputChannels = make(map[string]chan *TransitionTask, len(upstreams))
		for _, upstreamName := range asset.GetUpstreams() {
			channelAsset.InputChannels[upstreamName] = channelAssets[upstreamName].OutPutChannels[channelAsset.Name]
		}
	}
	return &channelAsset
}

func (dag *ChannelDag) Push(taskId string, data interface{}, resultChan chan map[string]interface{}) chan map[string]interface{} {

	if resultChan != nil {
		dag.completeTasksResults[taskId] = &taskResult{
			results:              make(map[string]interface{}),
			remainingTasksNumber: new(atomic.Int32),
			resultChan:           resultChan,
		}
		dag.completeTasksResults[taskId].remainingTasksNumber.Store(int32(dag.numberOfFinalTasks))
	}

	log.Debug().Str("DAG", dag.DagInstanceName).Str("taskId", taskId).Int("results", dag.numberOfFinalTasks).Msg("New task has been registred")
	for _, assetName := range dag.dagGrpah[0] {
		routine := dag.dagRoutineMap[assetName]
		dag.propagateTask(taskId, "", false, false, routine.InputChannels, data)
	}
	return resultChan
}

func (dag *ChannelDag) Stop() {
	log.Debug().Str("DAG", dag.DagInstanceName).Str("taskId", STOP_TASK_ID).Int("results", dag.numberOfFinalTasks).Msg("Stop task has been registred")

	for _, assetName := range dag.dagGrpah[0] {
		routine := dag.dagRoutineMap[assetName]
		dag.propagateTask(STOP_TASK_ID, "", true, false, routine.InputChannels, nil)
	}
}

func (routine *DagRoutine) run(wg *sync.WaitGroup) {
	log.Info().
		Str("DAG", routine.dag.DagInstanceName).
		Str("assetName", routine.Name).
		Msg("Starting routine")

	defer wg.Done()
	for {
		params := make(map[string]interface{}, len(routine.InputChannels))
		var ignore bool
		var taskId string
		for channelName, inputChannel := range routine.InputChannels {
			inputTask := <-inputChannel
			log.Debug().
				Str("DAG", routine.dag.DagInstanceName).
				Str("channelName", channelName).
				Str("assetName", routine.Name).
				Str("taskId", inputTask.TaskID).Msg("task received")
			if inputTask.StopSignal {
				routine.dag.propagateTask(inputTask.TaskID, routine.Name, true, true, routine.OutPutChannels, nil)
				log.Debug().
					Str("DAG", routine.dag.DagInstanceName).
					Str("channelName", channelName).
					Str("taskId", inputTask.TaskID).
					Str("assetName", routine.Name).Msg("Stop signal received")
				return
			}
			if inputTask.IngoreSignal {
				ignore = true
			}
			params[channelName] = inputTask.Data
			taskId = inputTask.TaskID
		}

		if !ignore {
			log.Debug().
				Str("DAG", routine.dag.DagInstanceName).
				Str("assetName", routine.Name).
				Str("taskId", taskId).
				Msg("Executing asset...")
			startTaskTs := time.Now().UnixMilli()
			outputData, err := routine.Asset.Execute(params)
			stopTaskTs := time.Now().UnixMilli()
			if err != nil {
				log.Error().Caller().Stack().
					Str("DAG", routine.dag.DagInstanceName).
					Str("assetName", routine.Name).
					Str("taskId", taskId).
					Float64("durationSec", float64(stopTaskTs-startTaskTs)/1000.0).
					Err(err).
					Msg("Asset Error")
				routine.dag.propagateTask(taskId, routine.Name, false, true, routine.OutPutChannels, nil)
			} else {
				if outputData != nil {
					log.Debug().
						Str("DAG", routine.dag.DagInstanceName).
						Str("assetName", routine.Name).
						Str("taskId", taskId).
						Msgf("Complete with data: %v", outputData)
				}
				testResults := routine.Asset.RunTests(routine.testsMap)

				// Log test results if any tests were run
				if len(testResults) > 0 {
					passed := 0
					failed := 0
					for _, result := range testResults {
						if result.Status == processing.TestStatusSuccess {
							passed++
						} else if result.Status == processing.TestStatusFailed {
							failed++
						}
					}
					log.Info().
						Str("DAG", routine.dag.DagInstanceName).
						Str("assetName", routine.Name).
						Str("taskId", taskId).
						Int("totalTests", len(testResults)).
						Int("passed", passed).
						Int("failed", failed).
						Msg("Tests executed")
				}

				stopTaskTs := time.Now().UnixMilli()
				log.Info().
					Str("DAG", routine.dag.DagInstanceName).
					Str("assetName", routine.Name).
					Str("taskId", taskId).
					Float64("durationSec", float64(stopTaskTs-startTaskTs)/1000.0).
					Msg("Asset complete")
				routine.dag.propagateTask(taskId, routine.Name, false, false, routine.OutPutChannels, outputData)
			}
		} else {
			log.Warn().
				Str("DAG", routine.dag.DagInstanceName).
				Str("assetName", routine.Name).
				Str("taskId", taskId).
				Msg("Task has been ingored")
			routine.dag.propagateTask(taskId, routine.Name, false, true, routine.OutPutChannels, nil)
		}
	}

}

func (dag *ChannelDag) propagateTask(taskId string, assetName string, stop bool, ingore bool, channels map[string]chan *TransitionTask, data interface{}) {

	if channels == nil {
		log.Debug().
			Str("DAG", dag.DagInstanceName).
			Str("assetName", assetName).
			Str("taskId", taskId).
			Msg("No next channels found")

		if taskId != "STOP" {
			resultTask, ok := dag.completeTasksResults[taskId]
			if ok {
				resultTask.remainingTasksNumber.Add(-1)
				resultTask.results[assetName] = data
				if int(resultTask.remainingTasksNumber.Load()) == 0 {
					log.Debug().
						Str("DAG", dag.DagInstanceName).
						Str("assetName", assetName).
						Str("taskId", taskId).
						Msg("Task complete")
					resultTask.resultChan <- resultTask.results
					delete(dag.completeTasksResults, taskId)
				}
			}
		}
		return
	}

	for _, output := range channels {
		output <- &TransitionTask{
			TaskID:       taskId,
			StopSignal:   stop,
			IngoreSignal: ingore,
			Data:         data,
		}
		if stop {
			close(output)
		}
	}
}
