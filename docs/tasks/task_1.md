✅ COMPLETED - Task 1: Rework interfaces to use TaskContext

1. ✅ Reworked interfaces:
   - Execute(ctx *TaskContext, input map[string]interface{}) (interface{}, error)
   - RunTests(ctx *TaskContext, testsMap map[string]ModelTesting) []TestResult

✅ TaskContext struct implemented with fields:
   - TaskID: string
   - TaskUUID: string (assigned as UUID in PUSH method)
   - InstanceName: string
   - InstanceUUID: string (assigned as UUID in constructor)

2. ✅ SQL asset methods updated to accept TaskContext:
   - createView(ctx *TaskContext)
   - createTable(ctx *TaskContext)
   - Execute(ctx *TaskContext, ...)
   - customQuery(ctx *TaskContext)
   - insertToTable(ctx *TaskContext)
   - getDataFrame(ctx *TaskContext)

3. ✅ Created FromTaskContext(ctx *TaskContext) function in processing package:
   - {{ TaskID }} - Returns task identifier
   - {{ TaskUUID }} - Returns unique task UUID
   - {{ InstanceName }} - Returns DAG instance name
   - {{ InstanceUUID }} - Returns DAG instance UUID

4. ✅ Updated README with:
   - New template functions documentation
   - Updated raw assets ExecutorFunc signature
   - TaskContext explanation

5. ✅ Added TaskUUID tracking in DTOs:
   - DagExecutionResponseDTO now includes taskUuid field
   - DagRunRequestDTO now includes optional taskUuid field
   - TaskSummaryDTO now includes taskUuid field
   - AssetExecuteRequestDTO already had taskUuid field
   - DebugDag stores taskId to taskUUID mapping
   - GetTaskUUID method added to retrieve UUID for a taskId