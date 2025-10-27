package processing

import (
	"fmt"
	"strings"
	"time"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/gota/dataframe"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/models"

	"github.com/rs/zerolog/log"
)

type SQLModelAsset struct {
	descriptor *models.SQLModelDescriptor
	functions  pongo2.Context
}

func InitSQLModelAsset(descriptor *models.SQLModelDescriptor) Asset {
	return &SQLModelAsset{
		descriptor: descriptor,
		functions:  make(pongo2.Context),
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
func (s *SQLModelAsset) Execute(ctx *TaskContext) (interface{}, error) {

	var data *dataframe.DataFrame
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	dbConnection.ConcurrencyLock()
	defer dbConnection.ConcurrencyUnlock()

	log.Debug().
		Str("taskId", ctx.TaskID).
		Str("taskUUID", ctx.TaskUUID).
		Str("assetName", s.descriptor.Name).
		Str("connection", s.descriptor.ModelProfile.Connection).
		Msg("Executing asset")
	log.Debug().
		Str("taskId", ctx.TaskID).
		Str("taskUUID", ctx.TaskUUID).
		Str("assetName", s.descriptor.Name).
		Msgf("input params: %v", ctx.Input)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Str("taskId", ctx.TaskID).
			Str("taskUUID", ctx.TaskUUID).
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return nil, err
	}

	isSchemaExists := dbConnection.CheckSchemaExists(tx, s.descriptor.Name)

	if !isSchemaExists {
		splitted := strings.Split(s.descriptor.Name, ".")
		log.Info().
			Str("taskId", ctx.TaskID).
			Str("taskUUID", ctx.TaskUUID).
			Str("assetName", s.descriptor.Name).
			Msgf("Schema %s does not exist", splitted[0])
		// TODO: Move this to the driver
		err = dbConnection.Exec(tx, fmt.Sprintf("CREATE SCHEMA %s;", splitted[0]))
		if err != nil {
			defer dbConnection.Rallback(tx)
			log.Error().Caller().
				Str("taskId", ctx.TaskID).
				Str("taskUUID", ctx.TaskUUID).
				Str("assetName", s.descriptor.Name).
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
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to create schema or table")
		defer dbConnection.Rallback(tx)
		return nil, err
	}

	switch s.descriptor.ModelProfile.Materialization {
	case configs.MAT_INCREMENTAL:
		if s.descriptor.ModelProfile.PersistInputs {
			err := s.persistInputs(ctx.Input)
			if err != nil {
				log.Error().Caller().
					Str("taskId", ctx.TaskID).
					Str("taskUUID", ctx.TaskUUID).
					Str("assetName", s.descriptor.Name).
					Err(err).
					Msg("Failed to persist inputs")
				return nil, err
			}
		}
		if s.descriptor.ModelProfile.IsDataFramed {
			data, err = s.getDataFrame(ctx, isTableExists)
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
				err := s.persistInputs(ctx.Input)
				if err != nil {
					log.Error().Caller().
						Err(err).
						Msg("Failed to persist inputs")
					return nil, err
				}
			}
			if s.descriptor.ModelProfile.IsDataFramed {
				log.Warn().
				Str("taskId", ctx.TaskID).
				Str("taskUUID", ctx.TaskUUID).
				Str("assetName", s.descriptor.Name).
				Msg("Dataframe can slow this operation, considner custom or incremental materialization")
				data, err = s.getDataFrame(ctx, false)
				if err != nil {
					return nil, err
				}
			}
			err = s.createTable(ctx)
		} else {
			if err = s.truncateTable(); err == nil {
				log.Debug().
					Str("taskId", ctx.TaskID).
					Str("taskUUID", ctx.TaskUUID).
					Str("assetName", s.descriptor.Name).
					Msg("table has been truncated")
				if s.descriptor.ModelProfile.PersistInputs {
					err := s.persistInputs(ctx.Input)
					if err != nil {
						log.Error().Caller().
							Err(err).
							Msg("Failed to persist inputs")
						return nil, err
					}
				}
				if s.descriptor.ModelProfile.IsDataFramed {
					log.Warn().
				Str("taskId", ctx.TaskID).
				Str("taskUUID", ctx.TaskUUID).
				Str("assetName", s.descriptor.Name).
				Msg("Dataframe can slow this operation, considner custom or incremental materialization")
					data, err = s.getDataFrame(ctx, false)
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
						Str("assetName", s.descriptor.Name).
						Err(err).
						Msg("Failed to insert to table")
					return nil, err
				}
			} else {
				log.Error().Caller().
					Str("taskId", ctx.TaskID).
					Str("taskUUID", ctx.TaskUUID).
					Str("assetName", s.descriptor.Name).
					Err(err).
					Msg("Failed to truncate table")
				return nil, err
			}
		}
	case configs.MAT_VIEW:

		if s.descriptor.ModelProfile.PersistInputs {
			err := s.persistInputs(ctx.Input)
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
			log.Warn().
				Str("taskId", ctx.TaskID).
				Str("taskUUID", ctx.TaskUUID).
				Str("assetName", s.descriptor.Name).
				Msg("Dataframe can slow this operation, considner custom or incremental materialization")
			data, err = s.getDataFrame(ctx, false)
			if err != nil {
				return nil, err
			}
		}
		if !isTableExists {
			err = s.createView(ctx)
		}
	case configs.MAT_CUSTOM:

		if s.descriptor.ModelProfile.PersistInputs {
			err := s.persistInputs(ctx.Input)
			if err != nil {
				log.Error().Caller().
					Str("taskId", ctx.TaskID).
					Str("taskUUID", ctx.TaskUUID).
					Str("assetName", s.descriptor.Name).
					Err(err).
					Msg("Failed to persist inputs")
				return nil, err
			}
		}

		if s.descriptor.ModelProfile.IsDataFramed {
			data, err = s.getDataFrame(ctx, false)
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

	log.Info().
		Str("taskId", ctx.TaskID).
		Str("taskUUID", ctx.TaskUUID).
		Str("assetName", s.descriptor.Name).
		Msgf("Testing %s", s.descriptor.Name)
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
				log.Info().
					Str("taskId", ctx.TaskID).
					Str("taskUUID", ctx.TaskUUID).
					Str("assetName", s.descriptor.Name).
					Str("testName", testName).
					Int64("durationMs", result.DurationMs).
					Msg("Success")
			} else {
				result.Status = TestStatusFailed
				result.Error = err
				result.Message = testName
				log.Error().
					Str("taskId", ctx.TaskID).
					Str("taskUUID", ctx.TaskUUID).
					Str("assetName", s.descriptor.Name).
					Str("testName", testName).
					Err(err).
					Int64("durationMs", result.DurationMs).
					Msg("Failed")
			}
		} else {
			result.Status = TestStatusNotFound
			result.DurationMs = time.Since(startTime).Milliseconds()
			result.Message = fmt.Sprintf("Test %s not found", testConfig.Name)
			log.Warn().
				Str("taskId", ctx.TaskID).
				Str("taskUUID", ctx.TaskUUID).
				Str("assetName", s.descriptor.Name).
				Msg(result.Message)
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
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	createViewSQLTemplate, err := pongo2.FromString(s.descriptor.CreateViewSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Str("sql", s.descriptor.CreateViewSQL).
			Err(err).
			Msg("Failed to parse view SQL")
		return err
	}

	context := MergePongo2Context(
		FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions),
		FromTaskContextPongo2(ctx),
	)
	sqlQuery, err := createViewSQLTemplate.Execute(context)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Str("sql", sqlQuery).
			Err(err).
			Msg("Failed to render view SQL")
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to create view")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) createTable(ctx *TaskContext) error {

	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	s.functions["IsIncremental"] = func() bool {
		return false
	}

	createTableSQLTempl, err := pongo2.FromString(s.descriptor.CreateTableSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Str("sql", s.descriptor.CreateTableSQL).
			Err(err).
			Msg("Failed to parse table SQL template")
		return err
	}

	context := MergePongo2Context(
		FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions),
		FromTaskContextPongo2(ctx),
	)
	sqlQuery, err := createTableSQLTempl.Execute(context)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Str("sql", sqlQuery).
			Err(err).
			Msg("Failed to execute table SQL template")
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to create table")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) truncateTable() error {

	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	err = dbConnection.Exec(tx, s.descriptor.TruncateTableSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to run truncate")
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
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	simleSQLQueryTemplate, err := pongo2.FromString(s.descriptor.RawSQL)
	if err != nil {
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to parse asset query")
		return err
	}
	context := MergePongo2Context(
		FromConnectionContext(dbConnection, nil, s.descriptor.Name, s.functions),
		FromTaskContextPongo2(ctx),
	)
	sqlQuery, err := simleSQLQueryTemplate.Execute(context)
	if err != nil {
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msgf("Rendering template: %s", sqlQuery)
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to run a custom query")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) insertToTable(ctx *TaskContext) error {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to begin transaction")
		defer dbConnection.Rallback(tx)
		return err
	}

	s.functions["IsIncremental"] = func() bool {
		return s.descriptor.ModelProfile.Materialization == configs.MAT_INCREMENTAL
	}

	runSQLTemplate, err := pongo2.FromString(s.descriptor.InsertSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Str("sql", s.descriptor.InsertSQL).
			Err(err).
			Msg("Failed to parse insert SQL template")
		return err
	}

	context := MergePongo2Context(
		FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions),
		FromTaskContextPongo2(ctx),
	)
	sqlQuery, err := runSQLTemplate.Execute(context)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Str("sql", sqlQuery).
			Err(err).
			Msg("Failed to render insert SQL template")
		return err
	}
	err = dbConnection.Exec(tx, sqlQuery)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to insert")
		return err
	}
	return dbConnection.Commit(tx)
}

