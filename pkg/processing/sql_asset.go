package processing

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

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

// GetDescriptor implements Asset.
func (s *SQLModelAsset) GetDescriptor() any {
	return s.descriptor
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
func (s *SQLModelAsset) Execute(ctx *TaskContext, input map[string]interface{}) (interface{}, error) {

	var data *dataframe.DataFrame
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	dbConnection.ConcurrencyLock()
	defer dbConnection.ConcurrencyUnlock()

	log.Debug().
		Str("taskId", ctx.TaskID).
		Str("s.descriptor.Name", s.descriptor.Name).
		Str("s.descriptor.ModelProfile.Connection", s.descriptor.ModelProfile.Connection).
		Msg("Executing asset")
	log.Debug().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Str("s.descriptor.Name", s.descriptor.Name).Msgf("input params: %v", input)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return nil, err
	}

	isSchemaExists := dbConnection.CheckSchemaExists(tx, s.descriptor.Name)

	if !isSchemaExists {
		splitted := strings.Split(s.descriptor.Name, ".")
		log.Info().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Msgf("Schema %s does not exist", splitted[0])
		// TODO: Move this to the driver
		err = dbConnection.Exec(tx, fmt.Sprintf("CREATE SCHEMA %s;", splitted[0]))
		if err != nil {
			defer dbConnection.Rallback(tx)
			log.Error().Caller().
				Str("taskId", ctx.TaskID).
				Str("taskUUID", ctx.TaskUUID).
				Err(err).
				Msg("Failed to create schema")

			return nil, err
		}
	}

	isTableExists := dbConnection.CheckTableExists(tx, s.descriptor.Name)

	err = dbConnection.Commit(tx)

	if err != nil {
		log.Error().Caller().
			Str("taskId", ctx.TaskID).
			Str("taskUUID", ctx.TaskUUID).
			Err(err).
			Msg("Failed to create schema or table")
		defer dbConnection.Rallback(tx)
		return nil, err
	}

	switch s.descriptor.ModelProfile.Materialization {
	case configs.MAT_INCREMENTAL:
		if s.descriptor.ModelProfile.PersistInputs {
			err := s.persistInputs(input)
			if err != nil {
				log.Error().Caller().
					Str("taskId", ctx.TaskID).
					Str("taskUUID", ctx.TaskUUID).
					Err(err).
					Msg("Failed to persist inputs")
				return nil, err
			}
		}
		if s.descriptor.ModelProfile.IsDataFramed {
			data, err = s.getDataFrame(ctx)
			if err != nil {
				return nil, err
			}
		}

		if !isTableExists {
			err = s.createTable(ctx)
		} else {
			err = s.insertToTable(ctx)
		}

	case configs.MAT_TABLE:

		if !isTableExists {
			if s.descriptor.ModelProfile.PersistInputs {
				err := s.persistInputs(input)
				if err != nil {
					log.Error().Caller().
						Err(err).
						Msg("Failed to persist inputs")
					return nil, err
				}
			}
			if s.descriptor.ModelProfile.IsDataFramed {
				log.Warn().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Msg("Dataframe can slow this operation, considner custom or incremental materialization")
				data, err = s.getDataFrame(ctx)
				if err != nil {
					return nil, err
				}
			}
			err = s.createTable(ctx)
		} else {
			if err = s.truncateTable(); err == nil {
				log.Debug().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Msg("table has been truncated")
				if s.descriptor.ModelProfile.PersistInputs {
					err := s.persistInputs(input)
					if err != nil {
						log.Error().Caller().
							Err(err).
							Msg("Failed to persist inputs")
						return nil, err
					}
				}
				if s.descriptor.ModelProfile.IsDataFramed {
					log.Warn().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Msg("Dataframe can slow this operation, considner custom or incremental materialization")
					data, err = s.getDataFrame(ctx)
					if err != nil {
						return nil, err
					}
				}
				err = s.insertToTable(ctx)

				if err != nil {
					defer dbConnection.Rallback(tx)
					log.Error().Caller().
						Str("taskId", ctx.TaskID).
						Str("taskUUID", ctx.TaskUUID).
						Err(err).
						Msg("Failed to insert to table")
					return nil, err
				}
			} else {
				log.Error().Caller().
					Str("taskId", ctx.TaskID).
					Str("taskUUID", ctx.TaskUUID).
					Err(err).
					Msg("Failed to truncate table")
				return nil, err
			}
		}
	case configs.MAT_VIEW:

		if s.descriptor.ModelProfile.PersistInputs {
			err := s.persistInputs(input)
			if err != nil {
				defer dbConnection.Rallback(tx)
				log.Error().Caller().
					Str("taskId", ctx.TaskID).
					Str("taskUUID", ctx.TaskUUID).
					Err(err).
					Msg("Failed to persist inputs")
				return nil, err
			}
		}

		if s.descriptor.ModelProfile.IsDataFramed {
			log.Warn().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Msg("Dataframe can slow this operation, considner custom or incremental materialization")
			data, err = s.getDataFrame(ctx)
			if err != nil {
				return nil, err
			}
		}
		if !isTableExists {
			err = s.createView(ctx)
		}
	case configs.MAT_CUSTOM:

		if s.descriptor.ModelProfile.PersistInputs {
			err := s.persistInputs(input)
			if err != nil {
				log.Error().Caller().
					Str("taskId", ctx.TaskID).
					Str("taskUUID", ctx.TaskUUID).
					Err(err).
					Msg("Failed to persist inputs")
				return nil, err
			}
		}

		if s.descriptor.ModelProfile.IsDataFramed {
			data, err = s.getDataFrame(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			err = s.customQuery(ctx)
		}
	case configs.MAT_RAW:
		panic("SQL Model can not be raw")
	}

	return data, err
}

// RunTests implements Asset.
func (s *SQLModelAsset) RunTests(ctx *TaskContext, testsMap map[string]ModelTesting) []TestResult {
	results := make([]TestResult, 0)
	
	if len(s.descriptor.ModelProfile.Tests) == 0 {
		return results
	}
	
	log.Info().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Msgf("Testing %s", s.descriptor.Name)
	for _, testConfig := range s.descriptor.ModelProfile.Tests {
		startTime := time.Now()
		result := TestResult{
			TestName: testConfig.Name,
		}
		
		if testCase, ok := testsMap[testConfig.Name]; ok {
			status, testName, err := testCase.Execute(ctx)
			result.DurationMs = time.Since(startTime).Milliseconds()
			
			if status {
				result.Status = TestStatusSuccess
				result.Message = testName
				log.Info().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Str("testCase", testName).Int64("durationMs", result.DurationMs).Msg("Success")
			} else {
				result.Status = TestStatusFailed
				result.Error = err
				result.Message = testName
				log.Error().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Str("testCase", testName).Err(err).Int64("durationMs", result.DurationMs).Msg("Failed")
			}
		} else {
			result.Status = TestStatusNotFound
			result.DurationMs = time.Since(startTime).Milliseconds()
			result.Message = fmt.Sprintf("Test %s not found", testConfig.Name)
			log.Warn().Str("taskId", ctx.TaskID).Str("taskUUID", ctx.TaskUUID).Msg(result.Message)
		}
		
		results = append(results, result)
	}
	
	return results
}

func (s *SQLModelAsset) createView(ctx *TaskContext) error {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	var sqlQuery bytes.Buffer
	createViewSQLTemplate, err := template.New("createViewSQLTemplate").
		Funcs(MergeTemplateFuncs(
			FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions),
			FromTaskContext(ctx),
		)).
		Parse(s.descriptor.CreateViewSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Str("sql", s.descriptor.CreateViewSQL).Msg("Failed to parse view SQL")
		return err
	}

	err = createViewSQLTemplate.Execute(&sqlQuery, nil)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Str("sql", sqlQuery.String()).Msg("Failed to render view SQL")
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery.String())
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Msg("Failed to create view")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) createTable(ctx *TaskContext) error {

	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	s.functions["IsIncremental"] = func() bool {
		return false
	}

	createTableSQLTempl, err := template.New("createTableSQLTempl").
		Funcs(MergeTemplateFuncs(
			FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions),
			FromTaskContext(ctx),
		)).
		Parse(s.descriptor.CreateTableSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Str("sql", s.descriptor.CreateTableSQL).Msg("Failed to parse table SQL template")
		return err
	}

	var sqlQuery bytes.Buffer

	err = createTableSQLTempl.Execute(&sqlQuery, nil)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Str("sql", sqlQuery.String()).Msg("Failed to execute table SQL template")
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery.String())
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Msg("Failed to create table")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) truncateTable() error {

	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	err = dbConnection.Exec(tx, s.descriptor.TruncateTableSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Msg("Failed to run truncate")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) customQuery(ctx *TaskContext) error {

	s.functions["IsIncremental"] = func() bool {
		return s.descriptor.ModelProfile.Materialization == configs.MAT_INCREMENTAL
	}

	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	simleSQLQueryTemplate, err := template.New("simleSQLQueryTemplate").
		Funcs(MergeTemplateFuncs(
			FromConnectionContext(dbConnection, nil, s.descriptor.Name, s.functions),
			FromTaskContext(ctx),
		)).
		Parse(s.descriptor.RawSQL)
	if err != nil {
		log.Error().Caller().Stack().
			Err(err).
			Msg("Failed to parse asset query")
		return err
	}
	var sqlQuery bytes.Buffer
	err = simleSQLQueryTemplate.Execute(&sqlQuery, nil)
	if err != nil {
		log.Error().Caller().Stack().
			Err(err).
			Msgf("Rendering template: %s", sqlQuery.String())
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery.String())
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Msg("Failed to run a custom query")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) insertToTable(ctx *TaskContext) error {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	var sqlQuery bytes.Buffer
	s.functions["IsIncremental"] = func() bool {
		return s.descriptor.ModelProfile.Materialization == configs.MAT_INCREMENTAL
	}

	runSQLTemplate, err := template.New("Insert SQL template").
		Funcs(MergeTemplateFuncs(
			FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions),
			FromTaskContext(ctx),
		)).
		Parse(s.descriptor.InsertSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Str("sql", s.descriptor.InsertSQL).Msg("Failed to parse insert SQL template")
		return err
	}

	err = runSQLTemplate.Execute(&sqlQuery, nil)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Str("sql", sqlQuery.String()).Msg("Failed to render insert SQL template")
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery.String())
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().Err(err).Msg("Failed to insert")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) getDataFrame(ctx *TaskContext) (*dataframe.DataFrame, error) {

	s.functions["IsIncremental"] = func() bool {
		return s.descriptor.ModelProfile.Materialization == configs.MAT_INCREMENTAL
	}

	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	simleSQLQueryTemplate, err := template.New("simleSQLQueryTemplate").
		Funcs(MergeTemplateFuncs(
			FromConnectionContext(dbConnection, nil, s.descriptor.Name, s.functions),
			FromTaskContext(ctx),
		)).
		Parse(s.descriptor.RawSQL)
	if err != nil {
		log.Error().Caller().Stack().
			Err(err).
			Msg("Failed to parse asset query")
		return nil, err
	}
	var sqlQuery bytes.Buffer
	err = simleSQLQueryTemplate.Execute(&sqlQuery, nil)
	if err != nil {
		log.Error().Caller().Stack().
			Err(err).
			Msgf("Rendering template: %s", sqlQuery.String())
		return nil, err
	}

	data, err := dbConnection.ToDataFrame(sqlQuery.String())
	if err != nil {
		log.Error().Caller().Stack().
			Err(err).
			Msgf("Failed to create a DataFrame for: %s", sqlQuery.String())
		return nil, err
	}
	return data, nil
}

func (s *SQLModelAsset) persistInputs(inputs map[string]interface{}) error {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	for sourceModelName, inputValue := range inputs {
		switch df := inputValue.(type) {
		case *dataframe.DataFrame:
			if df == nil {
				continue
			}
			// log.Debug().Str("sourceModelName", sourceModelName).Msgf("persisting %v", df)
			tempName := "tmp_" + strings.ReplaceAll(sourceModelName, ".", "_")
			err := dbConnection.PersistDataFrame(tx, tempName, df)
			if err != nil {

				log.Error().Caller().Stack().
					Err(err).
					Str("model", s.descriptor.Name).
					Str("connection", s.descriptor.ModelProfile.Connection).
					Msg("Failed to persist inputs")
				defer dbConnection.Rallback(tx)
				return err
			}
		default:
			log.Warn().Str("sourceModelName", sourceModelName).Msg("Input is not a *dataframe")
		}
	}
	return dbConnection.Commit(tx)
}
