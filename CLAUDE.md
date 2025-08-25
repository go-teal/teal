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

# API endpoints available:
# GET  /api/dag                  - Get DAG structure
# GET  /api/tests                - Get test profiles
# POST /api/dag/run              - Execute DAG with task ID
# GET  /api/dag/status/:taskId   - Get task execution status
# GET  /api/dag/tasks            - List all task executions
# POST /api/dag/reset            - Clear execution history
```

## Development Notes

- Always check for existing CLAUDE.md improvements when using `teal init`
- Generated Go files use the module name from `config.yaml`
- DuckDB driver requires uncommenting import in generated main.go
- Test execution requires `InitChannelDagWithTests` instead of `InitChannelDag`
- Raw assets must be registered in main function before DAG execution
- Production and UI binaries are generated automatically by `teal gen`
- Task names in production are unique by default (timestamp-based) to support tracking