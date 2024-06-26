package configs

type Config struct {
	ProjectPath string
	Version     string                `yaml:"version"`
	Module      string                `yaml:"module"`
	Connections []*DBConnectionConfig `yaml:"connections"`
	Cores       int                   `yaml:"cores"`
}

type DBConnectionConfig struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Config struct {
		Host        string   `yaml:"host"`
		Port        int      `yaml:"port"`
		Database    string   `yaml:"database"`
		User        string   `yaml:"user"`
		Password    int      `yaml:"password"`
		Path        string   `yaml:"path"`
		Extensions  []string `yaml:"extensions"`
		ExtraParams []struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		} `yaml:"extraParams"`
	} `yaml:"config"`
}
