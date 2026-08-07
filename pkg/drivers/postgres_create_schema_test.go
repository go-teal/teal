package drivers

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/go-teal/teal/pkg/configs"
)

// Needs a live PostgreSQL. Point it at a throwaway database:
//
//	TEAL_TEST_PG_HOST=127.0.0.1 TEAL_TEST_PG_PORT=5432 TEAL_TEST_PG_USER=teal \
//	TEAL_TEST_PG_DATABASE=teal_test go test ./pkg/drivers/ -run CreateSchema
func newTestPostgresEngine(t *testing.T, poolMaxConns int) *PostgresDBEngine {
	t.Helper()

	host := os.Getenv("TEAL_TEST_PG_HOST")
	if host == "" {
		t.Skip("TEAL_TEST_PG_HOST is not set, skipping the live PostgreSQL test")
	}

	port := 5432
	if value := os.Getenv("TEAL_TEST_PG_PORT"); value != "" {
		parsed, err := strconv.Atoi(value)
		require.NoError(t, err)
		port = parsed
	}

	// DBConnectionConfig.Config is an anonymous struct, so yaml is the least
	// painful way to build one.
	raw := fmt.Sprintf(`
name: test
type: postgres
config:
  host: %s
  port: %d
  database: %s
  user: %s
  password: %s
  db_sslnmode: disable
  pool_max_conns: %d
`,
		host, port,
		os.Getenv("TEAL_TEST_PG_DATABASE"),
		os.Getenv("TEAL_TEST_PG_USER"),
		os.Getenv("TEAL_TEST_PG_PASSWORD"),
		poolMaxConns)

	var connectionConfig configs.DBConnectionConfig
	require.NoError(t, yaml.Unmarshal([]byte(raw), &connectionConfig))

	engine := &PostgresDBEngine{dbConnection: &connectionConfig}
	require.NoError(t, engine.Connect())
	t.Cleanup(func() { engine.Close() })

	return engine
}

// TestPostgresCreateSchemaConcurrent reproduces docs/issues/008: several assets
// of the same stage start at once against a database where the stage schema does
// not exist yet. Each of them runs the sql_asset.go sequence - own pooled
// connection, own transaction, check, create - and none of them may fail.
func TestPostgresCreateSchemaConcurrent(t *testing.T) {
	const assets = 8

	engine := newTestPostgresEngine(t, assets)
	schemaName := "teal_test_concurrent_create"

	tx, err := engine.Begin()
	require.NoError(t, err)
	require.NoError(t, engine.Exec(tx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", schemaName)))
	require.NoError(t, engine.Commit(tx))

	errs := make([]error, assets)
	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := range assets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			tx, err := engine.Begin()
			if err != nil {
				errs[i] = err
				return
			}

			if !engine.CheckSchemaExists(tx, schemaName+".some_model") {
				if err := engine.CreateSchema(tx, schemaName); err != nil {
					errs[i] = err
					engine.Rollback(tx)
					return
				}
			}
			errs[i] = engine.Commit(tx)
		}()
	}

	close(start)
	wg.Wait()

	for i, err := range errs {
		assert.NoError(t, err, "asset %d failed to get the schema created", i)
	}

	tx, err = engine.Begin()
	require.NoError(t, err)
	assert.True(t, engine.CheckSchemaExists(tx, schemaName+".some_model"))
	require.NoError(t, engine.Exec(tx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", schemaName)))
	require.NoError(t, engine.Commit(tx))
}

// TestPostgresCreateSchemaIsIdempotent - the schema is already there, every
// caller still succeeds and the outer transaction stays usable afterwards.
func TestPostgresCreateSchemaIsIdempotent(t *testing.T) {
	engine := newTestPostgresEngine(t, 2)
	schemaName := "teal_test_idempotent_create"

	for range 3 {
		tx, err := engine.Begin()
		require.NoError(t, err)
		require.NoError(t, engine.CreateSchema(tx, schemaName))
		// the transaction must still be alive after a no-op create
		assert.True(t, engine.CheckSchemaExists(tx, schemaName+".some_model"))
		require.NoError(t, engine.Commit(tx))
	}

	tx, err := engine.Begin()
	require.NoError(t, err)
	require.NoError(t, engine.Exec(tx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", schemaName)))
	require.NoError(t, engine.Commit(tx))
}
