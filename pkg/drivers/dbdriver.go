package drivers

import (
	"github.com/go-teal/gota/dataframe"
	"github.com/go-teal/teal/pkg/configs"
)

type DBDriver interface {
	Connect() error
	Begin() (interface{}, error)
	Commit(tx interface{}) error
	Rallback(tx interface{}) error
	Close() error
	Exec(tx interface{}, sql string) error
	ToDataFrame(sql string) (*dataframe.DataFrame, error)
	GetListOfFields(tx interface{}, tableName string) []string
	CheckTableExists(tx interface{}, tableName string) bool
	CheckSchemaExists(tx interface{}, schemaName string) bool
	IsPermanent() bool
	MountSource(sourceProfile *configs.SourceProfile) error
	UnMountSource(sourceProfile *configs.SourceProfile) error
	GetRawConnection() interface{}
}

type DBconnectionFactory interface {
	CreateConnection(connection configs.DBConnectionConfig) (DBDriver, error)
}
