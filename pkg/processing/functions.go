package processing

import (
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

// MergePongo2Context merges multiple pongo2.Context maps into one
func MergePongo2Context(contexts ...pongo2.Context) pongo2.Context {
	result := make(pongo2.Context)
	for _, ctx := range contexts {
		for key, value := range ctx {
			result[key] = value
		}
	}
	return result
}

// FromTaskContextPongo2 creates a pongo2.Context with task-related functions
func FromTaskContextPongo2(ctx *TaskContext) pongo2.Context {
	functions := make(pongo2.Context)

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
