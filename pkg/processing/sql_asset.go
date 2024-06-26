package processing

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

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

	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	fmt.Printf("Executing: %s on %s \n", s.descriptor.Name, s.descriptor.ModelProfile.Connection)
	log.Debug().
		Str("s.descriptor.Name", s.descriptor.Name).
		Str("s.descriptor.ModelProfile.Connection", s.descriptor.ModelProfile.Connection).
		Msg("Executing asset")

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
		fmt.Printf("Schema %s does not exist \n", splitted[0])
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

	isTableExists := dbConnection.CheckTableExists(tx, s.descriptor.Name)

	switch s.descriptor.ModelProfile.Materialization {
	case configs.MAT_INCREMENTAL:
		if !isTableExists {
			return s.createTable(tx)
		} else {
			return s.updateTable(tx)
		}

	case configs.MAT_TABLE:
		if !isTableExists {
			return s.createTable(tx)
		} else {
			if err := s.truncateTable(tx); err == nil {
				return s.updateTable(tx)
			}
			return nil, err
		}

	case configs.MAT_VIEW:
		if !isTableExists {
			return s.createView(tx)
		}

	case configs.MAT_CUSTOM:
		panic("not implemented")
	case configs.MAT_UPSTREAM:
		panic("not implemented")
	}

	return nil, nil
}

func (s *SQLModelAsset) createView(tx interface{}) (interface{}, error) {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	var sqlQuery bytes.Buffer
	createViewSQLTemplate, err := template.New("createViewSQLTemplate").
		Funcs(FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions)).
		Parse(s.descriptor.CreateViewSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("parse view SQL %s", s.descriptor.CreateViewSQL)
		return nil, err
	}

	err = createViewSQLTemplate.Execute(&sqlQuery, nil)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("Render view SQL %s", sqlQuery.String())
		return nil, err
	}
	err = dbConnection.Exec(tx, sqlQuery.String())
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msg("Failed to create view")
		return nil, err
	}
	return nil, dbConnection.Commit(tx)
}

func (s *SQLModelAsset) createTable(tx interface{}) (interface{}, error) {

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
		return nil, err
	}

	var sqlQuery bytes.Buffer

	err = createTableSQLTempl.Execute(&sqlQuery, nil)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("Executing template: %s", sqlQuery.String())
		return nil, err
	}
	err = dbConnection.Exec(tx, sqlQuery.String())
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msg("Failed to create table")
		return nil, err
	}
	return nil, dbConnection.Commit(tx)
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

func (s *SQLModelAsset) updateTable(tx interface{}) (interface{}, error) {
	dbConnection := core.GetInstance().GetDBConnection(s.descriptor.ModelProfile.Connection)
	var sqlQuery bytes.Buffer
	s.functions["IsIncremental"] = func() bool {
		return s.descriptor.ModelProfile.Materialization == configs.MAT_INCREMENTAL
	}

	runSQLTemplate, err := template.New("runSQLTemplate").
		Funcs(FromConnectionContext(dbConnection, tx, s.descriptor.Name, s.functions)).
		Parse(s.descriptor.RunSQL)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("Parsing template: %s", s.descriptor.RunSQL)
		return nil, err
	}

	err = runSQLTemplate.Execute(&sqlQuery, nil)
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msgf("Rendering template: %s", sqlQuery.String())
		return nil, err
	}
	err = dbConnection.Exec(tx, sqlQuery.String())
	if err != nil {
		defer dbConnection.Rallback(tx)
		log.Error().Stack().Err(err).Msg("Failed to run incremental")
		return nil, err
	}
	return nil, dbConnection.Commit(tx)
}
