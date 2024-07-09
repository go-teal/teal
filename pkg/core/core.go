package core

import (
	"sync"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/drivers"
)

// GO singletone
type Core struct {
	dbConnections map[string]drivers.DBDriver
	Config        *configs.Config
	Profile       *configs.ProjectProfile
}

var core *Core
var once sync.Once

func GetInstance() *Core {
	once.Do(func() {

		core = &Core{
			dbConnections: make(map[string]drivers.DBDriver),
		}
	})
	return core
}

func (c *Core) Init(configFileName string, projectPath string) {
	cs := configs.InitConfigService()
	config, err := cs.GetConfig(configFileName, projectPath)
	if err != nil {
		panic(err)
	}
	profile, err := cs.GetProfileProfile(projectPath)
	if err != nil {
		panic(err)
	}

	c.Config = config
	c.Profile = profile

	for _, connectionConfig := range c.Config.Connections {
		dbConnection, err := drivers.EstablishDBConnection(connectionConfig)
		if err != nil {
			panic(err)
		}
		if dbConnection.IsPermanent() {
			err = dbConnection.Connect()
			if err != nil {
				panic(err)
			}
		}
		c.dbConnections[connectionConfig.Name] = dbConnection
	}
	// fmt.Printf("Connections %s have been initialized\n", c.dbConnections)
}

func (c *Core) GetDBConnection(connection string) drivers.DBDriver {
	return c.dbConnections[connection]
}

func (c *Core) Shutdown() {
	for _, dbConnection := range c.dbConnections {
		if dbConnection.IsPermanent() {
			dbConnection.Close()
		}
	}
}
