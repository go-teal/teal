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
- [Connection Management](#connection-management)
  - [POST /api/dag/connect](#post-apidagconnect)
  - [POST /api/dag/disconnect](#post-apidagdisconnect)
  - [GET /api/dag/connection-status](#get-apidagconnection-status)
- [Asset Operations](#asset-operations)
  - [POST /api/dag/asset/:name/mutate](#post-apidagassetnamemutate)
  - [GET /api/dag/asset/:name/data](#get-apidagassetnamedata)
  - [POST /api/dag/asset/:name/select](#post-apidagassetnameselect)
- [Test Operations](#test-operations)
  - [GET /api/tests](#get-apitests)
  - [GET /api/tests/results/:taskId](#get-apitestsresultstaskid)
  - [POST /api/tests/execute/:testName](#post-apitestsexecutetestname)
  - [GET /api/tests/data/:testName](#get-apitestsdatatestname)
- [Documentation](#documentation)
  - [GET /api/docs/readme](#get-apidocsreadme)
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
      "lastTestsDuration": 0,
      "taskGroupIndex": 0
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
  - `state` (string): Current execution state - "INITIAL", "IN_PROGRESS", "TESTING", "FAILED", "SUCCESS", "TESTS_FAILED"
  - `totalTests` (integer): Total number of tests for this node
  - `successfulTests` (integer): Number of tests that passed
  - `lastExecutionDuration` (integer): Last execution time in milliseconds
  - `lastTestsDuration` (integer): Last test execution time in milliseconds
  - `taskGroupIndex` (integer): Index of the task group (execution stage) in the DAG (0-based)

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
  "rootTestResults": [
    {
      "testName": "root.test_data_integrity",
      "status": "SUCCESS",
      "errorMsg": "",
      "durationMs": 125
    },
    {
      "testName": "root.test_final_validation",
      "status": "SUCCESS",
      "errorMsg": "",
      "durationMs": 89
    }
  ],
  "tasks": [
    {
      "name": "staging.hello",
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
      "name": "dds.world",
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
  "lastTaskName": "dds.world"
}
```

**Response: 200 OK (Tests Failed)**
```json
{
  "taskId": "task_20250125_143023",
  "status": "TESTS_FAILED",
  "completedAssets": 2,
  "totalAssets": 2,
  "failedAssets": 0,
  "inProgressAssets": 0,
  "tasks": [
    {
      "name": "staging.hello",
      "state": "TESTS_FAILED",
      "order": 1,
      "startTime": 1737816622000,
      "endTime": 1737816625500,
      "executionTimeMs": 1500,
      "testExecutionTimeMs": 2000,
      "totalTests": 2,
      "passedTests": 1,
      "failedTests": 1,
      "message": "",
      "testResults": [
        {
          "testName": "test_hello_not_null",
          "status": "SUCCESS",
          "durationMs": 500
        },
        {
          "testName": "test_hello_valid",
          "status": "FAILED",
          "error": "Test failed: found 3 rows",
          "durationMs": 1500
        }
      ]
    },
    {
      "name": "dds.world",
      "state": "SUCCESS",
      "order": 2,
      "startTime": 1737816625600,
      "endTime": 1737816627000,
      "executionTimeMs": 1400,
      "message": ""
    }
  ],
  "lastTaskName": "dds.world"
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
      "name": "staging.hello",
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
      "name": "dds.world",
      "state": "IN_PROGRESS",
      "order": 2,
      "startTime": 1737816623500,
      "totalTests": 2,
      "passedTests": 0,
      "failedTests": 0,
      "testResults": []
    }
  ],
  "lastTaskName": "dds.world"
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
      "name": "staging.hello",
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
  "lastTaskName": "staging.hello"
}
```

**Field Descriptions:**
- `taskId` (string, required): Unique identifier for this execution
- `data` (object, optional): Input data to pass to the DAG
- `status` (string): Overall execution status - "NOT_STARTED", "IN_PROGRESS", "SUCCESS", "FAILED", "PENDING", "TESTS_FAILED"
- `completedAssets` (integer): Total number of completed assets across all tasks
- `totalAssets` (integer): Total number of assets in the DAG
- `failedAssets` (integer): Total number of failed assets
- `inProgressAssets` (integer): Total number of assets currently executing
- `rootTestResults` (array, optional): Array of root test execution results (tests with "root." prefix executed after DAG completion)
  - `testName` (string): Name of the root test
  - `status` (string): Test status - "SUCCESS", "FAILED", "NOT_FOUND"
  - `errorMsg` (string): Error message if test failed
  - `durationMs` (integer): Test execution duration in milliseconds
- `nodes` (array): Array of node execution status objects
  - `name` (string): Node/asset name
  - `state` (string): Node state - "INITIAL", "IN_PROGRESS", "TESTING", "FAILED", "SUCCESS", "TESTS_FAILED"
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
  "rootTestResults": [
    {
      "testName": "root.test_data_integrity",
      "status": "SUCCESS",
      "errorMsg": "",
      "durationMs": 125
    }
  ],
  "nodes": [
    {
      "name": "staging.hello",
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
  "lastTaskName": "staging.hello"
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

## Connection Management

Database connections must be explicitly established before executing DAG operations (RUN, SELECT, MUTATE). These endpoints manage the connection lifecycle.

### POST /api/dag/connect
Establishes connections to all configured databases.

**Response: 200 OK**
```json
{
  "message": "Successfully connected to all databases",
  "status": "CONNECTED"
}
```

**Response: 500 Internal Server Error**
```json
{
  "error": "Failed to connect to databases",
  "details": "Connection refused for database 'postgres_prod' at localhost:5432"
}
```

**Field Descriptions:**
- `message` (string): Success message
- `status` (string): Connection status ("CONNECTED")
- `error` (string): Error message if connection failed
- `details` (string): Detailed error information

**Notes:**
- Must be called before any DAG execution or asset operation
- Connects to all databases defined in `config.yaml`
- Connection state is shared across all API calls

---

### POST /api/dag/disconnect
Closes all database connections gracefully.

**Response: 200 OK**
```json
{
  "message": "Successfully disconnected from all databases",
  "status": "DISCONNECTED"
}
```

**Response: 500 Internal Server Error**
```json
{
  "error": "Failed to disconnect from databases",
  "details": "Error closing connection to 'duckdb_analytics'"
}
```

**Field Descriptions:**
- `message` (string): Success message
- `status` (string): Connection status ("DISCONNECTED")
- `error` (string): Error message if disconnection failed
- `details` (string): Detailed error information

**Notes:**
- Safely closes all active database connections
- Should be called when shutting down or when connections are no longer needed
- Attempting DAG operations after disconnect will result in connection errors

---

### GET /api/dag/connection-status
Retrieves the current connection status and configuration details for all databases.

**Response: 200 OK**
```json
{
  "isConnected": true,
  "connections": [
    {
      "name": "memory_duck",
      "type": "duckdb",
      "path": ":memory:",
      "extensions": ["parquet", "json"]
    },
    {
      "name": "postgres_prod",
      "type": "postgres",
      "host": "localhost",
      "port": 5432,
      "database": "production",
      "user": "etl_user"
    }
  ]
}
```

**Field Descriptions:**
- `isConnected` (boolean): Whether the DAG is currently connected to databases
- `connections` (array): Array of connection configuration objects
  - `name` (string): Connection identifier from config.yaml
  - `type` (string): Database type ("duckdb", "postgres", etc.)
  - `host` (string, optional): Database host (for network databases)
  - `port` (integer, optional): Database port
  - `database` (string, optional): Database name
  - `user` (string, optional): Database user
  - `path` (string, optional): File path (for file-based databases)
  - `extensions` (array, optional): DuckDB extensions to load

**Notes:**
- Returns configuration details without sensitive information (passwords are excluded)
- `isConnected` reflects the overall connection state, not individual connection health
- Sensitive fields (passwords, certificates) are intentionally omitted from the response

---

## Asset Operations

### POST /api/dag/asset/:name/mutate
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
- `status` (string): Execution status - "INITIAL", "IN_PROGRESS", "TESTING", "FAILED", "SUCCESS", "TESTS_FAILED"
- `startTime` (integer, optional): Unix timestamp in milliseconds
- `endTime` (integer, optional): Unix timestamp in milliseconds
- `executionTimeMs` (integer): Execution duration in milliseconds
- `result` (object, optional): Execution result data
- `error` (string): Error message if failed
- `upstreamsUsed` (array): List of upstream assets used

---

### GET /api/dag/asset/:name/data
Retrieves the execution result and metadata for an asset from a specific task execution with optional pagination.

**Parameters:**
- `name` (path parameter): Asset name (e.g., "staging.hello")

**Query Parameters:**
- `taskId` (required): Task ID from DAG execution
- `offset` (optional): Starting index for pagination (0-based, default: 0)
- `limit` (optional): Maximum number of records to return (default: 0 = no limit)

**Response: 200 OK (DataFrame Result)**
```json
{
  "taskId": "task_20250125_143022",
  "assetName": "staging.hello",
  "status": "SUCCESS",
  "startTime": 1737816622000,
  "endTime": 1737816623500,
  "executionTimeMs": 1500,
  "result": [
    {"greeting": "Hello", "id": 1},
    {"greeting": "World", "id": 2}
  ],
  "upstreamsUsed": [],
  "totalRecords": 100,
  "offset": 0,
  "limit": 2
}
```

**Response: 200 OK (Map Result)**
```json
{
  "taskId": "task_20250125_143022",
  "assetName": "dds.world",
  "status": "SUCCESS",
  "startTime": 1737816623500,
  "endTime": 1737816625000,
  "executionTimeMs": 1500,
  "result": {
    "key1": "value1",
    "key2": 42
  },
  "upstreamsUsed": ["staging.hello"]
}
```

**Response: 202 Accepted (In Progress)**
```json
{
  "taskId": "task_20250125_143022",
  "assetName": "staging.hello",
  "status": "IN_PROGRESS",
  "startTime": 1737816622000,
  "error": "Execution timeout (still processing in background)",
  "upstreamsUsed": []
}
```

**Response: 200 OK (Not Executed)**
```json
{
  "taskId": "task_20250125_143022",
  "assetName": "staging.hello",
  "status": "INITIAL",
  "upstreamsUsed": []
}
```

**Response: 200 OK (Failed)**
```json
{
  "taskId": "task_20250125_143022",
  "assetName": "staging.hello",
  "status": "FAILED",
  "startTime": 1737816622000,
  "endTime": 1737816623500,
  "executionTimeMs": 1500,
  "error": "SQL execution failed: table not found",
  "upstreamsUsed": []
}
```

**Field Descriptions:**
- `taskId` (string): Task ID for this execution
- `assetName` (string): Name of the asset
- `status` (string): Execution status - "INITIAL", "IN_PROGRESS", "SUCCESS", "FAILED", "TESTS_FAILED"
- `startTime` (integer, optional): Unix timestamp in milliseconds
- `endTime` (integer, optional): Unix timestamp in milliseconds
- `executionTimeMs` (integer): Execution duration in milliseconds
- `result` (any, optional): The execution result. DataFrames are automatically serialized as arrays of objects. Paginated based on offset and limit
- `error` (string, optional): Error message if execution failed
- `upstreamsUsed` (array): List of upstream assets used in execution
- `totalRecords` (integer, optional): Total number of records available (before pagination)
- `offset` (integer, optional): Starting index used for this response
- `limit` (integer, optional): Maximum records returned in this response (0 = no limit)

**Notes:**
- DataFrame results are automatically serialized to JSON as arrays of objects, where each object represents a row
- Pagination is applied after serialization: `offset` specifies the starting record index (0-based), and `limit` specifies the maximum number of records to return
- `totalRecords` shows the complete dataset size before pagination, allowing clients to calculate total pages
- **Important**: The `result` field contains data from the node's `LastResult`, which is overwritten on each new execution. If Task B executes after Task A, querying `?taskId=taskA` will return Task B's result, not Task A's original result. The endpoint only guarantees metadata (status, timing, error) is task-specific; the actual data reflects the most recent execution of that asset
- This endpoint returns execution metadata from a specific task, unlike `/select` which executes a fresh query

**Pagination Examples:**
- `GET /api/dag/asset/staging.hello/data?taskId=task_123` - Returns all records
- `GET /api/dag/asset/staging.hello/data?taskId=task_123&limit=10` - Returns first 10 records
- `GET /api/dag/asset/staging.hello/data?taskId=task_123&offset=10&limit=10` - Returns records 11-20
- `GET /api/dag/asset/staging.hello/data?taskId=task_123&offset=100` - Returns all records starting from index 100

---

### POST /api/dag/asset/:name/select
Executes the asset's SQL query using the `ToDataFrame` method, renders the SQL template (executing all template functions like `Ref` and `IsIncremental`), and saves the result to the node's `LastResult`. Returns within 10 seconds. The result can then be retrieved using the `/api/dag/asset/:name/data` endpoint.

**Parameters:**
- `name` (path parameter): Asset name (e.g., "staging.hello")

**Request Body:**
```json
{
  "taskId": "task_select_20250115"
}
```

**Field Descriptions (Request):**
- `taskId` (string, required): Task ID to associate with this execution

**Response: 200 OK (Success)**
```json
{
  "taskId": "task_select_20250115",
  "assetName": "staging.hello",
  "status": "SUCCESS",
  "startTime": 1737816622000,
  "endTime": 1737816623500,
  "executionTimeMs": 1500
}
```

**Response: 202 Accepted (Still Running)**
```json
{
  "taskId": "task_select_20250115",
  "assetName": "staging.hello",
  "status": "IN_PROGRESS",
  "startTime": 1737816622000,
  "error": "Execution timeout (still processing in background)"
}
```

**Response: 200 OK (Failed)**
```json
{
  "taskId": "task_select_20250115",
  "assetName": "staging.hello",
  "status": "FAILED",
  "error": "Failed to execute query: syntax error at line 5"
}
```

**Field Descriptions (Response):**
- `taskId` (string): Task ID for this execution
- `assetName` (string): Name of the asset
- `status` (string): Execution status - "SUCCESS", "IN_PROGRESS", or "FAILED"
- `startTime` (integer, optional): Unix timestamp in milliseconds when execution started
- `endTime` (integer, optional): Unix timestamp in milliseconds when execution completed
- `executionTimeMs` (integer, optional): Execution duration in milliseconds
- `error` (string, optional): Error message if query failed or timeout message if still processing

**Notes:**
- This endpoint only works with SQL assets (not raw assets)
- **Template Rendering**: The SQL template is rendered before execution, which means:
  - `{{ Ref "stage.model" }}` functions are executed to resolve table references
  - `{{ IsIncremental() }}` returns `true` if materialization is incremental and the table exists
  - All other template functions are evaluated
- The result is saved to the node's `LastResult` and can be retrieved via `GET /api/dag/asset/:name/data?taskId=<taskId>`
- To get paginated data, use `GET /api/dag/asset/:name/data?taskId=<taskId>&offset=0&limit=100`
- The result is stored against the taskId in execution metadata
- Uses the asset's configured database connection
- **Important**: The data in `LastResult` is overwritten on subsequent executions (see `/data` endpoint notes)

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
      "connectionType": "duckdb"
    },
    {
      "name": "test_world_count",
      "sql": "SELECT * FROM dds.world WHERE count < 0",
      "connectionName": "memory_duck",
      "connectionType": "duckdb"
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

---

### GET /api/tests/results/:taskId
Retrieves test execution results for a specific task execution.

**Parameters:**
- `taskId` (path parameter): The task ID from DAG execution

**Response: 200 OK**
```json
{
  "taskId": "task_20250125_143023",
  "tests": [
    {
      "name": "test_hello_exists",
      "sql": "SELECT * FROM staging.hello WHERE greeting IS NULL",
      "connectionName": "memory_duck",
      "connectionType": "duckdb"
    },
    {
      "name": "test_world_count",
      "sql": "SELECT * FROM dds.world WHERE count < 0",
      "connectionName": "memory_duck",
      "connectionType": "duckdb"
    }
  ]
}
```

**Field Descriptions:**
- `taskId` (string): The task ID for this test execution
- `tests` (array): Array of test profiles for this specific task execution
  - `name` (string): Unique test identifier
  - `sql` (string): SQL query that was executed
  - `connectionName` (string): Database connection used
  - `connectionType` (string): Type of database connection

---

### POST /api/tests/execute/:testName
Executes a specific test independently, outside of DAG execution. Returns execution status and metrics.

**Parameters:**
- `testName` (path parameter): Name of the test to execute (e.g., "dds.test_dim_airports_unique")

**Request Body:**
```json
{
  "taskId": "manual_test_20250125_143022"
}
```

**Response: 200 OK (Test Passed)**
```json
{
  "testName": "dds.test_dim_airports_unique",
  "description": "## 🔍 Airport Dimension Uniqueness Test\n\n**Test Type**: Data Quality - Primary Key Constraint...",
  "taskId": "manual_test_20250125_143022",
  "status": "SUCCESS",
  "rowCount": 0,
  "errorMsg": "",
  "durationMs": 45,
  "executedAt": "2025-01-25T14:30:22Z"
}
```

**Response: 200 OK (Test Failed)**
```json
{
  "testName": "dds.test_dim_airports_unique",
  "description": "## 🔍 Airport Dimension Uniqueness Test\n\n**Test Type**: Data Quality - Primary Key Constraint...",
  "taskId": "manual_test_20250125_143022",
  "status": "FAILED",
  "rowCount": 3,
  "errorMsg": "",
  "durationMs": 67,
  "executedAt": "2025-01-25T14:30:22Z"
}
```

**Response: 500 Internal Server Error**
```json
{
  "testName": "dds.test_dim_airports_unique",
  "description": "",
  "taskId": "manual_test_20250125_143022",
  "status": "FAILED",
  "rowCount": 0,
  "errorMsg": "Failed to execute query: syntax error at line 5",
  "durationMs": 12,
  "executedAt": "2025-01-25T14:30:22Z"
}
```

**Field Descriptions:**
- `testName` (string): Name of the test executed
- `description` (string, optional): Markdown description from test profile (base64 decoded)
- `taskId` (string): Task ID associated with this execution
- `status` (string): Test result - "SUCCESS" (0 rows) or "FAILED" (>0 rows or error)
- `rowCount` (integer): Number of violation rows returned (0 = pass, >0 = fail)
- `errorMsg` (string, optional): Error message if test execution failed
- `durationMs` (integer): Test execution duration in milliseconds
- `executedAt` (string): ISO 8601 timestamp of execution

**Notes:**
- Test passes if it returns **zero rows** (no violations found)
- Test fails if it returns **one or more rows** (violations exist)
- The test SQL is automatically wrapped with `SELECT COUNT(*) FROM (...) HAVING count > 0 LIMIT 1` for execution
- Database connection must be established first using `POST /api/dag/connect`
- **TaskUUID Generation**: A unique UUID is automatically generated and associated with the taskId if one doesn't already exist
- **Data Storage**: Test results (including violation DataFrames) are stored in `DebugDag.TestExecutionMap` keyed by `taskId -> testName`
- **Logs**: All structured logs are captured with consistent field names:
  - `taskId` - The task identifier
  - `taskUUID` - The generated UUID for this task
  - `testName` - The test name (e.g., "dds.test_dim_airports_unique")
  - Logs can be retrieved via `GET /api/logs/:taskId`

---

### GET /api/tests/data/:testName
Retrieves the detailed violation data from a test execution, showing the actual rows that failed the test.

**Parameters:**
- `testName` (path parameter): Name of the test (e.g., "dds.test_dim_airports_unique")

**Query Parameters:**
- `taskId` (required): Task ID from test execution

**Response: 200 OK (Test Passed)**
```json
{
  "testName": "dds.test_dim_airports_unique",
  "description": "## 🔍 Airport Dimension Uniqueness Test\n\n**Test Type**: Data Quality - Primary Key Constraint...",
  "taskId": "manual_test_20250125_143022",
  "status": "SUCCESS",
  "rowCount": 0,
  "data": [],
  "executedAt": "2025-01-25T14:30:22Z"
}
```

**Response: 200 OK (Test Failed with Violations)**
```json
{
  "testName": "dds.test_dim_airports_unique",
  "description": "## 🔍 Airport Dimension Uniqueness Test\n\n**Test Type**: Data Quality - Primary Key Constraint...",
  "taskId": "manual_test_20250125_143022",
  "status": "FAILED",
  "rowCount": 3,
  "data": [
    {
      "airport_key": "abc123def456",
      "duplicate_count": 2
    },
    {
      "airport_key": "xyz789ghi012",
      "duplicate_count": 3
    },
    {
      "airport_key": "mno345pqr678",
      "duplicate_count": 2
    }
  ],
  "executedAt": "2025-01-25T14:30:22Z"
}
```

**Response: 404 Not Found**
```json
{
  "testName": "unknown_test",
  "description": "",
  "taskId": "manual_test_20250125_143022",
  "status": "INITIAL",
  "rowCount": 0,
  "data": [],
  "executedAt": ""
}
```

**Field Descriptions:**
- `testName` (string): Name of the test
- `description` (string, optional): Markdown description from test profile (base64 decoded)
- `taskId` (string): Task ID for this test execution
- `status` (string): Test status - "SUCCESS", "FAILED", "INITIAL"
- `rowCount` (integer): Number of violation rows returned
- `data` (array): Array of violation records (empty if test passed)
  - Structure depends on the test SQL SELECT columns
  - Each object represents one row that violated the test constraint
- `executedAt` (string): ISO 8601 timestamp of execution

**Notes:**
- This endpoint returns the **raw violation data** from the test SQL query
- Unlike the COUNT-wrapped query used for pass/fail determination, this returns actual rows
- The `data` array is **always present** (empty array when test passes, populated when test fails)
- Useful for debugging and understanding exactly what data violated the test constraint
- The test must be executed first using `POST /api/tests/execute/:testName`
- **Data Retrieval**: Violation data is retrieved from `DebugDag.TestExecutionMap[taskId][testName].DataFrame`
- Database connection must be established first using `POST /api/dag/connect`
- Each record in the `data` array corresponds to one row returned by the test query showing constraint violations

---

## Documentation

### GET /api/docs/readme
Retrieves the project README documentation in markdown format.

**Response: 200 OK**
```markdown
# Project Name

## Overview
...project documentation content...
```

**Response Headers:**
- `Content-Type`: `text/markdown; charset=utf-8`

**Error Response: 404 Not Found**
```json
{
  "error": "README path not configured"
}
```

**Error Response: 500 Internal Server Error**
```json
{
  "error": "Failed to read README file",
  "details": "open ./docs/README.md: no such file or directory"
}
```

**Notes:**
- The README path is configured during server initialization (default: `./docs/README.md`)
- Returns raw markdown content that can be rendered by markdown viewers
- Useful for displaying project documentation directly in UI tools

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
      "message": "Executing SQL select query",
      "taskId": "task_20250125_143022",
      "taskUUID": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "assetName": "staging.hello",
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
- `FAILED` - Asset execution failed
- `SUCCESS` - Execution successful (all tests passed)
- `TESTS_FAILED` - Asset executed successfully but one or more tests failed

### DAG Execution Status
- `NOT_STARTED` - Execution not initiated
- `IN_PROGRESS` - Currently executing
- `SUCCESS` - All assets and tests executed successfully
- `FAILED` - One or more assets failed execution
- `PENDING` - Execution ongoing (returned after timeout)
- `TESTS_FAILED` - All assets executed successfully but one or more tests failed

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

1. **Connection Requirement**: Database connections must be established using `POST /api/dag/connect` before executing any DAG operations (`/api/dag/run`, `/api/dag/asset/:name/mutate`, `/api/dag/asset/:name/select`). Attempting to execute operations without an active connection will return a FAILED status with an error message: "Database connections not established. Please connect to databases first using POST /api/dag/connect".

2. **Timeout Behavior**: POST endpoints (`/api/dag/run`, `/api/dag/asset/:name/mutate`, `/api/dag/asset/:name/select`) return within 10 seconds. If execution is not complete, they return status `PENDING` or `IN_PROGRESS` with HTTP 202 Accepted. The execution continues in the background and results can be retrieved using `/api/dag/asset/:name/data` or `/api/dag/status/:taskId`.

3. **Task ID Format**: Recommended format is `task_YYYYMMDD_HHMMSS` or any unique string identifier.

4. **Data Persistence**: Asset execution results are stored in memory and cleared on server restart or when `/api/dag/reset` is called.

5. **DataFrame Serialization**: DataFrames are converted to JSON arrays of objects where each object represents a row.

6. **Cross-Database Support**: Assets can use different database connections as specified in their configuration.

7. **Log Storage**: When UI mode is enabled with StoringConsoleWriter (default), all structured log entries are captured in memory organized by task ID. Logs are preserved across DAG executions but cleared on server restart. The log writer extracts task IDs from either the log field `taskId` or from the execution context.

8. **Background Execution**: When POST endpoints return 202 Accepted due to timeout, the execution continues in the background. The mutex lock is only held briefly for state updates, not during long-running operations like SQL queries, ensuring the server remains responsive.

9. **Connection Lifecycle**: Connections are managed separately from DAG execution state. Disconnecting (`POST /api/dag/disconnect`) does not reset DAG state or clear execution history. Use `POST /api/dag/reset` to clear execution data.

10. **TaskUUID Generation**: All individual operations (test execution, asset select, asset mutate) automatically generate a unique UUID associated with the taskId if one doesn't already exist. The mapping is stored in `DebugDag.TaskUUIDMap[taskId]` and persists for the lifetime of the server session. This ensures every task has a unique identifier for tracing and correlation.

11. **Test Data Storage**: Test execution results, including violation DataFrames, are stored in `DebugDag.TestExecutionMap` with the structure `taskId -> testName -> TestExecutionResult`. This allows retrieval of test violation data via `GET /api/tests/data/:testName?taskId=<taskId>`. Test data is separate from asset node data and persists until server restart or DAG reset.

12. **Logging Field Standardization**: All structured logs use consistent field names for better querying and filtering:
    - `assetName` - Name of the asset being executed (not `asset`)
    - `testName` - Name of the test being executed (not `test`)
    - `taskId` - Task identifier
    - `taskUUID` - Unique UUID for the task (auto-generated)
    - Example: `{"level":"debug", "taskId":"task_001", "taskUUID":"abc-123", "assetName":"staging.hello", "message":"Executing SQL select query"}`