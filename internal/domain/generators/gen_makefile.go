package generators

import (
	_ "embed"
	"os"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/Makefile.tmpl
var makefileTemplate string

type GenMakefile struct {
	config  *configs.Config
	profile *configs.ProjectProfile
}

func InitGenMakefile(config *configs.Config, projectProfile *configs.ProjectProfile) *GenMakefile {
	return &GenMakefile{
		config:  config,
		profile: projectProfile,
	}
}

func (g *GenMakefile) GetFileName() string {
	return "Makefile"
}

func (g *GenMakefile) GetFullPath() string {
	return g.config.ProjectPath + "/" + g.GetFileName()
}

func (g *GenMakefile) RenderToFile() (error, bool) {
	tmpl, err := pongo2.FromString(makefileTemplate)
	if err != nil {
		return err, false
	}

	output, err := tmpl.Execute(pongo2.Context{
		"Config":         g.config,
		"ProjectProfile": g.profile,
	})
	if err != nil {
		return err, false
	}

	file, err := os.Create(g.GetFullPath())
	if err != nil {
		return err, false
	}
	defer file.Close()

	_, err = file.WriteString(output)
	return err, false
}
