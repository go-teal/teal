package dags

import "sync"

type DAG interface {
	Run() *sync.WaitGroup
	Push(taskName string, data interface{}, resultChan chan map[string]interface{}) chan map[string]interface{}
	Stop()
}
