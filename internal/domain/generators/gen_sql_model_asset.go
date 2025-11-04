package generators

import (
	_ "embed"
	"encoding/base64"
	"os"

	pongo2 "github.com/flosch/pongo2/v6"
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

func (g *GenSQLModelAsset) RenderToFile() (error, bool) {

	dirName := g.config.ProjectPath + "/internal/assets/"
	utils.CreateDir(dirName)

	// Base64 encode the description to avoid issues with special characters in templates
	if g.modelConfig.ModelProfile != nil && g.modelConfig.ModelProfile.Description != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(g.modelConfig.ModelProfile.Description))
		g.modelConfig.ModelProfile.Description = encoded
	}

	goTempl, err := pongo2.FromString(dwhModelTemplate)
	if err != nil {
		return err, false
	}

	g.modelConfig.ModelFieldsFunc = "{{ ModelFields }}"

	// Extract materialization for easier template access
	var materialization string
	if g.modelConfig.ModelProfile != nil {
		materialization = string(g.modelConfig.ModelProfile.Materialization)
	}

	output, err := goTempl.Execute(pongo2.Context{
		"ModelName":            g.modelConfig.ModelName,
		"GoName":               g.modelConfig.GoName,
		"NameUpperCase":        g.modelConfig.NameUpperCase,
		"SqlByteBuffer":        g.modelConfig.SqlByteBuffer.String(),
		"ModelFieldsFunc":      g.modelConfig.ModelFieldsFunc,
		"ModelProfile":         g.modelConfig.ModelProfile,
		"Materialization":      materialization,
		"PrimaryKeyExpression": g.modelConfig.PrimaryKeyExpression,
		"Indexes":              g.modelConfig.Indexes,
		"Upstreams":            g.modelConfig.Upstreams,
		"Downstreams":          g.modelConfig.Downstreams,
	})
	if err != nil {
		return err, false
	}

	file, err := os.Create(g.GetFullPath())

	if err != nil {
		panic(err)
	}

	defer file.Close()

	_, err = file.WriteString(output)
	return err, false
}
