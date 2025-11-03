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
   - **Testing**: SQL-based tests that pass when returning zero rows. Test queries are automatically wrapped with `SELECT COUNT(*) ... HAVING count > 0 LIMIT 1` during code generation - users write only the constraint-checking SQL

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
   - **Important**: This binary only provides the Debug API server; the UI Dashboard is served by the teal CLI itself

This dual-generation approach ensures:
- Production binaries have no unnecessary dependencies
- Clean separation between production and debugging code
- Developers get powerful debugging tools without affecting production
- Both versions share the same asset and test implementations

### UI Dashboard Architecture

When running `teal ui`, two separate servers are started:

1. **UI Assets Server** (port 8081 by default):
   - Embedded in the `teal` CLI binary itself (NOT in generated projects)
   - Located in: `internal/domain/services/ui_assets_server.go` in teal source
   - Serves static React-based dashboard files using `//go:embed dist`
   - **Persists across API server restarts** during hot-reload
   - Managed by `AssetObserver` in `internal/domain/services/asset_observer.go`

2. **Debug API Server** (port 8080 by default):
   - Located in generated project: `cmd/<project-name>-ui/<project-name>-ui.go`
   - Provides REST API endpoints for DAG operations, tests, logs, and data access
   - **Restarts automatically** when assets, config, or profile files change
   - Spawned as a child process by `teal ui` command

This architecture ensures:
- UI Dashboard remains responsive during code regeneration and API server restarts
- Zero-dependency deployment (all UI assets embedded in teal binary)
- Clean separation between static frontend (teal binary) and dynamic API (generated project)

### Key Concepts

- **Template Engine**: Uses pongo2/v6 (Django-compatible template syntax with `{{ }}` and `{% %}`)
- **Ref Function**: `{{ Ref("stage.model") }}` creates DAG dependencies and is resolved at generation-time
- **this Function**: `{{ this() }}` returns current model name and is resolved at generation-time
- **Runtime Variables**: `{{ TaskID }}`, `{{ TaskUUID }}`, `{{ ENV(...) }}` are evaluated during DAG execution
- **Materializations**: table, incremental, view, custom, raw
- **Cross-DB References**: Models can consume data from different database connections using dataframes
- **Model Profiles**: Configuration embedded in SQL files or profile.yaml
- **Graph Visualization**: Generates Mermaid (.mmd) diagrams for DAG visualization

### Template System

Teal uses **pongo2/v6**, a Django-compatible template engine for Go. All code generation and runtime SQL rendering uses pongo2 syntax.

**Template Syntax:**
- Variables: `{{ variable }}`
- Control structures: `{% if condition %}...{% endif %}`, `{% for item in items %}...{% endfor %}`
- Filters: `{{ variable|lower }}`, `{{ variable|safe }}`
- Whitespace control: `{%- ... -%}` to trim surrounding whitespace
- Comments: `{# This is a comment #}`

**Code Generation Templates:**
All generator templates in `internal/domain/generators/templates/` use pongo2:
- `dwh_sql_model_asset.tmpl` - Generates SQL model Go code
- `dwh_raw_model_asset.tmpl` - Generates raw asset Go code
- `graph.mmd.tmpl` - Generates Mermaid DAG diagrams
- `readme.md.tmpl` - Generates project documentation
- `go.mod.tmpl`, `testing_config.go.tmpl`, etc.

**Important Code Generation Details:**
- Model descriptions are base64-encoded before template execution to avoid special character issues
- Field access on pointer structs requires extracting values to direct variables (pongo2 limitation)
- Use `pongo2.Context` (map[string]interface{}) instead of structs for template data
- Template functions are registered as context values, not as FuncMap

**Example Generator Pattern:**
```go
// Extract pointer fields to direct variables
var materialization string
if modelConfig.ModelProfile != nil {
    materialization = string(modelConfig.ModelProfile.Materialization)
}

// Base64 encode descriptions
if modelConfig.ModelProfile != nil && modelConfig.ModelProfile.Description != "" {
    encoded := base64.StdEncoding.EncodeToString([]byte(modelConfig.ModelProfile.Description))
    modelConfig.ModelProfile.Description = encoded
}

// Execute template with pongo2.Context
output, err := tmpl.Execute(pongo2.Context{
    "ModelName":       modelConfig.ModelName,
    "Materialization": materialization,  // Direct variable, not pointer access
    "ModelProfile":    modelConfig.ModelProfile,
})
```

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
- `GET  /api/tests/results/:taskId` - Get test execution results for specific task
- `POST /api/tests/execute/:testName` - Execute individual test query independently
  - Request body: `{"taskId": "task_123"}`
  - Response:
    ```json
    {
      "testName": "dds.test_dim_airports_unique",
      "taskId": "task_123",
      "status": "SUCCESS|FAILED",
      "rowCount": 0,
      "errorMsg": "",
      "durationMs": 45,
      "executedAt": "2024-01-01T12:00:00Z"
    }
    ```
  - Test success logic: 0 rows = SUCCESS (no constraint violations), >0 rows = FAILED (violations found)
