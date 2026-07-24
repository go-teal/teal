package configs

import (
	"os"
	"runtime"
	"runtime/debug"

	"gopkg.in/yaml.v2"
)

// TEAL_VERSION is the CLI version. Overridden at build/release time via
// -ldflags "-X github.com/go-teal/teal/pkg/configs.TEAL_VERSION=vX.Y.Z".
var TEAL_VERSION = "dev"

// ResolvedVersion returns the version to display/emit. A release build stamps
// TEAL_VERSION via -ldflags and that wins. When the binary is installed with
// `go install github.com/go-teal/teal/cmd/teal@vX.Y.Z` (no ldflags), it falls
// back to the module version recorded in the build info, so `teal version`
// reports the real tag instead of "dev".
func ResolvedVersion() string {
	if TEAL_VERSION != "" && TEAL_VERSION != "dev" {
		return TEAL_VERSION
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
		for _, d := range bi.Deps {
			if d.Path == "github.com/go-teal/teal" && d.Version != "" {
				return d.Version
			}
		}
	}
	return "dev"
}

type ConfigService struct {
}

func InitConfigService() *ConfigService {
	return &ConfigService{}
}

func (configService *ConfigService) GetProfileProfile(projectPath string) (*ProjectProfile, error) {
	data, err := os.ReadFile(projectPath + "/profile.yaml")
	if err != nil {
		panic(err)
	}

	// Parse the YAML file
	var profile ProjectProfile
	err = yaml.Unmarshal(data, &profile)
	if err != nil {
		panic(err)
	}

	return &profile, nil
}

func (configService *ConfigService) GetConfig(configFileName string, projectPath string) (*Config, error) {
	data, err := os.ReadFile(configFileName)
	if err != nil {
		panic(err)
	}

	// Parse the YAML file
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	config.ProjectPath = projectPath
	if config.Cores == 0 {
		config.Cores = runtime.NumCPU()
	}
	return &config, nil
}
