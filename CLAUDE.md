# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Building and Installing
```bash
# Build the teal CLI tool
make

# Install teal CLI globally
make install

# Or directly with go
go build -o bin/teal ./cmd/teal
go install ./cmd/teal
```

### Testing Generated Projects
```bash
# Generate test project from scaffold
make test

# Clean generated test files  
make test_clean
```

### Running Teal Commands
```bash
# Initialize a new project
teal init

# Generate Go code from SQL models
teal gen [--project-path=<path>] [--config-file=<path>]

# Clean generated files
teal clean [--project-path=<path>] [--clean-main]

# Show version
teal version
```

## Architecture Overview

Teal is a Go-based ETL tool that generates Go code from SQL models, creating data pipelines with DAG-based execution. The architecture consists of:

### Core Components

1. **CLI Tool (`cmd/teal/`)**: Entry point for all Teal commands (init, gen, clean, version)

2. **Asset Generation Pipeline (`internal/application/`, `internal/domain/`)**: 
   - Parses SQL models from `assets/models/<stage>/*.sql`
   - Extracts profiles from YAML templates within SQL files
   - Generates Go code for each model with proper DAG dependencies
   - Creates test assets from `assets/tests/` directory

3. **DAG Execution Engine (`pkg/dags/`)**: 
   - `ChannelDag`: Implements concurrent execution using Go channels and goroutines
   - Each node represents an asset, edges are Go channels for data flow
   - Supports both with and without test execution modes

4. **Database Drivers (`pkg/drivers/`)**:
   - **DuckDB**: Supports extensions, dataframe operations, and custom configurations
   - **PostgreSQL**: Full SSL support, environment variable configuration
   - Factory pattern for driver instantiation
   - Cross-database references via `gota.DataFrame` when `is_data_framed: true`

5. **Processing Layer (`pkg/processing/`)**:
   - **SQL Assets**: Table, incremental, view, and custom materializations
   - **Raw Assets**: Custom Go functions implementing `ExecutorFunc` interface
   - **Testing**: Simple SQL-based tests that pass when returning zero rows

### Project Structure After Generation

```
project/
├── assets/models/          # SQL model files organized by stages
│   ├── staging/
│   ├── dds/
│   └── mart/
├── assets/tests/           # SQL test files
├── cmd/
│   ├── <project-name>/     # Production binary (Channel DAG only)
│   │   └── <project-name>.go
│   └── <project-name>-ui/  # Debug UI binary (Debug DAG + UI server)
│       └── <project-name>-ui.go
├── internal/assets/        # Generated Go code for each model
│   ├── configs.go         # DAG configuration
│   └── <stage>.<model>.go # Individual model implementations
├── internal/model_tests/   # Generated test implementations
├── config.yaml            # Database connections configuration
└── profile.yaml           # Project and model profiles
```

### File Generation Principles

The `teal gen` command generates two separate main entry points to ensure clean separation of concerns:

1. **Production Binary (`cmd/<project-name>/`)**:
   - Uses Channel DAG for efficient concurrent execution
   - No UI or debugging dependencies
   - Generates unique task names with timestamps (format: `<project_name>_<timestamp>`)
   - Optimized for production deployments
   - Supports custom task names via `--task-name` flag

2. **Debug UI Binary (`cmd/<project-name>-ui/`)**:
   - Uses Debug DAG for visualization and monitoring
   - Includes UI server with REST API endpoints
   - Provides execution tracking and task history
   - Designed for development and debugging
   - Runs on configurable port (default 8080)

This dual-generation approach ensures:
- Production binaries have no unnecessary dependencies
- Clean separation between production and debugging code
- Developers get powerful debugging tools without affecting production
- Both versions share the same asset and test implementations

### Key Concepts

- **Ref Function**: Template function `{{ Ref "stage.model" }}` creates DAG dependencies
- **Materializations**: table, incremental, view, custom, raw
- **Cross-DB References**: Models can consume data from different database connections using dataframes
- **Model Profiles**: Configuration embedded in SQL files or profile.yaml
- **Static vs Dynamic Templates**: `{{}}` for generation-time, `{{{}}}` for runtime execution

## Usage Examples

### Running Production Binary
```bash
# Run with auto-generated task name
go run cmd/<project-name>/<project-name>.go --input-data '{"key":"value"}' --with-tests

# Run with custom task name
go run cmd/<project-name>/<project-name>.go --task-name "etl_batch_001" --log-output raw

# Build and run production binary
go build -o bin/<project-name> cmd/<project-name>/<project-name>.go
./bin/<project-name> --log-level info
```

### Running Debug UI Binary
```bash
# Start UI server on default port 8080
go run cmd/<project-name>-ui/<project-name>-ui.go

# Start on custom port
go run cmd/<project-name>-ui/<project-name>-ui.go --port 9090
```

### API Endpoints

