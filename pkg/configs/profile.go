package configs

type MatType string

const (
	MAT_TABLE       MatType = "table"
	MAT_VIEW        MatType = "view"
	MAT_INCREMENTAL MatType = "incremental"
	// From DataFrame to DB
	MAT_UPSTREAM MatType = "upstream"
	// From DB to DataFrame
	MAT_DOWNSTREAM MatType = "downstream"
	MAT_CUSTOM     MatType = "custom"
)

type Profile struct {
	Version    string `yaml:"version"`
	Name       string `yaml:"name"`
	Connection string `yaml:"connection"`
	Models     struct {
		Stages []struct {
			Name   string `yaml:"name"`
			Models []ModelProfile
		} `yaml:"stages"`
	} `yaml:"models"`
	Sources []SourceProfile `yaml:"sources"`
}

type SourceProfile struct {
	Name       string   `yaml:"name"`
	Connection string   `yaml:"connection"`
	Type       string   `yaml:"type"`
	ReadOnly   bool     `yaml:"readOnly"`
	Tables     []string `yaml:"tables"`
	Params     []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	} `yaml:"params"`
}

type ModelProfile struct {
	Name            string  `yaml:"name"`
	Connection      string  `yaml:"connection"`
	Materialization MatType `yaml:"materialization"`
}

func (p Profile) GetModelProfile(stage string, name string) ModelProfile {
	for _, s := range p.Models.Stages {
		if s.Name == stage {
			for _, m := range s.Models {
				if m.Name == name {
					return m
				}
			}
		}
	}
	return ModelProfile{
		Connection:      p.Connection,
		Materialization: MAT_TABLE,
	}
}
