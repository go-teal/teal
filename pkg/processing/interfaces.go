package processing

type Asset interface {
	Execute(input map[string]interface{}) (interface{}, error)
	GetUpstreams() []string
	GetDownstreams() []string
	GetName() string
}

type ModelTesting interface {
	Execute() (bool, string, error)
}
