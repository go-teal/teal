package generators

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"os"
	"text/template"

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

	// Base64 encode the description to avoid issues with special characters in templates
	if g.testConfig.TestProfile != nil && g.testConfig.TestProfile.Description != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(g.testConfig.TestProfile.Description))
		g.testConfig.TestProfile.Description = encoded
	}

	goTempl, err := template.New(g.GetFileName()).Parse(dwhModelTestTemplate)
	if err != nil {
		return err
	}
	var goByteBuffer bytes.Buffer
	err = goTempl.Execute(&goByteBuffer, g.testConfig)
	if err != nil {
		return err
	}

	file, err := os.Create(g.GetFullPath())

	if err != nil {
		panic(err)
	}

	defer file.Close()

	_, err = file.Write(goByteBuffer.Bytes())
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