- `GET  /api/tests/data/:testName?taskId=task_123` - Get test execution data for specific task
  - Response includes actual query results (rows that failed the test):
    ```json
    {
      "testName": "dds.test_dim_airports_unique",
      "taskId": "task_123",
      "status": "SUCCESS|FAILED",
      "rowCount": 2,
      "data": [
        {"airport_id": 1, "count": 2},
        {"airport_id": 5, "count": 3}
      ],
      "executedAt": "2024-01-01T12:00:00Z"
    }
    ```
  - Returns 404 if test hasn't been executed for this taskId

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

### Git Workflow

#### Conventional Commits

This project follows the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for commit messages.

**Commit Message Format:**
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Common Types:**
- `feat`: New feature (correlates to MINOR in semantic versioning)
- `fix`: Bug fix (correlates to PATCH in semantic versioning)
- `docs`: Documentation changes
- `refactor`: Code refactoring without functionality changes
- `test`: Adding or updating tests
- `chore`: Maintenance tasks, dependencies, build process
- `perf`: Performance improvements
- `style`: Code style/formatting changes
- `ci`: CI/CD configuration changes
- `build`: Build system or external dependencies changes

**Scopes (optional but recommended):**
- `cli`: CLI command changes
- `gen`: Code generation
- `dag`: DAG execution engine
- `drivers`: Database drivers (duckdb, postgres)
- `ui`: UI Dashboard
- `api`: REST API endpoints
- `templates`: Template changes
- `tests`: Testing infrastructure
- `docs`: Documentation

**Breaking Changes:**
Indicate breaking changes with `!` after type/scope or in footer:
```
feat(api)!: remove deprecated endpoint

BREAKING CHANGE: /api/old-endpoint has been removed, use /api/new-endpoint instead
```

**Examples:**
```bash
# Feature additions
feat(cli): add comprehensive help system for all commands
feat(drivers): add MySQL database driver support
feat(ui): add real-time log streaming to dashboard

# Bug fixes
fix(dag): resolve deadlock in channel-based execution
fix(postgres): handle SSL certificate validation errors
fix(gen): correct template rendering for pointer fields

# Documentation
docs(readme): add CLI commands reference section
docs(api): document test execution endpoints

# Refactoring
refactor(drivers): extract common DataFrame operations
refactor(templates): migrate from text/template to pongo2

# Chores
chore: update dependencies to latest versions
chore(release): bump version to v1.0.2

# Multiple scopes
feat(cli,docs): enhance help system and update documentation
```

**Full Commit Example:**
```bash
git commit -m "$(cat <<'EOF'
feat(cli): add comprehensive help system for all commands

- Add --help and -h flag support for global and per-command help
- Document all commands: init, gen, clean, ui, version
- Complete flag documentation with defaults and examples
- Usage examples for each command

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

#### Push Policy
- **IMPORTANT**: NEVER run `git push` unless the user explicitly requests it
- You may stage files (`git add`) and create commits (`git commit`) when appropriate
- Always wait for explicit permission before pushing changes to remote repository

### General
- Always check for existing CLAUDE.md improvements when using `teal init`
- Generated Go files use the module name from `config.yaml`
- DuckDB driver requires uncommenting import in generated main.go
- Test execution requires `InitChannelDagWithTests` instead of `InitChannelDag`
- Raw assets must be registered in main function before DAG execution
- Production and UI binaries are generated automatically by `teal gen`
- Task names in production are unique by default (timestamp-based) to support tracking
- When implementing new database drivers, ensure proper DataFrame serialization for cross-database compatibility
- Database drivers are initialized once and reused throughout the DAG execution

### Testing
- Write test SQL to return rows that violate constraints (test passes if 0 rows, fails if >0 rows)
- Do NOT add `LIMIT 1` or `HAVING test_count > 0` - the code generator automatically wraps test queries with `SELECT COUNT(*) ... HAVING count > 0 LIMIT 1`
- Test template: `internal/domain/generators/templates/dwh_sql_model_test.tmpl`
- Generated tests create two SQL constants: `RAW_SQL_*` (user's query) and `COUNT_TEST_SQL_*` (wrapped version)

### Template Engine (pongo2)
- All templates use pongo2/v6 (Django-compatible syntax)
- Cannot access fields on pointer structs directly in templates - extract to variables first
- Always use `pongo2.Context` (map) for template data, never structs
- Runtime template functions in `pkg/processing/functions.go`:
  - `MergePongo2Context()` - Merges multiple pongo2 contexts
  - `FromTaskContextPongo2()` - Converts task context to pongo2 context
  - `FromConnectionContext()` - Database connection context (returns pongo2.Context)
- Generation-time evaluation: `{{ Ref("stage.model") }}` and `{{ this() }}` are evaluated during `teal gen`
- Runtime evaluation: `{{ TaskID }}`, `{{ TaskUUID }}`, `{{ ENV(...) }}`, and control structures are evaluated during execution
- Use `|safe` filter for SQL to prevent HTML escaping
- Whitespace control: `{%- endif -%}` removes surrounding blank lines

### Graph Visualization
- Generates Mermaid (.mmd) diagrams, not PlantUML (.wsd)
- Node IDs must be sanitized (dots → underscores) for Mermaid compatibility
- Graph generator creates `GraphNode` struct with pre-sanitized IDs
- Template: `internal/domain/generators/templates/graph.mmd.tmpl`
- Output: `docs/graph.mmd` in generated projects