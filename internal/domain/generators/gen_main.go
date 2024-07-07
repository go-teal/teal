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

type GenMain struct {
	config  *configs.Config
	profile *configs.ProjectProfile
}

// GetFileName implements Generator.
func (g *GenMain) GetFileName() string {
	return g.profile.Name + "._go"
}

// GetFullPath implements Generator.
func (g *GenMain) GetFullPath() string {
	return g.config.ProjectPath + "/cmd/" + g.profile.Name + "/" + g.GetFileName()
}

func InitGenMain(config *configs.Config, profile *configs.ProjectProfile) Generator {
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

	templ, err := template.New(g.GetFileName()).Parse(mainTemplate)
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
		Profile *configs.ProjectProfile
		Config  *configs.Config
	}{
		Profile: g.profile,
		Config:  g.config,
	}
	err = templ.Execute(file, data)
	return err
}
