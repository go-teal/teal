package ui

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-teal/teal/pkg/dags"
	"github.com/go-teal/teal/pkg/services/debugging"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type UIServer struct {
	ProjectName      string
	ModuleName       string
	Port             int
	debuggingService *debugging.DebuggingService
	logWriter        interface{} // Store the log writer (interface to avoid import cycle)
}

func NewUIServer(projectName, moduleName string, port int, dag *dags.DebugDag) *UIServer {
	return &UIServer{
		ProjectName:      projectName,
		ModuleName:       moduleName,
		Port:             port,
		debuggingService: debugging.NewDebuggingService(dag),
	}
}

func NewUIServerWithLogWriter(projectName, moduleName string, port int, dag *dags.DebugDag, logWriter interface{}) *UIServer {
	return &UIServer{
		ProjectName:      projectName,
		ModuleName:       moduleName,
		Port:             port,
		debuggingService: debugging.NewDebuggingService(dag),
		logWriter:        logWriter,
	}
}

type DagResponseDTO struct {
	ProjectName     string                 `json:"projectName"`
	ModuleName      string                 `json:"moduleName"`
	DagInstanceName string                 `json:"dagInstanceName"`
	Nodes           []debugging.DagNodeDTO `json:"nodes"`
}

type TestProfilesResponseDTO struct {
	Tests []debugging.TestProfileDTO `json:"tests"`
}

func (s *UIServer) Start() error {
	log.Info().Int("port", s.Port).Msg("Starting debug UI server with Gin")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Configure CORS to allow all origins
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	r.GET("/api/dag", s.handleDagData)
	r.GET("/api/tests", s.handleTestProfiles)
	r.GET("/api/tests/:taskId", s.handleTestResultsForTask)
	r.POST("/api/dag/run", s.handleDagRun)
	r.GET("/api/dag/status/:taskId", s.handleDagStatus)
	r.GET("/api/dag/tasks", s.handleDagTasks)
	r.POST("/api/dag/asset/:name/mutate", s.handleAssetMutate)
	r.GET("/api/dag/asset/:name/data", s.handleAssetData)
	r.POST("/api/dag/asset/:name/select", s.handleAssetSelect)
	r.POST("/api/dag/reset", s.handleDagReset)

	// Log endpoints (only available when logWriter is configured)
	if s.logWriter != nil {
		r.GET("/api/logs/:taskId", s.handleGetLogs)
		r.GET("/api/logs", s.handleGetAllLogs)
		r.DELETE("/api/logs/:taskId", s.handleClearLogs)
		r.DELETE("/api/logs", s.handleClearAllLogs)
	}

	addr := fmt.Sprintf(":%d", s.Port)
	log.Info().Str("address", addr).Msg("UI server listening")
	return r.Run(addr)
}

func (s *UIServer) handleDagData(c *gin.Context) {
	nodes := s.debuggingService.GetDagNodes()

	if len(nodes) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DAG not initialized or empty"})
		return
	}

	c.JSON(http.StatusOK, DagResponseDTO{
		ProjectName:     s.ProjectName,
		ModuleName:      s.ModuleName,
		DagInstanceName: s.debuggingService.GetDagInstanceName(),
		Nodes:           nodes,
	})
}

func (s *UIServer) handleTestProfiles(c *gin.Context) {
	tests := s.debuggingService.GetTestProfiles()

	c.JSON(http.StatusOK, TestProfilesResponseDTO{
		Tests: tests,
	})
}

func (s *UIServer) handleTestResultsForTask(c *gin.Context) {
	taskId := c.Param("taskId")

	if taskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required"})
		return
	}

	tests := s.debuggingService.GetTestResultsForTask(taskId)

	c.JSON(http.StatusOK, gin.H{
		"taskId": taskId,
		"tests":  tests,
	})
}

func (s *UIServer) handleDagRun(c *gin.Context) {
	var request debugging.DagRunRequestDTO

	// Parse request body
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Validate taskId
	if request.TaskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required"})
		return
	}

	// Execute DAG with timeout
	responseChan := s.debuggingService.ExecuteDag(request.TaskId, request.Data)

	// Wait for response (will timeout after 10 seconds as configured in ExecuteDag)
	response := <-responseChan

	// Always return 200 OK regardless of test failures
	// Only return error status for actual DAG execution failures (not test failures)
	statusCode := http.StatusOK
	if response.Status == debugging.DagExecutionStatusPending {
		statusCode = http.StatusAccepted // 202 for async operations still in progress
	}

	c.JSON(statusCode, response)
}

func (s *UIServer) handleDagStatus(c *gin.Context) {
	taskId := c.Param("taskId")

	if taskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required"})
		return
	}

	// Get current task status
	status := s.debuggingService.GetTaskStatus(taskId)

	// Return appropriate status code based on execution status
	statusCode := http.StatusOK
	if status.Status == debugging.DagExecutionStatusNotStarted {
		statusCode = http.StatusNotFound
	}

	c.JSON(statusCode, status)
}

func (s *UIServer) handleDagTasks(c *gin.Context) {
	taskList := s.debuggingService.GetAllTasks()
	c.JSON(http.StatusOK, taskList)
}

