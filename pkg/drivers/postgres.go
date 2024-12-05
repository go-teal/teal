package drivers

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/go-teal/gota/dataframe"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/rs/zerolog/log"
)

type PostgresDBEngine struct {
	dbConnection *configs.DBConnectionConfig
	db           *sql.DB
}

// PersistDataFrame implements DBDriver.
func (d *PostgresDBEngine) PersistDataFrame(tx interface{}, name string, df *dataframe.DataFrame) error {
	panic("unimplemented")
}

// ToDataFrame implements DBDriver.
func (d *PostgresDBEngine) ToDataFrame(sql string) (*dataframe.DataFrame, error) {
	panic("unimplemented")
}

// MountSource implements DBDriver.
func (d *PostgresDBEngine) MountSource(sourceProfile *configs.SourceProfile) error {

	return nil
}

// UnMountSource implements DBDriver.
func (d *PostgresDBEngine) UnMountSource(sourceProfile *configs.SourceProfile) error {
	return nil
}

type PostgresDBEngineFactory struct {
}

// Rallback implements DBEngine.
func (d *PostgresDBEngine) Rallback(tx interface{}) error {
	return tx.(*sql.Tx).Rollback()
}

// Connect implements DBEngine.
func (d *PostgresDBEngine) Connect() error {
	var err error
	d.db, err = sql.Open("duckdb", d.dbConnection.Config.Path)
	log.Debug().Str("path", d.dbConnection.Config.Path).Msg("Connected")
	if err != nil {
		return err
	}
	return nil
}

// CreateConnection implements DBconnectionFactory.
func (d *PostgresDBEngineFactory) CreateConnection(connection configs.DBConnectionConfig) (DBDriver, error) {
	return initDuckDb(&connection)
}

func InitPostgresDBEnginFactory() DBconnectionFactory {
	return &PostgresDBEngineFactory{}
}

// CheckSchemaExists implements DBEngine.
func (d *PostgresDBEngine) CheckSchemaExists(tx interface{}, tableName string) bool {
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
func (d *PostgresDBEngine) Begin() (interface{}, error) {
	return d.db.Begin()
}

// CheckTableExists implements DBEngine.
func (d *PostgresDBEngine) CheckTableExists(tx interface{}, tableName string) bool {
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
func (d *PostgresDBEngine) Close() error {
	log.Debug().Str("path", d.dbConnection.Config.Path).Msg("disconnected")
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

// Commit implements DBEngine.
func (d *PostgresDBEngine) Commit(tx interface{}) error {
	return tx.(*sql.Tx).Commit()
}

// Exec implements DBEngine.
func (d *PostgresDBEngine) Exec(tx interface{}, sqlQuery string) error {
	log.Debug().Msg(sqlQuery)
	_, result := tx.(*sql.Tx).Exec(sqlQuery)
	if result != nil {
		log.Error().Msg(sqlQuery)
	}
	return result
}

// GetListOfFields implements DBEngine.
func (d *PostgresDBEngine) GetListOfFields(tx interface{}, tableName string) []string {
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

func (d *PostgresDBEngine) GetRawConnection() interface{} {
	return d.db
}

func initPostgresDb(dbConnectionConfig *configs.DBConnectionConfig) (DBDriver, error) {

	PostgresDBConnection := &PostgresDBEngine{
		dbConnection: dbConnectionConfig,
	}

	log.Debug().Msgf("Init DuckDB %s at %s\n", dbConnectionConfig.Name, dbConnectionConfig.Config.Path)
	_, err := os.Stat(dbConnectionConfig.Config.Path)

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
	return PostgresDBConnection, nil
}

func (d *PostgresDBEngine) ConcurrencyLock() {

}

func (d *PostgresDBEngine) ConcurrencyUnlock() {

}
