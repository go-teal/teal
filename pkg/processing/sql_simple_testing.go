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

func (mt *SQLModelTestCase) Execute() (bool, string, error) {
	dbConnection := core.GetInstance().GetDBConnection(mt.descriptor.TestProfile.Connection)

	if !dbConnection.IsPermanent() {
		err := dbConnection.Connect()
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to connect to database")
			defer dbConnection.Close()
			return false, "", err
		}
		defer dbConnection.Close()
	}

	sqlTestTemplate, err := template.New("runSQTestTemplate").
		Funcs(FromConnectionContext(dbConnection, nil, mt.descriptor.Name, make(template.FuncMap))).
		Parse(mt.descriptor.CountTestSQL)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("Parsing template: %s", mt.descriptor.CountTestSQL)
		return false, mt.descriptor.Name, err
	}

	var sqlQuery bytes.Buffer

	err = sqlTestTemplate.Execute(&sqlQuery, nil)

	if err != nil {
		log.Error().Stack().Err(err).Msgf("Execute template: %s", mt.descriptor.CountTestSQL)
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
