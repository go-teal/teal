package application

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
)

//go:embed templates/scaffold.tar.gz
var scaffoldTar string

func (app *Application) InitProject() error {
	fmt.Println("initializing a new project")
	file, err := os.Create("scaffold.tar.gz")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Scaffold tar %d bytes\n", len(scaffoldTar))
	_, err = file.WriteString(scaffoldTar)
	if err != nil {
		panic(err)
	}
	file.Close()

	cmd := exec.Command("tar", "-xzf", "scaffold.tar.gz", "-C", ".")
	_, err = cmd.Output()
	if err != nil {
		panic(err)
	}

	cmd = exec.Command("rm", "scaffold.tar.gz")
	_, err = cmd.Output()
	if err != nil {
		panic(err)
	}

	return nil
}
