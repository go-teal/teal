package generators

import (
	_ "embed"
	"os"
	"text/template"
	
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

func (g *GenMakefile) RenderToFile() error {
	tmpl, err := template.New("makefile").Parse(makefileTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(g.GetFullPath())
	if err != nil {
		return err
	}
	defer file.Close()

	var templateData = struct {
		Config         *configs.Config
		ProjectProfile *configs.ProjectProfile
	}{
		Config:         g.config,
		ProjectProfile: g.profile,
	}

	if err := tmpl.Execute(file, templateData); err != nil {
		return err
	}

	return nil
}