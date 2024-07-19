package models

import "github.com/go-teal/teal/pkg/configs"

type SQLModelDescriptor struct {
	Name             string
	CreateTableSQL   string
	CreateViewSQL    string
	InsertSQL        string
	RawSQL           string
	DropTableSQL     string
	DropViewSQL      string
	TruncateTableSQL string
	Upstreams        []string
	Downstreams      []string
	ModelProfile     *configs.ModelProfile
}

type SQLModelTestDescriptor struct {
	Name         string
	RawSQL       string
	CountTestSQL string
	TestProfile  *configs.TestProfile
}
