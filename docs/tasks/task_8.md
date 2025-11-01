# UI Assets Server ✅

## Overview
Static file server for serving the Teal UI frontend assets using Go's embedded filesystem.

## Implementation

### 1. Makefile Integration ✅
**Location**: `Makefile`

- Copies files from `../teal-ui/dist` to `internal/domain/services/dist` during `make` command
- Conditional copy with warning if teal-ui/dist not found
- Runs before the main build step

```makefile
# Copy teal-ui dist files to internal/domain/services/dist
if [ -d ../teal-ui/dist ]; then \
    rm -rf internal/domain/services/dist; \
    mkdir -p internal/domain/services/dist; \
    cp -r ../teal-ui/dist/* internal/domain/services/dist/; \
    echo "UI assets copied from teal-ui/dist"; \
else \
    echo "Warning: ../teal-ui/dist not found, skipping UI assets copy"; \
fi
```

### 2. UI Assets Server ✅
**Location**: `internal/domain/services/ui_assets_server.go`

**Features**:
- Uses `//go:embed dist` to embed all UI assets into the binary
- Serves index.html for root path (`/`)
- SPA routing: serves index.html for 404s (client-side routing)
- Proper content-type headers based on file extension
- Cache-Control headers:
  - `public, max-age=31536000` for `/assets/*` (1 year)
  - `no-cache` for index.html
- CORS headers for API integration
- Directory index support

**Supported Content Types**:
- HTML, CSS, JavaScript
- JSON, XML
- Images: PNG, JPG, GIF, SVG, ICO
- Fonts: WOFF, WOFF2, TTF, EOT

### 3. Port Configuration ✅
- UI API server: port specified by user (default 8080)
- UI assets server: UI port + 1 (default 8081)
- Automatically started in goroutine when UI server starts

### 4. Integration ✅
**Location**: `pkg/ui/server.go`

The UI assets server is launched automatically when the UI server starts:

```go
// Start UI assets server on port+1
uiAssetsPort := s.Port + 1
uiAssetsServer := services.NewUIAssetsServer(uiAssetsPort)

// Run UI assets server in a goroutine
go func() {
    if err := uiAssetsServer.Start(); err != nil {
        log.Error().Err(err).Int("port", uiAssetsPort).Msg("UI assets server failed")
    }
}()
```

### 5. Terminal Output ✅
Clickable URL with ANSI escape codes for terminal hyperlinks:

```
✨ UI available at: http://localhost:8081
```

(Actual output includes terminal hyperlink escape sequences for click-to-open)

## Usage

### Development
1. Build teal-ui: `cd ../teal-ui && npm run build`
2. Build teal: `make`
3. Run generated UI binary: `go run cmd/<project>-ui/<project>-ui.go`

### Access
- **API Server**: `http://localhost:8080/api/*`
- **UI Frontend**: `http://localhost:8081/`

## Files Embedded
```
dist/
├── index.html
├── favicon.ico
├── favicon.svg
└── assets/
    ├── index.js
    └── index.css
```

## Architecture Benefits
- **Single Binary**: All UI assets embedded, no external dependencies
- **Fast Serving**: Assets served from memory
- **Cache Optimization**: Long-term caching for hashed assets
- **SPA Support**: Client-side routing handled correctly
- **Production Ready**: No CORS issues, proper headers