package processing

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/go-teal/gota/dataframe"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/models"

	"github.com/rs/zerolog/log"
)

type SQLModelAsset struct {
	descriptor *models.SQLModelDescriptor
	functions  template.FuncMap
}

func InitSQLModelAsset(descriptor *models.SQLModelDescriptor) Asset {
	return &SQLModelAsset{
		descriptor: descriptor,
		functions:  make(template.FuncMap),
	}
}

// GetName implements Asset.
func (s *SQLModelAsset) GetName() string {
	return s.descriptor.Name
}

// GetDownstrem implements Asset.
func (s *SQLModelAsset) GetDownstreams() []string {
	return s.descriptor.Downstreams
}

// GetUpstream implements Asset.
func (s *SQLModelAsset) GetUpstreams() []string {
	return s.descriptor.Upstreams
}

// Execute implements Asset.
func (s *SQLModelAsset) Execute(input map[string]interface{}) (interface{}, error) {

	var data *dataframe.DataFrame
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	log.Debug().
		Str("s.descriptor.Name", s.descriptor.Name).
		Str("s.descriptor.ModelProfile.Connection", s.descriptor.ModelProfile.Connection).
		Msg("Executing asset")
	log.Debug().Str("s.descriptor.Name", s.descriptor.Name).Msgf("input params: %v", input)

	if !dbConnection.IsPermanent() {
		err := dbConnection.Connect()
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to connect to database")
			defer dbConnection.Close()
			return nil, err
		}
		defer dbConnection.Close()
	}

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return nil, err
	}

	isSchemaExists := dbConnection.CheckSchemaExists(tx, s.descriptor.Name)

	if !isSchemaExists {
		splitted := strings.Split(s.descriptor.Name, ".")
		log.Info().Msgf("Schema %s does not exist", splitted[0])
		err = dbConnection.Exec(tx, fmt.Sprintf("CREATE SCHEMA %s;", splitted[0]))
		if err != nil {
			defer dbConnection.Rallback(tx)
			log.Error().
				Err(err).
				Msg("Failed to create schema")

			return nil, err
		}
	}

	if s.descriptor.ModelProfile.PersistInputs {
		err := s.persistInputs(tx, input)
		if err != nil {
			defer dbConnection.Rallback(tx)
			log.Error().
				Err(err).
				Msg("Failed to persist inputs")
			return nil, err
		}
	}

	isTableExists := dbConnection.CheckTableExists(tx, s.descriptor.Name)

	switch s.descriptor.ModelProfile.Materialization {
	case configs.MAT_INCREMENTAL:

		if s.descriptor.ModelProfile.IsDataFramed {
			data, err = s.GetDataFrame()
			if err != nil {
				return nil, err
			}
		}

		if !isTableExists {
			err = s.createTable(tx)
		} else {
			err = s.insertToTable(tx)
		}

	case configs.MAT_TABLE:
		if s.descriptor.ModelProfile.IsDataFramed {
			log.Warn().Msg("Dataframe can slow this operation, considner custom or incremental materialization")
			data, err = s.GetDataFrame()
			if err != nil {
				return nil, err
			}
		}
		if !isTableExists {
			err = s.createTable(tx)
		} else {
			if err = s.truncateTable(tx); err == nil {
				err = s.insertToTable(tx)
			}
		}
	case configs.MAT_VIEW:
		if s.descriptor.ModelProfile.IsDataFramed {
			log.Warn().Msg("Dataframe can slow this operation, considner custom or incremental materialization")
			data, err = s.GetDataFrame()
			if err != nil {
				return nil, err
			}
		}
		if !isTableExists {
			err = s.createView(tx)
		}
	case configs.MAT_CUSTOM:
		panic("not implemented")
	}

	return data, err
}

// RunTests implements Asset.
func (s *SQLModelAsset) RunTests(testsMap map[string]ModelTesting) {
	log.Info().Msgf("Testing %s", s.descriptor.Name)
	for _, testConfig := range s.descriptor.ModelProfile.Tests {
		if testCase, ok := testsMap[testConfig.Name]; ok {
			status, testName, err := testCase.Execute()
			if status {
				log.Info().Str("Test Case", testName).Msg("Success")
			} else {
				log.Error().Str("Test Case", testName).Err(err).Msg("Failed")
			}
		} else {
			log.Warn().Msgf("Test %s not found", testConfig.Name)
		}
	}
}

