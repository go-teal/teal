package generators

import (
	_ "embed"
	"os"
	"text/template"

	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/main.go.tmpl
var mainTemplate string

const MAIN_FILE_NAME = "main._go"

type GenMain struct {
	config  *configs.Config
	profile *configs.Profile
}

// GetFileName implements Generator.
func (g *GenMain) GetFileName() string {
	return MAIN_FILE_NAME
}

// GetFullPath implements Generator.
func (g *GenMain) GetFullPath() string {
	return g.config.ProjectPath + "/cmd/" + g.profile.Name + "/" + MAIN_FILE_NAME
}

func InitGenMain(config *configs.Config, profile *configs.Profile) Generator {
	return &GenMain{
		config:  config,
		profile: profile,
	}
}

func (g *GenMain) RenderToFile() error {

	mainDirName := g.config.ProjectPath + "/cmd/" + g.profile.Name
	if g.config.ProjectPath == "." {
		mainDirName = "cmd/" + g.profile.Name
	}
	utils.CreateDir(mainDirName)

	templ, err := template.New(MAIN_FILE_NAME).Parse(mainTemplate)
	if err != nil {
		panic(err)
	}

	_, err = os.Stat(g.GetFullPath())

	if !os.IsNotExist(err) {
		return nil
	}

	file, err := os.Create(g.GetFullPath())

	if err != nil {
		panic(err)
	}

	defer file.Close()

	data := struct {
		Profile *configs.Profile
		Config  *configs.Config
	}{
		Profile: g.profile,
		Config:  g.config,
	}
	err = templ.Execute(file, data)
	return err
}
