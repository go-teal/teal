package processing

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/models"
	"github.com/rs/zerolog/log"
)

type SQLModelTestCase struct {
	descriptor *models.SQLModelTestDescriptor
	functions  template.FuncMap
}

func InitSQLModelTesting(descriptor *models.SQLModelTestDescriptor) ModelTesting {
	return &SQLModelTestCase{
		descriptor: descriptor,
		functions:  make(template.FuncMap),
	}
}

func (mt *SQLModelTestCase) Execute(ctx *TaskContext) (bool, string, error) {

	dbConnection := core.GetInstance().GetDBConnection(mt.descriptor.TestProfile.Connection)

	sqlTestTemplate, err := template.New("runSQTestTemplate").
		Funcs(MergeTemplateFuncs(
			FromConnectionContext(dbConnection, nil, mt.descriptor.Name, make(template.FuncMap)),
			FromTaskContext(ctx),
		)).
		Parse(mt.descriptor.CountTestSQL)
	if err != nil {
		log.Error().Caller().Stack().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Err(err).Str("sql", mt.descriptor.CountTestSQL).Msg("Failed to parse test SQL template")
		return false, mt.descriptor.Name, err
	}

	var sqlQuery bytes.Buffer

	err = sqlTestTemplate.Execute(&sqlQuery, nil)

	if err != nil {
		log.Error().Caller().Stack().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Err(err).Str("sql", mt.descriptor.CountTestSQL).Msg("Failed to execute test SQL template")
		return false, mt.descriptor.Name, err
	}

	msg, err := dbConnection.SimpleTest(sqlQuery.String())

	if err != nil {
		return false, mt.descriptor.Name, err
	}

	if msg != "" {
		return false, mt.descriptor.Name, fmt.Errorf("count test failed: %s", msg)
	}

	return true, mt.descriptor.Name, nil
}

// GetDescriptor implements ModelTesting.
func (mt *SQLModelTestCase) GetDescriptor() any {
	return mt.descriptor
}
