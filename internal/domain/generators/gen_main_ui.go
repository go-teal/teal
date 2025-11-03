package generators

import (
	_ "embed"
	"os"

	pongo2 "github.com/flosch/pongo2/v6"
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

func (g *GenMainUI) RenderToFile() (error, bool) {
	mainDirName := g.config.ProjectPath + "/cmd/" + g.profile.Name + "-ui"
	if g.config.ProjectPath == "." {
		mainDirName = "cmd/" + g.profile.Name + "-ui"
	}
	utils.CreateDir(mainDirName)

	templ, err := pongo2.FromString(mainUITemplate)
	if err != nil {
		panic(err)
	}

	_, err = os.Stat(g.GetFullPath())

	if !os.IsNotExist(err) {
		return nil, true
	}

	connectionsFlags := make(map[string]bool)

	for _, c := range g.config.Connections {
		connectionsFlags[c.Type] = true
	}

	output, err := templ.Execute(pongo2.Context{
		"Profile":     g.profile,
		"Config":      g.config,
		"Connections": connectionsFlags,
	})
	if err != nil {
		panic(err)
	}

	file, err := os.Create(g.GetFullPath())

	if err != nil {
		panic(err)
	}

	defer file.Close()

	_, err = file.WriteString(output)

	return err, false
}
