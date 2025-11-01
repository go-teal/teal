# Task 5: Serving docs/README.md ✅ COMPLETED

## Requirements
1. Add an endpoint to return README.md from the docs of the generated project
2. The path to this file is passed from the server service initialization from main.go
3. Update API specification
4. Update this task

## Implementation Summary

### Changes Made

**File 1:** `pkg/ui/server.go`

**Imports:**
- Added `os` import for file reading (line 6)

**UIServer Struct (lines 17-23):**
- Added `readmePath string` field to store the path to README.md

**New Constructor (lines 44-53):**
- Added `NewUIServerWithLogWriterAndReadme()` constructor
- Accepts `readmePath` parameter along with existing parameters
- Initializes all UIServer fields including the new readmePath

**Route Registration (line 96):**
- Added `GET /api/docs/readme` route mapped to `handleReadme` handler

**Handler Implementation (lines 512-534):**
- `handleReadme()` function reads and returns README.md content
- Checks if readmePath is configured (returns 404 if not)
- Reads file using `os.ReadFile()`
- Returns content with `text/markdown; charset=utf-8` content type
- Handles errors with appropriate HTTP status codes and logging

---

**File 2:** `internal/domain/generators/templates/main_ui.go.tmpl`

**Updated Server Initialization (lines 77-79):**
- Sets `readmePath` to `"./docs/README.md"` relative to project root
- Uses new `NewUIServerWithLogWriterAndReadme()` constructor
- Passes readme path during server initialization

---

**File 3:** `docs/API Specifications.md`

**Table of Contents (lines 27-28):**
- Added "Documentation" section with GET /api/docs/readme endpoint

**New Documentation Section (lines 889-925):**
- Complete API specification for the readme endpoint
- Response examples with markdown content type
- Error responses (404 Not Found, 500 Internal Server Error)
- Usage notes and configuration details

## How It Works

1. **During Generation:**
   - `teal gen` creates UI binary with hardcoded path `./docs/README.md`
   - Path is passed to UIServer constructor

2. **At Runtime:**
   - UI server stores readme path in struct field
   - When `GET /api/docs/readme` is called:
     - Validates path is configured
     - Reads file from filesystem
     - Returns raw markdown content

3. **Error Handling:**
   - Returns 404 if path not configured (though always set in template)
   - Returns 500 if file read fails (file missing/permission error)
   - Logs errors with details for debugging

## Benefits

✅ **Documentation Access** - UI tools can display project documentation
✅ **Simple Integration** - Markdown content ready for rendering
✅ **Configurable Path** - Path passed during initialization
✅ **Proper Error Handling** - Clear error responses with details
✅ **Standard Content Type** - Uses `text/markdown` MIME type

## API Endpoint

**Endpoint:** `GET /api/docs/readme`

**Success Response:**
```
HTTP/1.1 200 OK
Content-Type: text/markdown; charset=utf-8

# Project Documentation
...markdown content...
```

**Usage Example:**
```bash
curl http://localhost:8080/api/docs/readme
```