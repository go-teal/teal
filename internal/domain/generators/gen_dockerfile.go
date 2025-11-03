package generators

import (
	_ "embed"
	"os"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/Dockerfile.tmpl
var dockerfileTemplate string

type GenDockerfile struct {
	config  *configs.Config
	profile *configs.ProjectProfile
}

func InitGenDockerfile(config *configs.Config, projectProfile *configs.ProjectProfile) *GenDockerfile {
	return &GenDockerfile{
		config:  config,
		profile: projectProfile,
	}
}

func (g *GenDockerfile) GetFileName() string {
	return "Dockerfile"
}

func (g *GenDockerfile) GetFullPath() string {
	return g.config.ProjectPath + "/" + g.GetFileName()
}

func (g *GenDockerfile) RenderToFile() (error, bool) {
	// Check if Dockerfile already exists
	if _, err := os.Stat(g.GetFullPath()); err == nil {
		// File exists, skip generation
		return nil, true
	}

	tmpl, err := pongo2.FromString(dockerfileTemplate)
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
