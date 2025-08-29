# API Specifications

This document provides comprehensive API specifications for the Teal Debug UI server endpoints.

## Base URL
```
http://localhost:8080
```

## Table of Contents
- [DAG Operations](#dag-operations)
  - [GET /api/dag](#get-apidag)
  - [POST /api/dag/run](#post-apidagrun)
  - [GET /api/dag/status/:taskId](#get-apidagstatustaskid)
  - [GET /api/dag/tasks](#get-apidagtasks)
  - [POST /api/dag/reset](#post-apidagreset)
- [Asset Operations](#asset-operations)
  - [POST /api/dag/asset/:name/execute](#post-apidagassetnameexecute)
  - [GET /api/dag/asset/:name/data](#get-apidagassetnamedata)
- [Test Operations](#test-operations)
  - [GET /api/tests](#get-apitests)
- [Log Operations](#log-operations)
  - [GET /api/logs/:taskId](#get-apilogstaskid)
  - [GET /api/logs](#get-apilogs)
  - [DELETE /api/logs/:taskId](#delete-apilogstaskid)
  - [DELETE /api/logs](#delete-apilogs)

---

## DAG Operations

### GET /api/dag
Retrieves the complete DAG structure with all nodes and their relationships.

**Response: 200 OK**
```json
{
  "projectName": "hello-world",
  "moduleName": "github.com/you_git_user/your_project",
  "dagInstanceName": "hello-world-dag",
  "nodes": [
    {
      "name": "staging.hello",
      "downstreams": ["dds.world"],
      "upstreams": [],
      "sqlSelectQuery": "SELECT 'Hello' as greeting, 1 as id",
      "sqlCompiledQuery": "CREATE TABLE IF NOT EXISTS staging.hello AS SELECT 'Hello' as greeting, 1 as id",
      "materialization": "table",
      "connectionType": "duckdb",
      "connectionName": "memory_duck",
      "isDataFramed": false,
      "persistInputs": false,
      "tests": ["test_hello_exists"],
      "state": "INITIAL",
      "totalTests": 1,
      "successfulTests": 0,
      "lastExecutionDuration": 0,
      "lastTestsDuration": 0
    }
  ]
}
```

**Field Descriptions:**
- `projectName` (string): Name of the Teal project
- `moduleName` (string): Go module name from go.mod
- `dagInstanceName` (string): Name of the DAG instance
- `nodes` (array): Array of DAG node objects
  - `name` (string): Unique identifier for the node (format: "stage.model")
  - `downstreams` (array): Names of nodes that depend on this node
  - `upstreams` (array): Names of nodes this node depends on
  - `sqlSelectQuery` (string): Original SQL SELECT query
  - `sqlCompiledQuery` (string): Compiled SQL with materialization
  - `materialization` (string): Type of materialization - "table", "incremental", "view", "custom", "raw"
  - `connectionType` (string): Database type - "duckdb", "postgres", etc.
  - `connectionName` (string): Connection identifier from config.yaml
  - `isDataFramed` (boolean): Whether data is passed as DataFrame
  - `persistInputs` (boolean): Whether to persist input data
  - `tests` (array): Names of tests associated with this node
  - `state` (string): Current execution state - "INITIAL", "IN_PROGRESS", "TESTING", "FAILED", "SUCCESS"
  - `totalTests` (integer): Total number of tests for this node
  - `successfulTests` (integer): Number of tests that passed
  - `lastExecutionDuration` (integer): Last execution time in milliseconds
  - `lastTestsDuration` (integer): Last test execution time in milliseconds

---

### POST /api/dag/run
Triggers DAG execution with optional input data. Returns within 10 seconds with current status.

**Request Body:**
```json
{
  "taskId": "task_20250125_143022",
  "data": {
    "input_param1": "value1",
    "input_param2": 42
  }
}
```

**Response: 200 OK (Completed)**
```json
{
  "taskId": "task_20250125_143022",
  "status": "SUCCESS",
  "completedAssets": 2,
  "totalAssets": 2,
  "failedAssets": 0,
  "inProgressAssets": 0,
  "tasks": [
    {
      "name": "Executing staging.hello",
      "state": "SUCCESS",
      "order": 1,
      "startTime": 1737816622000,
      "endTime": 1737816623500,
      "executionTimeMs": 1500,
      "message": "",
      "totalTests": 1,
      "passedTests": 1,
      "failedTests": 0
    },
    {
      "name": "Executing dds.world",
      "state": "SUCCESS",
      "order": 2,
      "startTime": 1737816623500,
      "endTime": 1737816625000,
      "executionTimeMs": 1500,
      "message": "",
      "totalTests": 2,
      "passedTests": 2,
      "failedTests": 0
    }
  ],
  "lastTaskName": "Executing dds.world"
}
```

**Response: 202 Accepted (Still Running)**
```json
{
  "taskId": "task_20250125_143022",
  "status": "PENDING",
  "completedAssets": 1,
  "totalAssets": 2,
  "failedAssets": 0,
  "inProgressAssets": 1,
  "tasks": [
    {
      "name": "Executing staging.hello",
      "state": "SUCCESS",
      "order": 1,
      "startTime": 1737816622000,
      "endTime": 1737816623500,
      "executionTimeMs": 1500,
      "totalTests": 1,
      "passedTests": 1,
      "failedTests": 0,
      "testResults": [
        {
          "testName": "test_hello_not_empty",
          "status": "SUCCESS",
          "durationMs": 45
        }
      ]
    },
    {
      "name": "Executing dds.world",
      "state": "IN_PROGRESS",
      "order": 2,
      "startTime": 1737816623500,
      "totalTests": 2,
      "passedTests": 0,
      "failedTests": 0,
      "testResults": []
    }
  ],
  "lastTaskName": "Executing dds.world"
}
```

**Response: 500 Internal Server Error (Failed)**
```json
{
  "taskId": "task_20250125_143022",
  "status": "FAILED",
  "completedAssets": 0,
  "totalAssets": 1,
  "failedAssets": 1,
  "inProgressAssets": 0,
  "nodes": [
    {
      "name": "Executing staging.hello",
      "state": "FAILED",
      "order": 1,
      "startTime": 1737816622000,
      "endTime": 1737816623500,
      "executionTimeMs": 1500,
      "message": "SQL execution failed: table not found",
      "totalTests": 1,
      "passedTests": 0,
      "failedTests": 1,
      "testResults": []
    }
  ],
  "lastTaskName": "Executing staging.hello"
}
```

**Field Descriptions:**
- `taskId` (string, required): Unique identifier for this execution
- `data` (object, optional): Input data to pass to the DAG
- `status` (string): Overall execution status - "NOT_STARTED", "IN_PROGRESS", "SUCCESS", "FAILED", "PENDING"
- `completedAssets` (integer): Total number of completed assets across all tasks
- `totalAssets` (integer): Total number of assets in the DAG
- `failedAssets` (integer): Total number of failed assets
- `inProgressAssets` (integer): Total number of assets currently executing
- `nodes` (array): Array of node execution status objects
  - `name` (string): Node/asset name
  - `state` (string): Node state - "INITIAL", "IN_PROGRESS", "TESTING", "FAILED", "SUCCESS"
  - `order` (integer): Execution order (1-based)
  - `startTime` (integer, optional): Unix timestamp in milliseconds
  - `endTime` (integer, optional): Unix timestamp in milliseconds
  - `executionTimeMs` (integer): Execution duration in milliseconds
  - `message` (string): Error or status message
  - `totalTests` (integer): Total number of tests for this node
  - `passedTests` (integer): Number of tests that passed
  - `failedTests` (integer): Number of tests that failed
  - `testResults` (array, optional): Array of individual test results
    - `testName` (string): Name of the test
    - `status` (string): Test status - "SUCCESS", "FAILED", "NOT_FOUND"
    - `error` (string, optional): Error message if test failed
    - `durationMs` (integer): Test execution duration in milliseconds
- `lastTaskName` (string): Name of the last executed task

---

### GET /api/dag/status/:taskId
Retrieves the current status of a specific task execution.

**Parameters:**
- `taskId` (path parameter): The task ID to query

**Response: 200 OK**
```json
{
  "taskId": "task_20250125_143022",
  "status": "SUCCESS",
  "completedAssets": 1,
  "totalAssets": 1,
  "failedAssets": 0,
  "inProgressAssets": 0,
  "nodes": [
    {
      "name": "Executing staging.hello",
      "state": "SUCCESS",
      "order": 1,
      "startTime": 1737816622000,
      "endTime": 1737816623500,
      "executionTimeMs": 1500,
      "totalTests": 2,
      "passedTests": 2,
      "failedTests": 0,
      "testResults": [
        {
          "testName": "test_hello_not_empty",
          "status": "SUCCESS",
          "durationMs": 45
        },
        {
          "testName": "test_hello_has_greeting",
          "status": "SUCCESS",
          "durationMs": 32
        }
      ]
    }
  ],
  "lastTaskName": "Executing staging.hello"
}
```

**Response: 404 Not Found**
```json
{
  "taskId": "unknown_task",
  "status": "NOT_STARTED",
  "nodes": [],
  "lastTaskName": ""
}
```

---

### GET /api/dag/tasks
Retrieves a list of all task executions, sorted by start time (most recent first).

**Response: 200 OK**
```json
{
  "tasks": [
    {
      "taskId": "task_20250125_143022",
      "status": "SUCCESS",
      "startTime": 1737816622000,
      "endTime": 1737816625000,
      "totalAssets": 2,
      "completedAssets": 2,
      "failedAssets": 0,
      "inProgressAssets": 0
    },
    {
      "taskId": "task_20250125_142015",
      "status": "FAILED",
      "startTime": 1737816015000,
      "endTime": 1737816018000,
      "totalAssets": 2,
      "completedAssets": 1,
      "failedAssets": 1,
      "inProgressAssets": 0
    }
  ],
  "total": 2
}
```

**Field Descriptions:**
- `tasks` (array): Array of task summary objects
  - `taskId` (string): Unique task identifier
  - `status` (string): Overall task status
  - `startTime` (integer, optional): Unix timestamp in milliseconds
  - `endTime` (integer, optional): Unix timestamp in milliseconds
  - `totalAssets` (integer): Total number of assets in the task
  - `completedAssets` (integer): Number of successfully completed assets
  - `failedAssets` (integer): Number of failed assets
  - `inProgressAssets` (integer): Number of assets currently executing
- `total` (integer): Total number of tasks

---

### POST /api/dag/reset
Resets the DAG state, clearing all execution data and task history.

**Response: 200 OK**
```json
{
  "message": "DAG state reset successfully",
  "status": "SUCCESS"
}
```

**Response: 500 Internal Server Error**
```json
{
  "error": "Failed to reset DAG state",
  "details": "Cannot reset while execution is in progress"
}
```

---

## Asset Operations

### POST /api/dag/asset/:name/execute
Executes a specific asset within the DAG context. Returns within 10 seconds.

**Parameters:**
- `name` (path parameter): Asset name (e.g., "staging.hello")

**Request Body:**
```json
{
  "taskId": "task_20250125_143022"
}
```

**Response: 200 OK (Completed)**
```json
{
  "taskId": "task_20250125_143022",
  "assetName": "staging.hello",
  "status": "SUCCESS",
  "startTime": 1737816622000,
  "endTime": 1737816623500,
  "executionTimeMs": 1500,
  "result": {
    "rows_affected": 100,
    "data": [{"greeting": "Hello", "id": 1}]
  },
  "error": "",
  "upstreamsUsed": []
}
```

**Response: 202 Accepted (Still Running)**
```json
{
  "taskId": "task_20250125_143022",
  "assetName": "staging.hello",
  "status": "IN_PROGRESS",
  "startTime": 1737816622000,
  "upstreamsUsed": []
}
```

**Response: 500 Internal Server Error (Failed)**
```json
{
  "taskId": "task_20250125_143022",
  "assetName": "staging.hello",
  "status": "FAILED",
  "startTime": 1737816622000,
  "endTime": 1737816623500,
  "executionTimeMs": 1500,
  "error": "SQL execution failed: syntax error at line 5",
  "upstreamsUsed": []
}
```

**Field Descriptions:**
- `taskId` (string, required): Task ID for tracking
- `assetName` (string): Name of the asset being executed
- `status` (string): Execution status - "INITIAL", "IN_PROGRESS", "TESTING", "FAILED", "SUCCESS"
- `startTime` (integer, optional): Unix timestamp in milliseconds
- `endTime` (integer, optional): Unix timestamp in milliseconds
- `executionTimeMs` (integer): Execution duration in milliseconds
- `result` (object, optional): Execution result data
- `error` (string): Error message if failed
- `upstreamsUsed` (array): List of upstream assets used

---

### GET /api/dag/asset/:name/data
Retrieves the current data stored in an asset after execution.

**Parameters:**
- `name` (path parameter): Asset name (e.g., "staging.hello")

**Response: 200 OK (DataFrame)**
```json
{
  "assetName": "staging.hello",
  "hasData": true,
  "dataType": "dataframe",
  "isDataFramed": true,
  "data": [
    {"greeting": "Hello", "id": 1},
    {"greeting": "World", "id": 2}
  ],
  "rowCount": 2,
  "columnCount": 2,
  "columns": ["greeting", "id"],
  "error": ""
}
```

**Response: 200 OK (Map Data)**
```json
{
  "assetName": "dds.world",
  "hasData": true,
  "dataType": "map",
  "isDataFramed": false,
  "data": {
    "key1": "value1",
    "key2": 42,
    "nested": {
      "field": "data"
    }
  },
  "error": ""
}
```

**Response: 204 No Content**
```json
{
  "assetName": "staging.hello",
  "hasData": false,
  "dataType": "",
  "isDataFramed": false,
  "error": ""
}
```

**Response: 404 Not Found**
```json
{
  "assetName": "unknown.asset",
  "hasData": false,
  "dataType": "",
  "isDataFramed": false,
  "error": "Asset not found in DAG"
}
```

**Field Descriptions:**
- `assetName` (string): Name of the asset
- `hasData` (boolean): Whether the asset has data
- `dataType` (string): Type of data - "dataframe", "map", "array", "string", etc.
- `isDataFramed` (boolean): Whether data is stored as DataFrame
- `data` (any, optional): The actual data (structure depends on dataType)
- `rowCount` (integer, optional): Number of rows (for dataframes)
- `columnCount` (integer, optional): Number of columns (for dataframes)
- `columns` (array, optional): Column names (for dataframes)
- `error` (string): Error message if retrieval failed

---

## Test Operations

### GET /api/tests
Retrieves all test profiles defined in the DAG.

**Response: 200 OK**
```json
{
  "tests": [
    {
      "name": "test_hello_exists",
      "sql": "SELECT * FROM staging.hello WHERE greeting IS NULL",
      "connectionName": "memory_duck",
      "connectionType": "duckdb",
      "status": "INITIAL"
    },
    {
      "name": "test_world_count",
      "sql": "SELECT * FROM dds.world WHERE count < 0",
      "connectionName": "memory_duck",
      "connectionType": "duckdb",
      "status": "SUCCESS"
    }
  ]
}
```

**Field Descriptions:**
- `tests` (array): Array of test profile objects
  - `name` (string): Unique test identifier
  - `sql` (string): SQL query that should return zero rows to pass
  - `connectionName` (string): Database connection to use
  - `connectionType` (string): Type of database connection
  - `status` (string): Test execution status - "INITIAL", "IN_PROGRESS", "FAILED", "SUCCESS"

---

## Log Operations

The log endpoints are only available when the UI server is started with the StoringConsoleWriter logger configured (default in UI mode). These endpoints provide access to structured log data captured during DAG execution, organized by task ID.

### GET /api/logs/:taskId
Retrieves all log entries for a specific task.

**Parameters:**
- `taskId` (path parameter): The task ID to retrieve logs for

**Response: 200 OK**
```json
{
  "taskId": "task_20250125_143022",
  "logs": [
    {
      "level": "info",
      "time": "2025-01-25T14:30:22Z",
      "message": "Starting DAG execution",
      "taskId": "task_20250125_143022",
      "stage": "initialization"
    },
    {
      "level": "debug",
      "time": "2025-01-25T14:30:23Z",
      "message": "Executing asset: staging.hello",
      "taskId": "task_20250125_143022",
      "asset": "staging.hello",
      "connection": "memory_duck"
    }
  ],
  "count": 2
}
```

**Response: 501 Not Implemented (Logger not configured)**
```json
{
  "error": "Log writer not configured"
}
```

**Field Descriptions:**
- `taskId` (string): The task ID requested
- `logs` (array): Array of log entry objects with varying fields based on log content
  - Common fields include: `level`, `time`, `message`, `taskId`
  - Additional fields vary based on the specific log entry
- `count` (integer): Total number of log entries for this task

---

### GET /api/logs
Retrieves all log entries for all tasks.

**Response: 200 OK**
```json
{
  "logs": {
    "task_20250125_143022": [
      {
        "level": "info",
        "time": "2025-01-25T14:30:22Z",
        "message": "Starting DAG execution",
        "taskId": "task_20250125_143022"
      }
    ],
    "task_20250125_142015": [
      {
        "level": "error",
        "time": "2025-01-25T14:20:15Z",
        "message": "Asset execution failed",
        "taskId": "task_20250125_142015",
        "error": "Connection timeout"
      }
    ]
  },
  "taskCount": 2,
  "totalLogCount": 2
}
```

**Field Descriptions:**
- `logs` (object): Map of task IDs to their respective log entries
  - Keys are task IDs
  - Values are arrays of log entry objects
- `taskCount` (integer): Number of unique tasks with logs
- `totalLogCount` (integer): Total number of log entries across all tasks

---

### DELETE /api/logs/:taskId
Clears all log entries for a specific task.

**Parameters:**
- `taskId` (path parameter): The task ID to clear logs for

**Response: 200 OK**
```json
{
  "message": "Logs cleared for task: task_20250125_143022",
  "taskId": "task_20250125_143022"
}
```

**Response: 400 Bad Request**
```json
{
  "error": "taskId is required"
}
```

---

### DELETE /api/logs
Clears all log entries for all tasks.

**Response: 200 OK**
```json
{
  "message": "All logs cleared successfully"
}
```

**Response: 501 Not Implemented (Logger not configured)**
```json
{
  "error": "Log writer not configured"
}
```

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "Invalid request format: missing required field 'taskId'"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "details": "Database connection failed"
}
```

## Status Enumerations

### Node/Asset States
- `INITIAL` - Not yet executed
- `IN_PROGRESS` - Currently executing
- `TESTING` - Running tests
- `FAILED` - Execution failed
- `SUCCESS` - Execution successful

### DAG Execution Status
- `NOT_STARTED` - Execution not initiated
- `IN_PROGRESS` - Currently executing
- `SUCCESS` - All assets executed successfully
- `FAILED` - One or more assets failed
- `PENDING` - Execution ongoing (returned after timeout)

### Test Status
- `INITIAL` - Test not yet run
- `IN_PROGRESS` - Test executing
- `FAILED` - Test failed (returned rows)
- `SUCCESS` - Test passed (zero rows)

### Materialization Types
- `table` - Creates or replaces table
- `incremental` - Appends to existing table
- `view` - Creates or replaces view
- `custom` - Custom materialization logic
- `raw` - Raw Go function execution

## Notes

1. **Timeout Behavior**: POST endpoints (`/api/dag/run`, `/api/dag/asset/:name/execute`) return within 10 seconds. If execution is not complete, they return status `PENDING` or `IN_PROGRESS` with HTTP 202 Accepted.

2. **Task ID Format**: Recommended format is `task_YYYYMMDD_HHMMSS` or any unique string identifier.

3. **Data Persistence**: Asset execution results are stored in memory and cleared on server restart or when `/api/dag/reset` is called.

4. **DataFrame Serialization**: DataFrames are converted to JSON arrays of objects where each object represents a row.

5. **Cross-Database Support**: Assets can use different database connections as specified in their configuration.

6. **Log Storage**: When UI mode is enabled with StoringConsoleWriter (default), all structured log entries are captured in memory organized by task ID. Logs are preserved across DAG executions but cleared on server restart. The log writer extracts task IDs from either the log field `taskId` or from the execution context.