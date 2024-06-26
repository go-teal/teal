package models

import "github.com/go-teal/teal/pkg/configs"

type SQLModelDescriptor struct {
	Name             string
	CreateTableSQL   string
	CreateViewSQL    string
	RunSQL           string
	DropTableSQL     string
	DropViewSQL      string
	TruncateTableSQL string
	Upstreams        []string
	Downstreams      []string
	ModelProfile     *configs.ModelProfile
}
