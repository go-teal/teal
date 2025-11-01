# Changelog

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
