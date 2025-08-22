package ui

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-teal/teal/pkg/dags"
	"github.com/go-teal/teal/pkg/services/debugging"
	"github.com/rs/zerolog/log"
)

type UIServer struct {
	ProjectName      string
	ModuleName       string
	Port             int
	debuggingService *debugging.DebuggingService
}

func NewUIServer(projectName, moduleName string, port int, dag *dags.DebugDag) *UIServer {
	return &UIServer{
		ProjectName:      projectName,
		ModuleName:       moduleName,
		Port:             port,
		debuggingService: debugging.NewDebuggingService(dag),
	}
}

type DagResponseDTO struct {
	DagInstanceName string                  `json:"dagInstanceName"`
	Nodes           []debugging.DagNodeDTO `json:"nodes"`
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

	addr := fmt.Sprintf(":%d", s.Port)
	log.Info().Str("address", addr).Msg("UI server listening")
	return r.Run(addr)
}

func (s *UIServer) handleDagData(c *gin.Context) {
	nodes := s.debuggingService.GetDagNodes()

	if nodes == nil || len(nodes) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DAG not initialized or empty"})
		return
	}

	c.JSON(http.StatusOK, DagResponseDTO{
		DagInstanceName: s.debuggingService.GetDagInstanceName(),
		Nodes:           nodes,
	})
}
