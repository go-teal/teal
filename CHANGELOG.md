# Changelog

## [1.2.11] 2026-07-19

### Added

- Debug UI shows an amber **"0 records"** hint under the asset toolbar when a
  Run Select or Data Preview succeeds but returns an empty result set, instead of
  the green success message

## [1.2.10] 2026-07-19

### Added

- **Data Preview** in the debug UI: reads an asset's materialized relation
  (`SELECT * FROM <name>`) with pagination and shows it in the Data panel — a
  plain "what's in the table" view, independent of the model's transformation SQL
  - New endpoint `POST /api/dag/asset/:name/preview`
  - Magnifying-glass button in the asset toolbar for table/incremental/view assets

## [1.2.9] 2026-07-19

### Added

- Asset action **icon toolbar** under the asset name (Run Select, Run Mutation,
  Truncate, Delete Persisted Data), reachable from any tab
- **Drag-to-pan** on the DAG canvas — hold the mouse button and drag to scroll

### Fixed

- Debug UI bundle was served with a one-year cache; since teal-ui builds fixed
  asset filenames, browsers kept serving a stale dashboard across upgrades. Assets
  are now served with `no-cache`

## [1.2.8] 2026-07-18

### Added

- **Drop / Truncate persisted data** in the debug UI
  - `POST /api/dag/asset/:name/drop` — DROP TABLE (table/incremental) / DROP VIEW (view)
  - `POST /api/dag/asset/:name/truncate` — empty the table (table/incremental)
  - Confirm-dialog buttons in the asset panel

### Fixed

- Data panel now follows the most recent write, so `RUN SELECTED QUERY` results
  are no longer hidden behind a prior mutation's task

## [1.2.7] 2026-07-18

### Fixed

- `teal gen` no longer clobbers an existing `Makefile` in a generated project

## [1.2.6] 2026-07-18

### Fixed

- Correct materialization icons in the dashboard: the gopher marks **raw Go**
  assets only; **custom** materialization gets a table icon and `.sql` extension

## [1.2.5] 2026-07-18

### Changed

- Vendored the teal-ui build that renders result columns in SQL order (UI side of
  the 1.2.2 fix)

## [1.2.4] 2026-07-18

### Fixed

- Asset select previews the materialized relation (`SELECT * FROM <name>`) for
  models whose SQL is not a single read-only SELECT (SCD2 / custom scripts),
  instead of failing or running the model's mutations

## [1.2.3] 2026-07-18

### Fixed

- Postgres `int4` / `int8` columns no longer degrade to null in DataFrames
  (gota only understands `[]int`)

## [1.2.2] 2026-07-18

### Fixed

- Preserve SQL column order in select/test data responses via a `columns` field
  (result rows are maps, so their JSON keys come back alphabetized)

## [1.2.1] 2026-07-18

### Fixed

- `teal ui` passes `-tags teal_ui` when spawning the generated project's UI process

## [1.2.0] 2026-07-18

### Changed

- Updated dependencies to their latest versions (gin, pgx/v5, zerolog, pongo2/v6,
  gin-contrib/cors, and transitive updates)

## [1.1.1] 2026-05-19

### Added

- `pkg/ui` is now gated behind the `teal_ui` build tag, so the default library
  build stays slim; documented the build tag in the README and templates

## [1.1.0] 2026-05-19

### Fixed

- PostgreSQL pipelines now work out of the box (drivers + generation)

## [1.0.2] 2025-11-03

### Added

- **Enhanced CLI Help System**: Comprehensive help output for all commands
  - Added `--help` and `-h` flag support for global and per-command help
  - Documented all commands: `init`, `gen`, `clean`, `ui`, `version`
  - Complete flag documentation with defaults and descriptions
  - Usage examples for each command
  - Detailed command-specific help with `teal [command] --help`

- **Dockerfile Generator**: Automatic multi-stage Dockerfile generation
  - Generated only once (skip if exists) to allow customization
  - Optimized for DuckDB compatibility using Debian bookworm base images
  - CGO-enabled builds for DuckDB native bindings support
  - Security best practices: non-root user with home directory
  - Final image size ~311MB with embedded DuckDB bindings
  - Can be modified for Alpine-based images (~20-30MB) when not using DuckDB

- **Enhanced Clean Command**: New flags for selective file cleanup
  - `--clean-dockerfile` - Delete generated Dockerfile
  - `--clean-main-ui` - Delete UI debug binary main file
  - `--clean-go-mod` - Delete go.mod and go.sum files
  - `--clean-all` - Delete ALL generated files (with confirmation)
  - Smart confirmation logic: specific file flags work independently of model cleaning

### Changed

- **README.md Documentation**: Comprehensive updates and additions
  - Added CLI Commands Reference section with all commands, flags, and examples
  - Added Docker Deployment section explaining Dockerfile design and image characteristics
  - Documented files NOT overwritten by `teal gen` (Dockerfile, go.mod, main files)
  - Updated project structure examples to include Dockerfile
  - Getting Help section showing how to access CLI help
  - Improved discoverability of all CLI features

## [1.0.1] 2025-11-03

### Changed

- **Logging Separation**: Refactored logging approach for better separation of concerns
  - `internal/` packages now use `fmt.Printf` and `panic` instead of zerolog
  - `cmd/teal` uses standard `fmt` output for CLI logging
  - `pkg/` packages continue to use zerolog for library-level structured logging
  - UI assets server (`internal/domain/services/ui_assets_server.go`) now uses simple fmt-based logging
  - Fatal errors in internal packages now use `panic()` instead of `log.Fatal()`
  - Error output uses `fmt.Fprintf(os.Stderr, ...)` for consistency

