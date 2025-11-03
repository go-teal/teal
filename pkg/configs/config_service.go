package configs

import (
	"os"
	"runtime"

	"gopkg.in/yaml.v2"
)

const TEAL_VERSION = "v1.0.1"

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
