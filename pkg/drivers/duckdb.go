package drivers

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/rs/zerolog/log"
)

type DuckDBEngine struct {
	dbConnection *configs.DBConnectionConfig
	db           *sql.DB
	Mutex        *sync.Mutex
	// schemaMutex is deliberately separate from Mutex: Mutex is already held by
	// ConcurrencyLock() for the whole asset execution and Go mutexes are not
	// reentrant.
	schemaMutex sync.Mutex
}

type DuckDBEngineFactory struct {
}

// Rollback implements DBEngine.
func (d *DuckDBEngine) Rollback(tx interface{}) error {
	return tx.(*sql.Tx).Rollback()
}

// Connect implements DBEngine.
func (d *DuckDBEngine) Connect() error {
	var err error
	d.db, err = sql.Open("duckdb", d.dbConnection.Config.Path)
	log.Debug().Str("path", d.dbConnection.Config.Path).Msg("Connected")
	if err != nil {
		return err
	}
	for _, extentionName := range d.dbConnection.Config.Extensions {
		_, err = d.db.Exec(fmt.Sprintf("LOAD %s;", extentionName))
		if err != nil {
			return err
		}
		log.Debug().Msgf("load extension: %s\n", extentionName)
	}
	return nil
}

// CreateConnection implements DBconnectionFactory.
func (d *DuckDBEngineFactory) CreateConnection(connection configs.DBConnectionConfig) (DBDriver, error) {
	return initDuckDb(&connection)
}

func InitDuckDBEnginFactory() DBconnectionFactory {
	return &DuckDBEngineFactory{}
}

// CheckSchemaExists implements DBEngine.
func (d *DuckDBEngine) CheckSchemaExists(tx interface{}, tableName string) bool {
	splitted := strings.Split(tableName, ".")
	query := "SELECT count(DISTINCT schema_name) from information_schema.schemata WHERE schema_name=$1;"
	var count int
	err := tx.(*sql.Tx).QueryRow(query, splitted[0]).Scan(&count)
	if err != nil {
		panic(err)
	}
	return count > 0
}

// Begin implements DBEngine.
func (d *DuckDBEngine) Begin() (interface{}, error) {
	return d.db.Begin()
}

// CreateSchema implements DBEngine. The DDL is serialized by its own mutex, so
// two assets of the same stage can not create the schema at the same time.
func (d *DuckDBEngine) CreateSchema(tx interface{}, schemaName string) error {
	d.schemaMutex.Lock()
	defer d.schemaMutex.Unlock()

	_, err := tx.(*sql.Tx).Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", schemaName))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Debug().Str("schema", schemaName).Msg("Schema has been created by a concurrent session")
			return nil
		}
		log.Error().Caller().Str("schema", schemaName).Err(err).Msg("Failed to create schema")
		return err
	}
	return nil
}

// CheckTableExists implements DBEngine.
func (d *DuckDBEngine) CheckTableExists(tx interface{}, tableName string) bool {
	splitted := strings.Split(tableName, ".")
	query := "SELECT count(DISTINCT table_name) from information_schema.tables WHERE table_schema=$1 and table_name=$2;"
	var count int
	err := tx.(*sql.Tx).QueryRow(query, splitted[0], splitted[1]).Scan(&count)
	if err != nil {
		panic(err)
	}
	return count > 0
}

// Close implements DBEngine.
func (d *DuckDBEngine) Close() error {
	log.Debug().Str("path", d.dbConnection.Config.Path).Msg("disconnected")
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

// Commit implements DBEngine.
func (d *DuckDBEngine) Commit(tx interface{}) error {
	return tx.(*sql.Tx).Commit()
}

// Exec implements DBEngine.
func (d *DuckDBEngine) Exec(tx interface{}, sqlQuery string) error {
	log.Debug().Str("sql", sqlQuery).Msg("Executing SQL query")
	_, result := tx.(*sql.Tx).Exec(sqlQuery)
	if result != nil {
		log.Error().Caller().Str("sql", sqlQuery).Err(result).Msg("SQL execution failed")
	}
	return result
}

// GetListOfFields implements DBEngine.
func (d *DuckDBEngine) GetListOfFields(tx interface{}, tableName string) []string {
	var fields []string
	splitted := strings.Split(tableName, ".")
	rows, err := tx.(*sql.Tx).Query("SELECT column_name FROM information_schema.columns WHERE table_schema = $1 AND table_name = $2;", splitted[0], splitted[1])
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var field string
		err := rows.Scan(&field)
		if err != nil {
			panic(err)
		}
		fields = append(fields, field)
	}
	return fields
}

func (d *DuckDBEngine) GetRawConnection() interface{} {
	return d.db
}

func initDuckDb(dbConnectionConfig *configs.DBConnectionConfig) (DBDriver, error) {

	duckDBConnection := &DuckDBEngine{
		dbConnection: dbConnectionConfig,
		Mutex:        &sync.Mutex{},
	}

	log.Debug().Msgf("Init DuckDB %s at %s\n", dbConnectionConfig.Name, dbConnectionConfig.Config.Path)
	_, err := os.Stat(dbConnectionConfig.Config.Path)
	log.Warn().Err(err).Send()

	if os.IsNotExist(err) {
		db, err := sql.Open("duckdb", dbConnectionConfig.Config.Path)
		if err != nil {
			panic(err)
		}
		defer db.Close()
		if len(dbConnectionConfig.Config.Extensions) > 0 {
			log.Info().Msgf("Installing extensions: %v\n", dbConnectionConfig.Config.Extensions)
		}
		for _, extentionName := range dbConnectionConfig.Config.Extensions {
			_, err := db.Exec(fmt.Sprintf("INSTALL %s;", extentionName))
			if err != nil {
				panic(err)
			}
			_, err = db.Exec(fmt.Sprintf("LOAD %s;", extentionName))
			if err != nil {
				panic(err)
			}
			log.Info().Msgf("Installed extension: %s\n", extentionName)
		}
	}
	return duckDBConnection, nil
}

func (d *DuckDBEngine) ConcurrencyLock() {
	d.Mutex.Lock()
}

func (d *DuckDBEngine) ConcurrencyUnlock() {
	d.Mutex.Unlock()
}
