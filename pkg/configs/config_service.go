package configs

import (
	"os"
	"runtime"

	"gopkg.in/yaml.v2"
)

const TEAL_VERSION = "v0.1.0"

type ConfigService struct {
}

func InitConfigService() *ConfigService {
	return &ConfigService{}
}

func (configService *ConfigService) GetProfile(projectPath string) (*Profile, error) {
	data, err := os.ReadFile(projectPath + "/profile.yaml")
	if err != nil {
		panic(err)
	}

	// Parse the YAML file
	var profile Profile
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
