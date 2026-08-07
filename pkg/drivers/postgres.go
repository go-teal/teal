package drivers

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/rs/zerolog/log"
)

type PostgresDBEngine struct {
	dbConnection *configs.DBConnectionConfig
	db           *pgxpool.Pool
	schemaMutex  sync.Mutex
}

type PostgresDBEngineFactory struct {
}

// Rollback implements DBEngine.
func (d *PostgresDBEngine) Rollback(tx interface{}) error {
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

	if d.dbConnection.Config.PoolMaxConns > 0 {
		connectionParams = append(connectionParams, fmt.Sprintf("pool_max_conns=%d", d.dbConnection.Config.PoolMaxConns))
	}

	d.db, err = pgxpool.New(context.Background(), strings.Join(connectionParams, " "))
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

// CreateSchema implements DBEngine.
//
// CREATE SCHEMA IF NOT EXISTS is NOT race free in PostgreSQL: the catalog check
// is not atomic, so two sessions creating the same schema still collide on the
// pg_namespace unique index (see docs/issues/008). Since every asset gets its
// own pooled connection, the DDL is serialized explicitly - by a mutex for the
// goroutines of this process and by a transaction scoped advisory lock for
// other teal processes working on the same database.
func (d *PostgresDBEngine) CreateSchema(tx interface{}, schemaName string) error {
	d.schemaMutex.Lock()
	defer d.schemaMutex.Unlock()

	pgTx := tx.(pgx.Tx)
	ctx := context.Background()

	_, err := pgTx.Exec(ctx, "SELECT pg_advisory_xact_lock($1);", schemaAdvisoryLockID(schemaName))
	if err != nil {
		log.Error().Caller().Str("schema", schemaName).Err(err).Msg("Failed to lock the schema")
		return err
	}

	// The DDL runs inside a savepoint, so a lost race leaves the outer
	// transaction usable instead of aborting it.
	savepoint, err := pgTx.Begin(ctx)
	if err != nil {
		log.Error().Caller().Str("schema", schemaName).Err(err).Msg("Failed to open a savepoint")
		return err
	}

	_, err = savepoint.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", schemaName))
	if err != nil {
		defer savepoint.Rollback(ctx)
		if isDuplicateSchemaError(err) {
			log.Debug().Str("schema", schemaName).Msg("Schema has been created by a concurrent session")
			return nil
		}
		log.Error().Caller().Str("schema", schemaName).Err(err).Msg("Failed to create schema")
		return err
	}

	return savepoint.Commit(ctx)
}

// schemaAdvisoryLockID maps a schema name to a stable advisory lock id shared by
// every teal process.
func schemaAdvisoryLockID(schemaName string) int64 {
	hash := fnv.New64a()
	hash.Write([]byte("teal.schema." + schemaName))
	return int64(hash.Sum64())
}

func isDuplicateSchemaError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	// 42P06 - duplicate_schema, 23505 - unique_violation on pg_namespace
	return pgErr.Code == "42P06" || pgErr.Code == "23505"
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
	d.db.Close()
	return nil
}

// Commit implements DBEngine.
func (d *PostgresDBEngine) Commit(tx interface{}) error {
	return tx.(pgx.Tx).Commit(context.Background())
}

// Exec implements DBEngine.
func (d *PostgresDBEngine) Exec(tx interface{}, sqlQuery string) error {
	log.Debug().Str("sql", sqlQuery).Msg("Executing SQL query")
	_, result := tx.(pgx.Tx).Exec(context.Background(), sqlQuery)
	if result != nil {
		log.Error().Caller().Str("sql", sqlQuery).Err(result).Msg("SQL execution failed")
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

// pgxpool.Pool is safe for concurrent use; each Begin/Query checks out its own
// connection from the pool. No driver-level serialization needed.
func (d *PostgresDBEngine) ConcurrencyLock()   {}
func (d *PostgresDBEngine) ConcurrencyUnlock() {}
