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

//go:embed templates/dwh_sql_model_asset.tmpl
var dwhModelTemplate string

type GenSQLModelAsset struct {
	config         *configs.Config
	projectProfile *configs.ProjectProfile
	modelConfig    *internalmodels.ModelConfig
}

func InitGenModelSQLAsset(
	config *configs.Config,
	projectProfile *configs.ProjectProfile,
	modelConfig *internalmodels.ModelConfig,

) Generator {
	return &GenSQLModelAsset{
		config:         config,
		projectProfile: projectProfile,
		modelConfig:    modelConfig,
	}
}

// GetFileName implements Generator.
func (g *GenSQLModelAsset) GetFileName() string {
	return g.modelConfig.ModelName
}

// GetFullPath implements Generator.
func (g *GenSQLModelAsset) GetFullPath() string {
	return g.config.ProjectPath + "/internal/assets/" + g.GetFileName() + ".go"
}

func (g *GenSQLModelAsset) RenderToFile() error {

	dirName := g.config.ProjectPath + "/internal/assets/"
	utils.CreateDir(dirName)

	// Base64 encode the description to avoid issues with special characters in templates
	if g.modelConfig.ModelProfile != nil && g.modelConfig.ModelProfile.Description != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(g.modelConfig.ModelProfile.Description))
		g.modelConfig.ModelProfile.Description = encoded
	}

	goTempl, err := template.New(g.GetFileName()).Parse(dwhModelTemplate)
	if err != nil {
		return err
	}
	var goByteBuffer bytes.Buffer
	g.modelConfig.ModelFieldsFunc = "{{ ModelFields }}"
	err = goTempl.Execute(&goByteBuffer, g.modelConfig)
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
