package internalmodels

import (
	"bytes"

	"github.com/go-teal/teal/pkg/configs"
)

type ModelType int

const (
	DATABASE ModelType = iota
	SOURCE
	CUSTOM
)

type ModelConfig struct {
	GoName               string
	ModelName            string
	Stage                string
	NameUpperCase        string
	SqlByteBuffer        bytes.Buffer
	Config               *configs.Config
	Profile              *configs.ProjectProfile
	Upstreams            []string
	Downstreams          []string
	Priority             int
	ModelProfile         *configs.ModelProfile
	SourceProfile        *configs.SourceProfile
	ModelType            ModelType
	ModelFieldsFunc      string
	PrimaryKeyExpression string
}