func (s *SQLModelAsset) getDataFrame(ctx *TaskContext, isIncremental bool) (*dataframe.DataFrame, error) {

	s.functions["IsIncremental"] = func() bool {
		return isIncremental
	}

	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	simleSQLQueryTemplate, err := pongo2.FromString(s.descriptor.RawSQL)
	if err != nil {
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msg("Failed to parse asset query")
		return nil, err
	}
	context := MergePongo2Context(
		FromConnectionContext(dbConnection, nil, s.descriptor.Name, s.functions),
		FromTaskContextPongo2(ctx),
	)
	sqlQuery, err := simleSQLQueryTemplate.Execute(context)
	if err != nil {
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msgf("Rendering template: %s", sqlQuery)
		return nil, err
	}

	data, err := dbConnection.ToDataFrame(sqlQuery)
	if err != nil {
		log.Error().Caller().Stack().
			Str("assetName", s.descriptor.Name).
			Err(err).
			Msgf("Failed to create a DataFrame for: %s", sqlQuery)
		return nil, err
	}
	return data, nil
}

func (s *SQLModelAsset) persistInputs(inputs map[string]interface{}) error {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)

	tx, err := dbConnection.Begin()
	if err != nil {
		log.Error().Caller().
			Str("assetName", s.descriptor.Name).
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
					Str("assetName", s.descriptor.Name).
					Str("connection", s.descriptor.ModelProfile.Connection).
					Err(err).
					Msg("Failed to persist inputs")
				defer dbConnection.Rallback(tx)
				return err
			}
		default:
			log.Warn().
				Str("assetName", s.descriptor.Name).
				Str("sourceModelName", sourceModelName).
				Msg("Input is not a *dataframe")
		}
	}
	return dbConnection.Commit(tx)
}
