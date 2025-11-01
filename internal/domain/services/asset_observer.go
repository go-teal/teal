package services

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// ApplicationInterface defines the methods needed from application
type ApplicationInterface interface {
	GenegateAssets(projectPath string, configFilePath string, models string) error
}

// AssetObserver watches files for changes and manages UI process lifecycle
type AssetObserver struct {
	projectPath string
	port        int
	logLevel    string
	projectName string
	app         ApplicationInterface

	// File tracking
	fileHashes map[string]string
	hashMutex  sync.RWMutex

	// Child process management
	uiProcess     *exec.Cmd
	processMutex  sync.Mutex
	processActive bool

	// Debouncing
	lastChangeTime time.Time
	changeMutex    sync.Mutex
	debounceDelay  time.Duration

	// Shutdown handling
	shutdown     chan bool
	done         chan bool
	shuttingDown bool
}

// NewAssetObserver creates a new asset observer
func NewAssetObserver(projectPath string, port int, logLevel string, projectName string, app ApplicationInterface) *AssetObserver {
	return &AssetObserver{
		projectPath:   projectPath,
		port:          port,
		logLevel:      logLevel,
		projectName:   projectName,
		app:           app,
		fileHashes:    make(map[string]string),
		debounceDelay: 500 * time.Millisecond,
		shutdown:      make(chan bool),
		done:          make(chan bool),
	}
}

// calculateFileHash computes MD5 hash of a file
func (ao *AssetObserver) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// scanFiles recursively scans and calculates hashes for all watched files
func (ao *AssetObserver) scanFiles() error {
	newHashes := make(map[string]string)

	// Watch assets directory
	assetsPath := filepath.Join(ao.projectPath, "assets")
	if _, err := os.Stat(assetsPath); err == nil {
		err := filepath.Walk(assetsPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				hash, err := ao.calculateFileHash(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to calculate hash for %s: %v\n", path, err)
					return nil
				}
				newHashes[path] = hash
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Watch profile.yaml
	profilePath := filepath.Join(ao.projectPath, "profile.yaml")
	if _, err := os.Stat(profilePath); err == nil {
		hash, err := ao.calculateFileHash(profilePath)
		if err == nil {
			newHashes[profilePath] = hash
		}
	}

	// Watch config.yaml
	configPath := filepath.Join(ao.projectPath, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		hash, err := ao.calculateFileHash(configPath)
		if err == nil {
			newHashes[configPath] = hash
		}
	}

	ao.hashMutex.Lock()
	ao.fileHashes = newHashes
	ao.hashMutex.Unlock()

	return nil
}

// detectChanges checks for file changes and returns true if any detected
func (ao *AssetObserver) detectChanges() (bool, []string) {
	newHashes := make(map[string]string)
	changedFiles := []string{}

	// Scan assets directory
	assetsPath := filepath.Join(ao.projectPath, "assets")
	if _, err := os.Stat(assetsPath); err == nil {
		filepath.Walk(assetsPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			hash, err := ao.calculateFileHash(path)
			if err == nil {
				newHashes[path] = hash
			}
			return nil
		})
	}

	// Check profile.yaml
	profilePath := filepath.Join(ao.projectPath, "profile.yaml")
	if _, err := os.Stat(profilePath); err == nil {
		hash, err := ao.calculateFileHash(profilePath)
		if err == nil {
			newHashes[profilePath] = hash
		}
	}

	// Check config.yaml
	configPath := filepath.Join(ao.projectPath, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		hash, err := ao.calculateFileHash(configPath)
		if err == nil {
			newHashes[configPath] = hash
		}
	}

	// Compare hashes
	ao.hashMutex.RLock()
	oldHashes := ao.fileHashes
	ao.hashMutex.RUnlock()

	hasChanges := false

	// Check for modified files and NEW files
	for path, newHash := range newHashes {
		if oldHash, exists := oldHashes[path]; exists {
			// Existing file - check if modified
			if oldHash != newHash {
				hasChanges = true
				changedFiles = append(changedFiles, path+" (modified)")
			}
		} else {
			// New file added
			hasChanges = true
			changedFiles = append(changedFiles, path+" (new)")
		}
	}

	// Check for deleted files
	for path := range oldHashes {
		if _, exists := newHashes[path]; !exists {
			hasChanges = true
			changedFiles = append(changedFiles, path+" (deleted)")
		}
	}

	if hasChanges {
		ao.hashMutex.Lock()
		ao.fileHashes = newHashes
		ao.hashMutex.Unlock()
	}

	return hasChanges, changedFiles
}

// startUIProcess starts the UI child process
func (ao *AssetObserver) startUIProcess() error {
	ao.processMutex.Lock()
	defer ao.processMutex.Unlock()

	if ao.processActive {
		return fmt.Errorf("UI process already running")
	}

	uiPath := filepath.Join(ao.projectPath, "cmd", ao.projectName+"-ui", ao.projectName+"-ui.go")

	// Check if UI file exists
	if _, err := os.Stat(uiPath); os.IsNotExist(err) {
		return fmt.Errorf("UI file not found: %s", uiPath)
	}

	fmt.Printf("Starting UI process: %s (port=%d, log_level=%s)\n", uiPath, ao.port, ao.logLevel)

	cmd := exec.Command("go", "run", uiPath,
		"--port", fmt.Sprintf("%d", ao.port),
		"--log-level", ao.logLevel)
	cmd.Dir = ao.projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set up process group so signals are propagated properly
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start UI process: %w", err)
	}

	ao.uiProcess = cmd
	ao.processActive = true

	fmt.Printf("UI process started (PID=%d)\n", cmd.Process.Pid)

	return nil
}

