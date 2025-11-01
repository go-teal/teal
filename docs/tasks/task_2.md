# Task 2: Connection Management via REST API

## Status: COMPLETED ✅

## Requirements

1. for debug_dag let's create the following methods:
    - Connect to databases: `core.GetInstance().ConnectAll()`
    - Disconnect: `defer core.GetInstance().Shutdown()`

2. Move all connection and disconnection API calls to these methods
3. Implement REST API for these methods on ui/server.go
4. Update API spec
5. Implement Connection ERROR message if RUN/SELECT/MUTATE are called to the disconnected graph
6. CONNECTION AND DISCONNECTION ONLY BY REST API CALLS!

## Implementation Summary

### 1. Debug DAG Methods (pkg/dags/debug_dag.go)
✅ **Completed**
- Added `isConnected bool` field to track connection status
- Implemented `Connect()` method - establishes all database connections
- Implemented `Disconnect()` method - closes all database connections
- Implemented `IsConnected()` method - returns connection status
- Removed auto-connect/disconnect from `Push()` method
- Connection lifecycle now fully managed via explicit API calls

### 2. DebuggingService Methods (pkg/services/debugging/debugging_service.go)
✅ **Completed**
- Added `Connect()` wrapper method
- Added `Disconnect()` wrapper method
- Added `GetConnectionStatus()` method returning detailed configuration
- Created DTOs: `ConnectionConfigDTO`, `ConnectionStatusResponseDTO`

### 3. REST API Endpoints (pkg/ui/server.go)
✅ **Completed**
- `POST /api/dag/connect` - Establishes database connections
- `POST /api/dag/disconnect` - Closes database connections
- `GET /api/dag/connection-status` - Returns connection status with config details

### 4. Connection Error Checks
✅ **Completed**
- Added connection check in `ExecuteDag()` (RUN operation)
- Added connection check in `MutateAsset()` (MUTATE operation)
- Added connection check in `ExecuteAssetSelect()` (SELECT operation)
- Error message: "Database connections not established. Please connect to databases first using POST /api/dag/connect"

### 5. API Documentation (docs/API Specifications.md)
✅ **Completed**
- Added "Connection Management" section
- Documented all three new endpoints with request/response examples
- Added connection requirement notes
- Updated Table of Contents

## Key Changes

**File**: `pkg/dags/debug_dag.go`
- Line ~40: Added `isConnected bool` field
- Lines 300-330: Added Connect(), Disconnect(), IsConnected() methods
- Removed ConnectAll() and Shutdown() from Push() goroutine

**File**: `pkg/services/debugging/debugging_service.go`
- Lines 1190-1243: Added Connect(), Disconnect(), GetConnectionStatus() methods
- Lines 456-474: Added connection check in ExecuteDag
- Lines 733-739: Added connection check in MutateAsset
- Lines 1077-1083: Added connection check in ExecuteAssetSelect

**File**: `pkg/services/debugging/entities.go`
- Lines 165-181: Added ConnectionConfigDTO and ConnectionStatusResponseDTO

**File**: `pkg/ui/server.go`
- Lines 78-80: Added three connection API routes
- Lines 457-492: Added handleConnect, handleDisconnect, handleConnectionStatus handlers

**File**: `docs/API Specifications.md`
- Lines 17-20: Updated Table of Contents
- Lines 447-557: Added Connection Management section
- Line 1071: Added connection requirement note

## Testing Checklist

- [ ] Test POST /api/dag/connect returns 200 OK
- [ ] Test GET /api/dag/connection-status shows isConnected: true after connect
- [ ] Test POST /api/dag/run without connect returns connection error
- [ ] Test POST /api/dag/asset/:name/mutate without connect returns error
- [ ] Test POST /api/dag/asset/:name/select without connect returns error
- [ ] Test POST /api/dag/disconnect returns 200 OK
- [ ] Test GET /api/dag/connection-status shows isConnected: false after disconnect
- [ ] Test reconnect after disconnect works properly