#### DAG Operations
- `GET  /api/dag` - Get DAG structure with all nodes and connections
- `POST /api/dag/run` - Execute DAG with task ID (returns 200 OK even if tests fail)
  - Request body: `{"taskId": "task_123", "data": {}}`
  - Returns 200 for successful execution (including test failures)
  - Returns 202 for pending/timeout
  - Response includes `rootTestResults` array with root test execution results:
    ```json
    {
      "taskId": "task_123",
      "status": "SUCCESS",
      "rootTestResults": [
        {
          "testName": "root.test_name",
          "status": "SUCCESS|FAILED",
          "errorMsg": "error message if failed",
          "durationMs": 123
        }
      ]
    }
    ```
- `GET  /api/dag/status/:taskId` - Get task execution status
  - Response includes `rootTestResults` array with root test execution results (same structure as above)
- `GET  /api/dag/tasks` - List all task executions
- `POST /api/dag/reset` - Clear execution history

#### Test Operations  
- `GET  /api/tests` - Get all test profiles (returns name, SQL, connectionName, connectionType)
- `GET  /api/tests/:taskId` - Get test execution results for specific task

#### Asset Operations
- `POST /api/dag/asset/:name/execute` - Execute specific asset
- `GET  /api/dag/asset/:name/data` - Get asset data/results

#### Log Operations (when UI mode is enabled)
- `GET  /api/logs/:taskId` - Get logs for specific task execution
- `GET  /api/logs` - Get all logs
- `DELETE /api/logs/:taskId` - Clear logs for specific task
- `DELETE /api/logs` - Clear all logs

## Database Driver Architecture

### Pattern Overview

Teal uses a clean interface-based abstraction pattern for database drivers, implementing both Strategy and Factory patterns to support multiple database engines seamlessly.

### Core Components

1. **DBDriver Interface** (`pkg/drivers/dbdriver.go`):
   - Defines the contract all database drivers must implement
   - Methods for connection management, transactions, SQL execution, DataFrame operations
   - Schema introspection capabilities (table/schema existence, field listing)
   - Cross-database data movement via DataFrames

2. **Factory Pattern** (`pkg/drivers/factory.go`):
   - Central registry of available drivers
   - Each database type has its own factory implementation
   - Supports dynamic driver registration via `RegisterConnectionFactory()`
   - Clean instantiation through `EstablishDBConnection()`

3. **Supported Databases**:
   - **DuckDB** (`duckdb.go`): In-memory/file-based analytical database
     - Extension support (auto-install and load)
     - Source mounting for federated queries
     - Mutex-based concurrency control (single-writer limitation)
     - DataFrame integration via `duckdb_dataframe.go`
   
   - **PostgreSQL** (`postgres.go`): Production-grade relational database
     - Full SSL/TLS support with certificate authentication
     - Native pgx v5 driver for optimal performance
     - Context-aware operations for proper cancellation
     - DataFrame integration via `postgres_dataframe.go`

### Adding a New Database Driver

To add support for a new database, implement the `DBDriver` interface:

```go
type CustomDBEngine struct {
    dbConnection *configs.DBConnectionConfig
    // your database connection
}

// Implement all DBDriver interface methods
func (c *CustomDBEngine) Connect() error { /* ... */ }
func (c *CustomDBEngine) Begin() (interface{}, error) { /* ... */ }
func (c *CustomDBEngine) ToDataFrame(sql string) (*dataframe.DataFrame, error) { /* ... */ }
// ... other methods

// Create a factory
type CustomDBEngineFactory struct{}

func (c *CustomDBEngineFactory) CreateConnection(connection configs.DBConnectionConfig) (DBDriver, error) {
    return &CustomDBEngine{dbConnection: &connection}, nil
}

// Register in your init function or main
func init() {
    drivers.RegisterConnectionFactory("customdb", &CustomDBEngineFactory{})
}
```

### Key Design Decisions

1. **Generic Transaction Handling**: Uses `interface{}` for transactions to accommodate different transaction types (sql.Tx, pgx.Tx, etc.)

2. **DataFrame as Universal Data Format**: All drivers support converting query results to DataFrames, enabling cross-database data movement

3. **Concurrency Control**: Database-specific (DuckDB needs mutex, PostgreSQL uses MVCC)

4. **Configuration Flexibility**: Each driver can have custom configuration fields in `config.yaml`

### Usage in Assets

Generated assets automatically use the appropriate driver based on the connection configuration:

```yaml
# config.yaml
connections:
  - name: analytics_duck
    type: duckdb
    config:
      path: "./analytics.db"
      extensions: ["parquet", "json"]
  
  - name: production_pg
    type: postgres
    config:
      host: localhost
      port: 5432
      database: production
      user: etl_user
```

Assets can then reference these connections and data flows seamlessly between them when using DataFrames.

## Development Notes

- Always check for existing CLAUDE.md improvements when using `teal init`
- Generated Go files use the module name from `config.yaml`
- DuckDB driver requires uncommenting import in generated main.go
- Test execution requires `InitChannelDagWithTests` instead of `InitChannelDag`
- Raw assets must be registered in main function before DAG execution
- Production and UI binaries are generated automatically by `teal gen`
- Task names in production are unique by default (timestamp-based) to support tracking
- When implementing new database drivers, ensure proper DataFrame serialization for cross-database compatibility
- Database drivers are initialized once and reused throughout the DAG execution