## [1.0.0] 2025-11-01

### Added

#### UI Dashboard & Development Tools

- **UI Dashboard**: Complete web-based visual interface for monitoring and controlling data pipelines
  - React-based dashboard with embedded assets served by teal CLI binary
  - Static asset server using `//go:embed` for zero-dependency deployment
  - Interactive DAG visualization showing all assets and their dependencies
  - Real-time execution monitoring and task status tracking
  - Test execution results and data quality checks viewer
  - Asset data inspection with pagination support
  - Execution logs viewer for debugging
- **Hot-Reload Development Server**: `teal ui` command with automatic file watching
  - Monitors changes in `assets/`, `profile.yaml`, and `config.yaml`
  - Automatic code regeneration when files change
  - Smart restart: Only API server restarts, UI Dashboard persists
  - Built-in debouncing to prevent excessive regenerations
  - Graceful shutdown with proper signal handling
  - AssetObserver manages lifecycle of both API and UI servers

#### Debug & Monitoring

- **Debug DAG**: Pointer-based architecture with comprehensive monitoring
  - Enhanced task tracking with execution state management
  - Detailed test result tracking per task execution
  - Connection management via REST API (connect/disconnect/status)
  - Asset-level execution and data retrieval
  - Root test execution with detailed result storage
- **REST API Enhancements**:
  - Individual test execution endpoint with data storage (`POST /api/tests/execute/:testName`)
  - Test data retrieval endpoint (`GET /api/tests/data/:testName`)
  - Asset selection/query endpoint (`POST /api/dag/asset/:name/select`)
  - Pagination support for asset data (`offset` and `limit` parameters)
  - README/documentation endpoint (`GET /api/docs/readme`)
  - Connection management endpoints (connect, disconnect, status)
- **Logging Improvements**:
  - Structured SQL execution logging with detailed fields
  - StoringConsoleWriter for capturing logs per task execution
  - Enhanced DuckDB logging optimization
  - Default log level updated to debug for better development experience

#### Template Engine & Code Generation

- **Pongo2 Template Engine**: Migration from Go's `text/template` to pongo2 (v6)
  - Django/Jinja2-compatible template syntax
  - Support for inline `profile.yaml` in SQL files using `{{ define "profile.yaml" }}`
  - Improved whitespace control and template rendering
  - Better control structures and filters
  - Fixed template rendering issues with pointer access
- **Mermaid Diagram Generation**: Migration from PlantUML (`.wsd`) to Mermaid (`.mmd`)
  - Modern, widely-supported diagram format
  - Better integration with documentation tools
  - Proper node ID sanitization for Mermaid compatibility

#### Model & Schema Enhancements

- **Description Fields**: Added optional description field to models and tests
  - Displayed in UI and API responses
  - Improves documentation and understanding of data assets
  - Supports markdown formatting
- **Generated Model Descriptors**: Enhanced with description metadata
- **Airline Example Dataset**: Comprehensive scaffold example with:
  - Multi-stage pipeline (staging → DDS → mart)
  - Fact and dimension tables
  - Test cases demonstrating data quality checks
  - CSV data files for immediate testing

### Fixed

- SIGTERM signal propagation to child process groups during hot-reload
- Child process cleanup before regeneration to prevent port conflicts
- Port release delays ensuring clean restarts
- Pongo2 template rendering with pointer struct field access
- README template whitespace control
- Asset execution interface refactoring
- Generated file cleanup and .gitignore improvements

### Changed

- **Template Syntax**: All templates now use pongo2 (Django/Jinja2 style)
  - `{{ }}` for variables and expressions
  - `{% %}` for control structures
  - Generation-time: `{{ Ref() }}`, `{{ this() }}`
  - Runtime: `{{ TaskID }}`, `{{ ENV() }}`, `{% if IsIncremental() %}`
- **Graph Output**: Changed from `.wsd` (PlantUML) to `.mmd` (Mermaid)
- **Development Workflow**: `teal ui` is now the recommended way to run debug server
- **DAG Execution Stages**: Reordered for better execution flow in scaffold configs
- **Default Log Level**: Changed to `debug` for improved development experience

### Documentation

- Complete UI Dashboard architecture documentation
- Comprehensive hot-reload and development server guide
- Pongo2 template engine usage and examples
- REST API endpoint reference
- AI assistant integration guide (Claude Code, Cursor, Copilot)
- CLAUDE.md with development best practices

## [0.2.0] 2024-12-11

### Added

- PostgreSQL support

## [0.1.10] 2024-12-05

### Added

- Indexes

## [0.1.9] 2024-10-17

### Fixed

- Minor typos
- Docs upadted

## [0.1.8] 2024-10-14

### Fixed

- DuckDB mutex dead lock bugfux

## [0.1.7] 2024-10-12

### Added

- duckdb optimizations

## [0.1.6] 2024-08-26

### Added

- raw go assets
- custom sql assets

### Fixed

- dataframe marshaling

## [0.1.5] 2024-07-30

### Fixed

- teal project generation fails: panic: open ./internal/model_tests/configs.go: no such file or directory

## [0.1.4] 2024-07-28

### Added

- Inline tests (tests after the execution of the model)

### Fixed

- Minor bug fixex

## [0.1.3] 2024-07-19

### Added

- Simple model testing

## [0.1.2] - 2024-07-12

### Added

- Cross database references
- Documentation update

## [0.1.1] - 2024-06-26

### Added

- Initial MVP release!
