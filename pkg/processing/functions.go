package processing

import (
	"html/template"
	"os"
	"strings"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/teal/pkg/drivers"
)

func FromConnectionContext(dbConnection drivers.DBDriver, tx interface{}, modelName string, inPlaceFunctions pongo2.Context) pongo2.Context {
	functions := make(pongo2.Context)
	for funcName, f := range inPlaceFunctions {
		functions[funcName] = f
	}

	functions["ModelFields"] = func() string {
		if tx == nil {
			// TODO: fix this temporal solution
			tx, _ = dbConnection.Begin()
			defer dbConnection.Commit(tx)
		}
		fields := dbConnection.GetListOfFields(tx, modelName)
		return strings.Join(fields, ", ")
	}

	functions["ENV"] = func(envName string, defaultValue string) string {
		if envValue := os.Getenv(envName); envValue != "" {
			return envValue
		}
		return defaultValue
	}
	functions["this"] = func() string {
		return modelName
	}
	return functions
}

func MergeTemplateFuncs(funcMaps ...template.FuncMap) template.FuncMap {
	result := make(template.FuncMap)
	for _, funcMap := range funcMaps {
		for key, value := range funcMap {
			result[key] = value
		}
	}
	return result
}

func FromTaskContext(ctx *TaskContext) template.FuncMap {
	functions := make(template.FuncMap)

	functions["TaskID"] = func() string {
		return ctx.TaskID
	}

	functions["TaskUUID"] = func() string {
		return ctx.TaskUUID
	}

	functions["InstanceName"] = func() string {
		return ctx.InstanceName
	}

	functions["InstanceUUID"] = func() string {
		return ctx.InstanceUUID
	}

	return functions
}
