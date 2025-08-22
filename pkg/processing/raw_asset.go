package processing

import (
	"fmt"
	"sync"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/models"
	"github.com/rs/zerolog/log"
)

type ExecutorFunc func(input map[string]interface{}, modelProfile *configs.ModelProfile) (interface{}, error)

// GO singletone
type GlobalExecutors struct {
	Execurots map[string]ExecutorFunc
}

var globalExecutors *GlobalExecutors
var once sync.Once

func GetExecutors() *GlobalExecutors {
	once.Do(func() {

		globalExecutors = &GlobalExecutors{
			Execurots: make(map[string]ExecutorFunc),
		}
	})
	return globalExecutors
}

type RawModelAsset struct {
	descriptor *models.RawModelDescriptor
}

// Execute implements Asset.
func (r *RawModelAsset) Execute(input map[string]interface{}) (interface{}, error) {
	if f, ok := GetExecutors().Execurots[r.descriptor.Name]; ok {
		return f(input, r.descriptor.ModelProfile)
	} else {
		return nil, fmt.Errorf("executor %v is not registered", r.descriptor.Name)
	}

}

// GetDownstreams implements Asset.
func (r *RawModelAsset) GetDownstreams() []string {
	return r.descriptor.Downstreams
}

// GetName implements Asset.
func (r *RawModelAsset) GetName() string {
	return r.descriptor.Name
}

// GetDescriptor implements Asset.
func (r *RawModelAsset) GetDescriptor() any {
	return r.descriptor
}

// GetUpstreams implements Asset.
func (r *RawModelAsset) GetUpstreams() []string {
	return r.descriptor.Upstreams
}

// RunTests implements Asset.
func (r *RawModelAsset) RunTests(testsMap map[string]ModelTesting) {
	log.Warn().Msg("Raw Model Asset does not support tests")
}

func InitRawModelAsset(descriptor *models.RawModelDescriptor) Asset {
	return &RawModelAsset{
		descriptor: descriptor,
	}
}
