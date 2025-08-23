package drivers

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/rs/zerolog/log"
)

type PostgresDBEngine struct {
	dbConnection *configs.DBConnectionConfig
	db           *pgx.Conn
}

// MountSource implements PGDriver.
func (d *PostgresDBEngine) MountSource(sourceProfile *configs.SourceProfile) error {

	return nil
}

// UnMountSource implements PGDriver.
func (d *PostgresDBEngine) UnMountSource(sourceProfile *configs.SourceProfile) error {
	return nil
}

type PostgresDBEngineFactory struct {
}

// Rallback implements DBEngine.
func (d *PostgresDBEngine) Rallback(tx interface{}) error {
	return tx.(pgx.Tx).Rollback(context.Background())
}

// Connect implements DBEngine.
func (d *PostgresDBEngine) Connect() error {
	var err error
	connectionParams := make([]string, 0)

	connectionParams = append(connectionParams, fmt.Sprintf("host=%s", d.dbConnection.Config.Host))
	connectionParams = append(connectionParams, fmt.Sprintf("port=%d", d.dbConnection.Config.Port))
	connectionParams = append(connectionParams, fmt.Sprintf("user=%s", d.dbConnection.Config.User))
	connectionParams = append(connectionParams, fmt.Sprintf("database=%s", d.dbConnection.Config.Database))
	connectionParams = append(connectionParams, fmt.Sprintf("password=%s", d.dbConnection.Config.Password))

	if d.dbConnection.Config.DBSSLMode != "" {
		connectionParams = append(connectionParams, fmt.Sprintf("sslmode=%s", d.dbConnection.Config.DBSSLMode))
	}

	if d.dbConnection.Config.DBRootCert != "" {
		connectionParams = append(connectionParams, fmt.Sprintf("sslrootcert=%s", d.dbConnection.Config.DBRootCert))
	}

	if d.dbConnection.Config.DBCert != "" {
		connectionParams = append(connectionParams, fmt.Sprintf("sslcert=%s", d.dbConnection.Config.DBCert))
	}

	if d.dbConnection.Config.DBKey != "" {
		connectionParams = append(connectionParams, fmt.Sprintf("sslkey=%s", d.dbConnection.Config.DBKey))
	}

	d.db, err = pgx.Connect(context.Background(), strings.Join(connectionParams, " "))
	log.Debug().Msg("Connected")
	if err != nil {
		return err
	}
	return nil
}

// CreateConnection implements DBconnectionFactory.
func (d *PostgresDBEngineFactory) CreateConnection(connection configs.DBConnectionConfig) (DBDriver, error) {
	return initPostgresDb(&connection)
}

func InitPostgresDBEnginFactory() DBconnectionFactory {
	return &PostgresDBEngineFactory{}
}

// CheckSchemaExists implements DBEngine.
func (d *PostgresDBEngine) CheckSchemaExists(tx interface{}, tableName string) bool {
	splitted := strings.Split(tableName, ".")
	query := "SELECT count(DISTINCT schema_name) from information_schema.schemata WHERE schema_name=$1;"
	var count int
	err := tx.(pgx.Tx).QueryRow(context.Background(), query, splitted[0]).Scan(&count)
	if err != nil {
		panic(err)
	}
	return count > 0
}

// Begin implements DBEngine.
func (d *PostgresDBEngine) Begin() (interface{}, error) {
	return d.db.Begin(context.Background())
}

// CheckTableExists implements DBEngine.
func (d *PostgresDBEngine) CheckTableExists(tx interface{}, tableName string) bool {
	splitted := strings.Split(tableName, ".")
	query := "SELECT count(DISTINCT table_name) from information_schema.tables WHERE table_schema=$1 and table_name=$2;"
	var count int
	err := tx.(pgx.Tx).QueryRow(context.Background(), query, splitted[0], splitted[1]).Scan(&count)
	if err != nil {
		panic(err)
	}
	return count > 0
}

// Close implements DBEngine.
func (d *PostgresDBEngine) Close() error {
	log.Debug().Str("host", d.dbConnection.Config.Host).Int("port", d.dbConnection.Config.Port).Msg("disconnected")
	if d.db == nil {
		return nil
	}
	return d.db.Close(context.Background())
}

// Commit implements DBEngine.
func (d *PostgresDBEngine) Commit(tx interface{}) error {
	return tx.(pgx.Tx).Commit(context.Background())
}

// Exec implements DBEngine.
func (d *PostgresDBEngine) Exec(tx interface{}, sqlQuery string) error {
	log.Debug().Msg(sqlQuery)
	_, result := tx.(pgx.Tx).Exec(context.Background(), sqlQuery)
	if result != nil {
		log.Error().Caller().Msg(sqlQuery)
	}
	return result
}

// GetListOfFields implements DBEngine.
func (d *PostgresDBEngine) GetListOfFields(tx interface{}, tableName string) []string {
	var fields []string
	splitted := strings.Split(tableName, ".")
	rows, err := tx.(pgx.Tx).Query(context.Background(), "SELECT column_name FROM information_schema.columns WHERE table_schema = $1 AND table_name = $2;", splitted[0], splitted[1])
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

	log.Debug().Msgf("Init PostgreSQL %s at %s\n", dbConnectionConfig.Name, dbConnectionConfig.Config.Host)

	return PostgresDBConnection, nil
}

func (d *PostgresDBEngine) ConcurrencyLock() {

}

func (d *PostgresDBEngine) ConcurrencyUnlock() {

}
