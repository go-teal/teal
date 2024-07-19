package internalmodels

import (
	"bytes"

	"github.com/go-teal/teal/pkg/configs"
)

type TestConfig struct {
	GoName        string
	TestName      string
	NameUpperCase string
	SqlByteBuffer bytes.Buffer
	TestProfile   *configs.TestProfile
}