// stopUIProcess stops the UI child process
func (ao *AssetObserver) stopUIProcess() error {
	ao.processMutex.Lock()
	defer ao.processMutex.Unlock()

	if !ao.processActive || ao.uiProcess == nil {
		return nil
	}

	pid := ao.uiProcess.Process.Pid
	fmt.Printf("Stopping UI process (PID=%d)\n", pid)

	// Send SIGTERM to the entire process group (including go run and the binary)
	// Negative PID sends signal to the process group
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to send SIGTERM to process group: %v\n", err)
		// Fallback: try killing just the process
		ao.uiProcess.Process.Kill()
		ao.uiProcess.Wait()
		ao.uiProcess = nil
		ao.processActive = false
		return nil
	}

	// Wait for process to exit with short timeout
	done := make(chan error, 1)
	go func() {
		done <- ao.uiProcess.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Printf("UI process exited with error: %v\n", err)
		} else {
			fmt.Printf("UI process stopped gracefully\n")
		}
	case <-time.After(2 * time.Second):
		// Timeout - force kill the entire process group
		fmt.Fprintf(os.Stderr, "Warning: UI process did not stop in time, force killing\n")
		syscall.Kill(-pid, syscall.SIGKILL)
		ao.uiProcess.Process.Kill() // Also kill the main process as fallback
		// Wait again to ensure it's dead
		<-done
		fmt.Printf("UI process force killed\n")
	}

	ao.uiProcess = nil
	ao.processActive = false

	return nil
}

// regenerateAssets triggers code regeneration
func (ao *AssetObserver) regenerateAssets() error {
	fmt.Printf("Regenerating assets...\n")

	configFile := filepath.Join(ao.projectPath, "config.yaml")

	// Call internal API instead of CLI
	if err := ao.app.GenegateAssets(ao.projectPath, configFile, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to regenerate assets: %v\n", err)
		return err
	}

	fmt.Printf("Assets regenerated successfully\n")
	return nil
}

// watchLoop is the main watch loop
func (ao *AssetObserver) watchLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ao.shutdown:
			fmt.Printf("Watch loop shutting down\n")
			close(ao.done)
			return

		case <-ticker.C:
			if ao.shuttingDown {
				continue
			}

			hasChanges, changedFiles := ao.detectChanges()
			if hasChanges {
				fmt.Printf("File changes detected:\n")
				for _, file := range changedFiles {
					fmt.Printf("  - %s\n", file)
				}

				// Update last change time for debouncing
				ao.changeMutex.Lock()
				ao.lastChangeTime = time.Now()
				ao.changeMutex.Unlock()

				// Wait for debounce period
				time.Sleep(ao.debounceDelay)

				// Check if more changes occurred during debounce
				ao.changeMutex.Lock()
				if time.Since(ao.lastChangeTime) < ao.debounceDelay {
					ao.changeMutex.Unlock()
					fmt.Printf("More changes detected, waiting...\n")
					continue
				}
				ao.changeMutex.Unlock()

				// Stop UI process before regeneration - this MUST complete before we continue
				fmt.Printf("\n")
				if err := ao.stopUIProcess(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: Failed to stop UI process: %v\n", err)
					continue
				}

				// Verify process is stopped
				ao.processMutex.Lock()
				processActive := ao.processActive
				ao.processMutex.Unlock()

				if processActive {
					fmt.Fprintf(os.Stderr, "Error: UI process still active after stop attempt, skipping regeneration\n")
					continue
				}

				// Wait for port to be released
				fmt.Printf("Waiting for port to be released...\n")
				time.Sleep(1 * time.Second)

				// Regenerate assets
				if err := ao.regenerateAssets(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: Regeneration failed, not restarting UI: %v\n", err)
					continue
				}

				// Wait a moment before restarting
				time.Sleep(500 * time.Millisecond)

				// Restart UI process
				fmt.Printf("\n")
				if err := ao.startUIProcess(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: Failed to restart UI process: %v\n", err)
				}
			}
		}
	}
}

// Start begins watching files and managing the UI process
func (ao *AssetObserver) Start() error {
	fmt.Printf("Starting asset observer\n")

	// Initial file scan
	if err := ao.scanFiles(); err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	fmt.Printf("Initial file scan complete (%d files)\n", len(ao.fileHashes))

	// Start UI process
	if err := ao.startUIProcess(); err != nil {
		return fmt.Errorf("failed to start UI process: %w", err)
	}

	// Start watch loop
	go ao.watchLoop()

	return nil
}

// Stop gracefully stops the observer and child process
func (ao *AssetObserver) Stop() error {
	fmt.Printf("Stopping asset observer\n")

	ao.shuttingDown = true
	close(ao.shutdown)

	// Wait for watch loop to finish
	<-ao.done

	// Stop UI process
	return ao.stopUIProcess()
}

// Wait blocks until the observer is stopped
func (ao *AssetObserver) Wait() {
	<-ao.done
}
