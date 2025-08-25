package generators

import (
	_ "embed"
	"os"
	"text/template"

	"github.com/go-teal/teal/internal/domain/utils"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/main_ui.go.tmpl
var mainUITemplate string

type GenMainUI struct {
	config  *configs.Config
	profile *configs.ProjectProfile
}

// GetFileName implements Generator.
func (g *GenMainUI) GetFileName() string {
	return g.profile.Name + "-ui.go"
}

// GetFullPath implements Generator.
func (g *GenMainUI) GetFullPath() string {
	return g.config.ProjectPath + "/cmd/" + g.profile.Name + "-ui/" + g.GetFileName()
}

func InitGenMainUI(config *configs.Config, profile *configs.ProjectProfile) Generator {
	return &GenMainUI{
		config:  config,
		profile: profile,
	}
}

func (g *GenMainUI) RenderToFile() error {
	mainDirName := g.config.ProjectPath + "/cmd/" + g.profile.Name + "-ui"
	if g.config.ProjectPath == "." {
		mainDirName = "cmd/" + g.profile.Name + "-ui"
	}
	utils.CreateDir(mainDirName)

	templ, err := template.New(g.GetFileName()).Parse(mainUITemplate)
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

	connectionsFlags := make(map[string]bool)

	for _, c := range g.config.Connections {
		connectionsFlags[c.Type] = true
	}

	data := struct {
		Profile     *configs.ProjectProfile
		Config      *configs.Config
		Connections map[string]bool
	}{
		Profile:     g.profile,
		Config:      g.config,
		Connections: connectionsFlags,
	}

	err = templ.Execute(file, data)

	return err
}