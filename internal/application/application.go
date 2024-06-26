package application

import (
	"github.com/go-teal/teal/internal/domain/services"
	"github.com/go-teal/teal/pkg/configs"
)

type Application struct {
	configService   *configs.ConfigService
	dependnacyGraph *services.DependnacyGraph
}

func InitApplication() *Application {
	return &Application{
		configService:   configs.InitConfigService(),
		dependnacyGraph: services.InitDependnacyGraph(),
	}
}
