package generators

import (
	_ "embed"
	"os"

	pongo2 "github.com/flosch/pongo2/v6"
	"github.com/go-teal/teal/pkg/configs"
)

//go:embed templates/go.mod.tmpl
var goModTemplate string

const GO_MOD_FILE_NAME = "go.mod"

type GenGoMod struct {
	config  *configs.Config
	profile *configs.ProjectProfile
}

func InitGenGoMod(config *configs.Config, profile *configs.ProjectProfile) Generator {
	return &GenGoMod{
		config:  config,
		profile: profile,
	}
}

// GetFileName implements Generator.
func (g *GenGoMod) GetFileName() string {
	return GO_MOD_FILE_NAME
}

// GetFullPath implements Generator.
func (g *GenGoMod) GetFullPath() string {
	return g.config.ProjectPath + "/" + GO_MOD_FILE_NAME
}

func (g *GenGoMod) RenderToFile() error {

	_, err := os.Stat(g.GetFullPath())
	if err == nil {

		return nil
	}

	templ, err := pongo2.FromString(goModTemplate)
	if err != nil {
		panic(err)
	}

	output, err := templ.Execute(pongo2.Context{
		"Config":  g.config,
		"Version": configs.TEAL_VERSION,
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
	return err
}