func (s *SQLModelAsset) createView(tx interface{}) error {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	var sqlQuery bytes.Buffer
	createViewSQLTemplate, err := template.New("createViewSQLTemplate").
		Funcs(FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions)).
		Parse(s.descriptor.CreateViewSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("parse view SQL %s", s.descriptor.CreateViewSQL)
		return err
	}

	err = createViewSQLTemplate.Execute(&sqlQuery, nil)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("Render view SQL %s", sqlQuery.String())
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery.String())
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msg("Failed to create view")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) createTable(tx interface{}) error {

	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	s.functions["IsIncremental"] = func() bool {
		return false
	}

	createTableSQLTempl, err := template.New("createTableSQLTempl").
		Funcs(FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions)).
		Parse(s.descriptor.CreateTableSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("Parsing template: %s", s.descriptor.CreateTableSQL)
		return err
	}

	var sqlQuery bytes.Buffer

	err = createTableSQLTempl.Execute(&sqlQuery, nil)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("Executing template: %s", sqlQuery.String())
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery.String())
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msg("Failed to create table")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) truncateTable(tx interface{}) error {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	err := dbConnection.Exec(tx, s.descriptor.TruncateTableSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msg("Failed to run truncate")
		return err
	}
	return err
}

func (s *SQLModelAsset) insertToTable(tx interface{}) error {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	var sqlQuery bytes.Buffer
	s.functions["IsIncremental"] = func() bool {
		return s.descriptor.ModelProfile.Materialization == configs.MAT_INCREMENTAL
	}

	runSQLTemplate, err := template.New("Insert SQL template").
		Funcs(FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions)).
		Parse(s.descriptor.InsertSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("Parsing template: %s", s.descriptor.InsertSQL)
		return err
	}

	err = runSQLTemplate.Execute(&sqlQuery, nil)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("Rendering template: %s", sqlQuery.String())
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery.String())
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msg("Failed to run incremental")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) GetDataFrame() (*dataframe.DataFrame, error) {

	s.functions["IsIncremental"] = func() bool {
		return s.descriptor.ModelProfile.Materialization == configs.MAT_INCREMENTAL
	}

	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	simleSQLQueryTemplate, err := template.New("simleSQLQueryTemplate").
		Funcs(FromConnectionContext(dbConnection, nil, s.descriptor.Name, s.functions)).
		Parse(s.descriptor.RawSQL)
	if err != nil {
		log.Error().Stack().
			Err(err).
			Msg("Failed to parse asset query")
		return nil, err
	}
	var sqlQuery bytes.Buffer
	err = simleSQLQueryTemplate.Execute(&sqlQuery, nil)
	if err != nil {
		log.Error().Stack().
			Err(err).
			Msgf("Rendering template: %s", sqlQuery.String())
		return nil, err
	}

	data, err := dbConnection.ToDataFrame(sqlQuery.String())
	if err != nil {
		log.Error().Stack().
			Err(err).
			Msgf("Failed to create a DataFrame for: %s", sqlQuery.String())
		return nil, err
	}
	return data, nil
}

func (s *SQLModelAsset) persistInputs(tx interface{}, inputs map[string]interface{}) error {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	for sourceModelName, inputValue := range inputs {
		switch df := inputValue.(type) {
		case *dataframe.DataFrame:
			// log.Debug().Str("sourceModelName", sourceModelName).Msgf("persisting %v", df)
			tempName := "tmp_" + strings.ReplaceAll(sourceModelName, ".", "_")
			err := dbConnection.PersistDataFrame(tx, tempName, df)
			if err != nil {
				defer dbConnection.Rallback(tx)
				log.Error().Stack().
					Err(err).
					Str("model", s.descriptor.Name).
					Str("connection", s.descriptor.ModelProfile.Connection).
					Msg("Failed to persist inputs")
				return err
			}
		default:
			log.Warn().Str("sourceModelName", sourceModelName).Msg("Input is not a *dataframe")
		}
	}
	return nil
}
