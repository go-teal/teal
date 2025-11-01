# REST API endpoint for running test queries

## Overview
Add REST API endpoint to execute individual test queries independently from DAG execution, with result caching and data retrieval capabilities.

## Implementation Details

### 1. New API Endpoint: Execute Test Query

**Endpoint:** `POST /api/tests/execute/:testName`

**Request Body:**
```json
{
  "taskId": "task_123"
}
```

**Response:**
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

**IMPORTANT - Test Success Logic:**
- ✅ Test **PASSES** (status: SUCCESS) if query returns **ZERO rows**
- ❌ Test **FAILS** (status: FAILED) if query returns **ONE OR MORE rows**

This is because test queries are written to return rows that violate constraints. If nothing is returned, the constraint is satisfied.

**Implementation Requirements:**
- Execute the SQL query from the test profile
- Count returned rows
- Determine SUCCESS/FAILED based on row count (0 = success, >0 = fail)
- Store test execution result with taskId
- Cache test results per taskId for retrieval
- Track execution duration and timestamp

### 2. New API Endpoint: Get Test Data

**Endpoint:** `GET /api/tests/data/:testName?taskId=task_123`

**Response:**
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

**Implementation Requirements:**
- Return cached test execution results for the given taskId
- Include the actual query result data (rows that failed the test)
- Return 404 if test hasn't been executed for this taskId
- Support similar to asset data endpoint structure

### 3. Storage Structure

Store test execution results similar to node execution results:

```go
type TestExecutionResult struct {
    TestName    string
    TaskID      string
    Status      string // "SUCCESS" or "FAILED"
    RowCount    int
    Data        []map[string]interface{} // Actual query results
    ErrorMsg    string
    DurationMs  int64
    ExecutedAt  time.Time
}
```

Cache structure:
```
map[taskId]map[testName]*TestExecutionResult
```

### 4. Integration Points

**Debugging Service** (`pkg/services/debugging/`):
- Add `ExecuteTest(testName, taskId)` method
- Add `GetTestData(testName, taskId)` method
- Reuse existing test execution logic from DAG
- Store results in test execution cache

**UI Server** (`pkg/ui/server.go`):
- Add handler: `handleTestExecute(c *gin.Context)`
- Add handler: `handleTestData(c *gin.Context)`
- Register routes in server setup

### 5. Success Criteria

- Can execute individual test queries via API
- Test results correctly determined (0 rows = SUCCESS, >0 rows = FAILED)
- Query results cached and retrievable per taskId
- Failed test data viewable in UI (shows which rows violated the test)
- Consistent with existing asset execution patterns
- Properly tracks execution time and timestamps

## Use Cases

1. **Manual Test Validation**: Run specific tests without full DAG execution
2. **Test Debugging**: View actual rows that cause test failures
3. **UI Integration**: Display test execution results in debug UI
4. **Incremental Testing**: Test individual data quality checks during development