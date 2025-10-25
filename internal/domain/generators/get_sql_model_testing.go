package generators

import (
	_ "embed"
	"os"

	pongo2 "github.com/flosch/pongo2/v6"
	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/dwh_sql_model_test.tmpl
var dwhModelTestTemplate string

type GenSQLModelTest struct {
	config         *configs.Config
	projectProfile *configs.ProjectProfile
	testConfig     *internalmodels.TestConfig
}

// GetFileName implements Generator.
func (g *GenSQLModelTest) GetFileName() string {
	return g.testConfig.TestName
}

// GetFullPath implements Generator.
func (g *GenSQLModelTest) GetFullPath() string {
	return g.config.ProjectPath + "/internal/model_tests/" + g.GetFileName() + ".go"
}

// RenderToFile implements Generator.
func (g *GenSQLModelTest) RenderToFile() error {
	dirName := g.config.ProjectPath + "/internal/model_tests/"
	utils.CreateDir(dirName)

	goTempl, err := pongo2.FromString(dwhModelTestTemplate)
	if err != nil {
		return err
	}

	output, err := goTempl.Execute(pongo2.Context{
		"TestName":      g.testConfig.TestName,
		"GoName":        g.testConfig.GoName,
		"NameUpperCase": g.testConfig.NameUpperCase,
		"SqlByteBuffer": g.testConfig.SqlByteBuffer,
		"TestProfile":   g.testConfig.TestProfile,
	})
	if err != nil {
		return err
	}

	file, err := os.Create(g.GetFullPath())

	if err != nil {
		panic(err)
	}

	defer file.Close()

	_, err = file.WriteString(output)
	return err
}

func InitGenSQLModelTest(
	config *configs.Config,
	projectProfile *configs.ProjectProfile,
	testConfig *internalmodels.TestConfig,
) Generator {
	return &GenSQLModelTest{
		config:         config,
		projectProfile: projectProfile,
		testConfig:     testConfig,
	}
}
