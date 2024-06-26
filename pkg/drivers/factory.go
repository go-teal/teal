package drivers

import (
	"fmt"

	"github.com/go-teal/teal/pkg/configs"
)

var Factories = map[string]DBconnectionFactory{
	"duckdb": InitDuckDBEnginFactory(),
}

func RegisterConnectionFactory(driverName string, f DBconnectionFactory) {
	Factories[driverName] = f
}

func EstablishDBConnection(connection *configs.DBConnectionConfig) (DBDriver, error) {

	if factory, ok := Factories[connection.Type]; ok {
		return factory.CreateConnection(*connection)
	}
	return nil, fmt.Errorf("driver %s not found", connection.Type)
}
