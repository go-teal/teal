package services

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed dist
var uiAssets embed.FS

// UIAssetsServer serves embedded UI static assets
type UIAssetsServer struct {
	port int
	fs   http.FileSystem
}

// NewUIAssetsServer creates a new UI assets server
func NewUIAssetsServer(port int) *UIAssetsServer {
	// Extract the dist subdirectory from the embedded filesystem
	distFS, err := fs.Sub(uiAssets, "dist")
	if err != nil {
		panic(fmt.Sprintf("Failed to create sub filesystem for UI assets: %v", err))
	}

	return &UIAssetsServer{
		port: port,
		fs:   http.FS(distFS),
	}
}

// Start starts the UI assets server
func (s *UIAssetsServer) Start() error {
	mux := http.NewServeMux()

	// Serve static files
	mux.HandleFunc("/", s.serveFile)

	addr := fmt.Sprintf(":%d", s.port)

	fmt.Printf("UI assets server starting on port %d (http://localhost:%d)\n", s.port, s.port)

	// Print clickable URL for terminal
	fmt.Printf("\n✨ UI available at: \033]8;;http://localhost:%d\033\\http://localhost:%d\033]8;;\033\\\n\n", s.port, s.port)

	return http.ListenAndServe(addr, s.setCORSHeaders(mux))
}

// serveFile serves individual files with proper content types
func (s *UIAssetsServer) serveFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Root path serves index.html
	if path == "/" {
		path = "/index.html"
	}

	// Remove leading slash for filesystem access
	path = strings.TrimPrefix(path, "/")

	// Try to open the file
	file, err := s.fs.Open(path)
	if err != nil {
		// If file not found, serve index.html for SPA routing
		file, err = s.fs.Open("index.html")
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			// File not found - this is normal for SPA routing
			return
		}
		path = "index.html"
	}
	defer file.Close()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Fprintf(os.Stderr, "Failed to stat file %s: %v\n", path, err)
		return
	}

	// If it's a directory, serve index.html
	if stat.IsDir() {
		file.Close()
		file, err = s.fs.Open(filepath.Join(path, "index.html"))
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		defer file.Close()
		stat, _ = file.Stat()
		path = filepath.Join(path, "index.html")
	}

	// Set content type based on file extension
	contentType := getContentType(path)
	w.Header().Set("Content-Type", contentType)

	// Set caching headers for static assets
	if strings.HasPrefix(path, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year for assets
	} else {
		w.Header().Set("Cache-Control", "no-cache") // No cache for index.html
	}

	// Serve the file
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file.(io.ReadSeeker))
}

// setCORSHeaders adds CORS headers to allow API access
func (s *UIAssetsServer) setCORSHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getContentType returns the appropriate content type for a file
func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	contentTypes := map[string]string{
		".html":  "text/html; charset=utf-8",
		".css":   "text/css; charset=utf-8",
		".js":    "application/javascript; charset=utf-8",
		".json":  "application/json; charset=utf-8",
		".xml":   "application/xml; charset=utf-8",
		".svg":   "image/svg+xml",
		".png":   "image/png",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".gif":   "image/gif",
		".ico":   "image/x-icon",
		".woff":  "font/woff",
		".woff2": "font/woff2",
		".ttf":   "font/ttf",
		".eot":   "application/vnd.ms-fontobject",
	}

	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}

	return "application/octet-stream"
}
