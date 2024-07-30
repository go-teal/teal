package generators

import (
	_ "embed"
	"os"
	"text/template"

	internalmodels "github.com/go-teal/teal/internal/domain/internal_models"
	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/testing_config.go.tmpl
var goTesgingConfigTemplate string

const GO_TESTING_CONFIG_FILE_NAME = "configs.go"

type GenTestingConfig struct {
	config      *configs.Config
	profile     *configs.ProjectProfile
	testConfigs []*internalmodels.TestConfig
}

// GetFileName implements Generator.
func (g *GenTestingConfig) GetFileName() string {
	return GO_TESTING_CONFIG_FILE_NAME
}

// GetFullPath implements Generator.
func (g *GenTestingConfig) GetFullPath() string {
	return g.config.ProjectPath + "/internal/model_tests/" + GO_ASSETS_CONFIG_FILE_NAME
}

func InitGenTestConfig(
	config *configs.Config,
	profile *configs.ProjectProfile,
	testConfigs []*internalmodels.TestConfig,
) Generator {
	return &GenTestingConfig{
		config:      config,
		profile:     profile,
		testConfigs: testConfigs,
	}
}

func (g *GenTestingConfig) RenderToFile() error {
	// fmt.Printf("Rendering: %s", g.GetFullPath())
	dirName := g.config.ProjectPath + "/internal/model_tests/"
	utils.CreateDir(dirName)
	templ, err := template.New(GO_TESTING_CONFIG_FILE_NAME).Parse(goTesgingConfigTemplate)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(g.GetFullPath())

	if err != nil {
		panic(err)
	}

	defer file.Close()

	data := struct {
		Config *configs.Config
		Tests  []*internalmodels.TestConfig
	}{
		Config: g.config,
		Tests:  g.testConfigs,
	}
	err = templ.Execute(file, data)
	return err
}
