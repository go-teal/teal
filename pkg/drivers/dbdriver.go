package drivers

import (
	"github.com/go-teal/gota/dataframe"
	"github.com/go-teal/teal/pkg/configs"
)

type DBDriver interface {
	Connect() error
	Begin() (interface{}, error)
	Commit(tx interface{}) error
	Rollback(tx interface{}) error
	Close() error
	Exec(tx interface{}, sql string) error
	ToDataFrame(sql string) (*dataframe.DataFrame, error)
	PersistDataFrame(tx interface{}, name string, df *dataframe.DataFrame) error
	GetListOfFields(tx interface{}, tableName string) []string
	CheckTableExists(tx interface{}, tableName string) bool
	CheckSchemaExists(tx interface{}, schemaName string) bool
	// CreateSchema creates schemaName if it is missing. Implementations must be
	// idempotent and safe to call concurrently - several DAG nodes of the same
	// stage can discover the same missing schema at once.
	CreateSchema(tx interface{}, schemaName string) error
	GetRawConnection() interface{}
	SimpleTest(sql string) (string, error)
	ConcurrencyLock()
	ConcurrencyUnlock()
}

type DBconnectionFactory interface {
	CreateConnection(connection configs.DBConnectionConfig) (DBDriver, error)
}