func (s *UIServer) handleAssetMutate(c *gin.Context) {
	assetName := c.Param("name")

	if assetName == "" {
		// Still return 400 for bad request parameters
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset name is required"})
		return
	}

	var request debugging.AssetExecuteRequestDTO

	// Parse request body
	if err := c.ShouldBindJSON(&request); err != nil {
		// Still return 400 for bad request format
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Validate taskId
	if request.TaskId == "" {
		// Still return 400 for missing required parameters
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required"})
		return
	}

	// Generate taskUUID if not provided
	taskUUID := request.TaskUUID
	if taskUUID == "" {
		taskUUID = uuid.New().String()
	}

	// Execute asset with timeout
	responseChan := s.debuggingService.MutateAsset(assetName, request.TaskId, taskUUID)

	// Wait for response (will timeout after 10 seconds as configured in ExecuteAsset)
	response := <-responseChan

	// Always return 200 or 202, even if the execution failed
	// The failure status is included in the DTO response
	statusCode := http.StatusOK

	// Return 202 Accepted only if the operation is still in progress
	if response.Status == debugging.NodeStateInProgress {
		statusCode = http.StatusAccepted
	}

	// For all other states (including failures), return 200 OK
	// The actual execution status is in the response DTO
	c.JSON(statusCode, response)
}

func (s *UIServer) handleDagReset(c *gin.Context) {
	err := s.debuggingService.ResetDagState()

	if err != nil {
		log.Error().Err(err).Msg("Failed to reset DAG state")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reset DAG state",
			"details": err.Error(),
		})
		return
	}

	log.Info().Msg("DAG state reset successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "DAG state reset successfully",
		"status":  "SUCCESS",
	})
}

func (s *UIServer) handleAssetData(c *gin.Context) {
	assetName := c.Param("name")

	if assetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset name is required"})
		return
	}

	// Get taskId from query parameter
	taskId := c.Query("taskId")
	if taskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required as query parameter"})
		return
	}

	// Get pagination parameters
	offset := 0
	limit := 0 // 0 means no limit

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be a non-negative integer"})
			return
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val >= 0 {
			limit = val
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a non-negative integer"})
			return
		}
	}

	// Get asset data for the specific task
	response := s.debuggingService.GetAssetData(taskId, assetName, offset, limit)

	// Always return 200 or 202, similar to handleAssetExecute
	statusCode := http.StatusOK

	// Return 202 if still in progress
	if response.Status == debugging.NodeStateInProgress {
		statusCode = http.StatusAccepted
	}

	c.JSON(statusCode, response)
}

func (s *UIServer) handleAssetSelect(c *gin.Context) {
	assetName := c.Param("name")

	if assetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset name is required"})
		return
	}

	// Parse request body to get taskId
	var request debugging.AssetExecuteRequestDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Validate taskId
	if request.TaskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required"})
		return
	}

	// Execute the asset's SQL query using ToDataFrame and save to node's LastResult
	response := s.debuggingService.ExecuteAssetSelect(assetName, request.TaskId)

	// Return appropriate status code
	statusCode := http.StatusOK
	if response.Status == debugging.NodeStateFailed {
		statusCode = http.StatusOK // Still return 200 with error in response body
	}

	c.JSON(statusCode, response)
}

// Log handler functions

type LogWriter interface {
	GetLogs(taskId string) []interface{}
	GetAllLogs() map[string][]interface{}
	ClearLogs(taskId string)
	ClearAllLogs()
}

func (s *UIServer) handleGetLogs(c *gin.Context) {
	if s.logWriter == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Log writer not configured"})
		return
	}

	taskId := c.Param("taskId")
	if taskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required"})
		return
	}

	logWriter, ok := s.logWriter.(LogWriter)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid log writer type"})
		return
	}

	logs := logWriter.GetLogs(taskId)
	c.JSON(http.StatusOK, gin.H{
		"taskId": taskId,
		"logs":   logs,
		"count":  len(logs),
	})
}

func (s *UIServer) handleGetAllLogs(c *gin.Context) {
	if s.logWriter == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Log writer not configured"})
		return
	}

	logWriter, ok := s.logWriter.(LogWriter)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid log writer type"})
		return
	}

	allLogs := logWriter.GetAllLogs()

	// Calculate total log count
	totalCount := 0
	for _, logs := range allLogs {
		totalCount += len(logs)
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":          allLogs,
		"taskCount":     len(allLogs),
		"totalLogCount": totalCount,
	})
}

func (s *UIServer) handleClearLogs(c *gin.Context) {
	if s.logWriter == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Log writer not configured"})
		return
	}

	taskId := c.Param("taskId")
	if taskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required"})
		return
	}

	logWriter, ok := s.logWriter.(LogWriter)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid log writer type"})
		return
	}

	logWriter.ClearLogs(taskId)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Logs cleared for task: %s", taskId),
		"taskId":  taskId,
	})
}

func (s *UIServer) handleClearAllLogs(c *gin.Context) {
	if s.logWriter == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Log writer not configured"})
		return
	}

	logWriter, ok := s.logWriter.(LogWriter)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid log writer type"})
		return
	}

	logWriter.ClearAllLogs()
	c.JSON(http.StatusOK, gin.H{
		"message": "All logs cleared successfully",
	})
}
