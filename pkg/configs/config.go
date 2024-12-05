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
	Config *struct {
		Host        string   `yaml:"host"`
		HostEnv     string   `yaml:"host_env"`
		Port        int      `yaml:"port"`
		PortEnv     string   `yaml:"port_env"`
		Database    string   `yaml:"database"`
		DatabaseEnv string   `yaml:"database_env"`
		User        string   `yaml:"user"`
		UserEnv     string   `yaml:"user_env"`
		Password    string   `yaml:"password"`
		PasswordEnv string   `yaml:"password_env"`
		Path        string   `yaml:"path"`
		PathEnv     string   `yaml:"path_env"`
		Extensions  []string `yaml:"extensions"`
		ExtraParams []*struct {
			Name     string `yaml:"name"`
			Value    string `yaml:"value"`
			ValueEnv string `yaml:"value_env"`
		} `yaml:"extraParams"`
	} `yaml:"config"`
}
