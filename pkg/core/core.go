package core

import (
	"os"
	"strconv"
	"sync"

	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/drivers"
	"github.com/rs/zerolog/log"
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
		preLoadEnvs(connectionConfig)
		dbConnection, err := drivers.EstablishDBConnection(connectionConfig)
		if err != nil {
			panic(err)
		}

		err = dbConnection.Connect()
		if err != nil {
			panic(err)
		}
		c.dbConnections[connectionConfig.Name] = dbConnection
	}
}

func (c *Core) GetDBConnection(connection string) drivers.DBDriver {
	return c.dbConnections[connection]
}

func (c *Core) Shutdown() {
	for _, dbConnection := range c.dbConnections {
		dbConnection.Close()
	}
}

func preLoadEnvs(connectionConfig *configs.DBConnectionConfig) {
	if connectionConfig.Config.HostEnv != "" {
		if value, ok := os.LookupEnv(connectionConfig.Config.HostEnv); ok {
			connectionConfig.Config.Host = value
		}
	}

	if connectionConfig.Config.PortEnv != "" {
		if value, ok := os.LookupEnv(connectionConfig.Config.PortEnv); ok {
			var err error
			connectionConfig.Config.Port, err = strconv.Atoi(value)
			if err != nil {
				log.Error().Caller().Msgf("Error parsing port env: %s", connectionConfig.Config.PortEnv)
				panic(err)
			}
		}
	}

	if connectionConfig.Config.DatabaseEnv != "" {
		if value, ok := os.LookupEnv(connectionConfig.Config.DatabaseEnv); ok {
			connectionConfig.Config.Database = value
		}
	}

	if connectionConfig.Config.UserEnv != "" {
		if value, ok := os.LookupEnv(connectionConfig.Config.UserEnv); ok {
			connectionConfig.Config.User = value
		}
	}

	if connectionConfig.Config.PasswordEnv != "" {
		if value, ok := os.LookupEnv(connectionConfig.Config.PasswordEnv); ok {
			connectionConfig.Config.Password = value
		}
	}

	if connectionConfig.Config.PathEnv != "" {
		if value, ok := os.LookupEnv(connectionConfig.Config.PathEnv); ok {
			connectionConfig.Config.Path = value
		}
	}

	for _, env := range connectionConfig.Config.ExtraParams {
		if value, ok := os.LookupEnv(env.ValueEnv); ok {
			env.Value = value
		}
	}

	if connectionConfig.Config.DBRootCertEnv != "" {
		if value, ok := os.LookupEnv(connectionConfig.Config.DBRootCertEnv); ok {
			connectionConfig.Config.DBRootCert = value
		}
	}

	if connectionConfig.Config.DBRCertEnv != "" {
		if value, ok := os.LookupEnv(connectionConfig.Config.DBRCertEnv); ok {
			connectionConfig.Config.DBCert = value
		}
	}

	if connectionConfig.Config.DBKeyEnv != "" {
		if value, ok := os.LookupEnv(connectionConfig.Config.DBKeyEnv); ok {
			connectionConfig.Config.DBKey = value
		}
	}

	if connectionConfig.Config.DBSSLModeEnv != "" {
		if value, ok := os.LookupEnv(connectionConfig.Config.DBSSLModeEnv); ok {
			connectionConfig.Config.DBSSLModeEnv = value
		}
	}
}
