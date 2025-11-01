# Static assets service and change observer

## Overview
Hot-reload development server for Teal UI that watches for file changes and automatically regenerates code.

## Implementation Details

### 1. New CLI Command
- Command: `teal ui --port=8080 --log-level=debug`
- Flags:
  - `--port` (default: 8080) - Port for the UI server
  - `--log-level` (default: debug) - Log level for both watcher and UI process

### 2. File Watcher Service
Location: `internal/domain/services/asset_observer.go`

**Watched Files:**
- All files in `assets/` folders and subfolders
- `profile.yaml`
- `config.yaml`

**Behavior:**
- Calculate MD5 hashes for each watched file
- Store in `map[string]string` (file path -> hash)
- Check file hashes every 2 seconds
- Implement debouncing: wait 500ms after last detected change before triggering reload
  - This prevents multiple rapid regenerations when several files change at once

### 3. Child Process Management
**Initial Startup:**
- Execute: `go run ./cmd/<project-name>-ui/<project-name>-ui.go --port=<port> --log-level=<log-level>`
- Pass the port and log level flags to the child process
- Capture and log child process output in real-time
- Track child process PID

**On File Change Detection:**
1. Send **SIGTERM** signal to child process (allows graceful shutdown)
2. Wait for child process to exit
3. Trigger code regeneration via existing `gen.go` logic
4. If regeneration succeeds: restart child UI process with same port and log level
5. If regeneration fails: do NOT restart (old process already killed)
6. Resume hash checking loop

**Rationale for SIGTERM:**
- Allows UI server to close connections gracefully
- Properly cleanup resources
- Better than SIGKILL which terminates immediately

### 4. Graceful Shutdown
**Signal Handling:**
- Handle OS interrupt signals (SIGINT/SIGTERM from CTRL+C)
- Send SIGTERM to child UI process
- Wait for child to exit cleanly (with timeout)
- Clean up file watcher resources
- Exit main process

## Success Criteria
- File changes in assets/, profile.yaml, or config.yaml trigger reload
- Debouncing prevents excessive regenerations
- UI server restarts automatically on same port after successful regeneration
- Port and log level are correctly passed to child process
- UI server stays down after failed regeneration (clear error message)
- CTRL+C cleanly shuts down both watcher and UI processes