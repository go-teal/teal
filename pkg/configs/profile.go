package configs

type MatType string

const (
	MAT_TABLE       MatType = "table"
	MAT_VIEW        MatType = "view"
	MAT_INCREMENTAL MatType = "incremental"
	MAT_CUSTOM      MatType = "custom"
)

type ProjectProfile struct {
	Version    string `yaml:"version"`
	Name       string `yaml:"name"`
	Connection string `yaml:"connection"`
	Models     struct {
		Stages []*struct {
			Name   string `yaml:"name"`
			Models []*ModelProfile
		} `yaml:"stages"`
	} `yaml:"models"`
	Sources []SourceProfile `yaml:"sources"`
}

type SourceProfile struct {
	Name       string   `yaml:"name"`
	Connection string   `yaml:"connection"`
	Type       string   `yaml:"type"`
	ReadOnly   bool     `yaml:"read_only"`
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
	IsDataFramed    bool    `yaml:"is_data_framed"`
	PersistInputs   bool    `yaml:"persist_inputs"`
	Stage           string
}

func (mp *ModelProfile) GetTempName() string {
	return "tmp_" + mp.Stage + "_" + mp.Name
}

func (p ProjectProfile) ToMap() map[string]*ModelProfile {
	profilesMap := make(map[string]*ModelProfile)
	for _, s := range p.Models.Stages {
		for _, m := range s.Models {
			profilesMap[s.Name+"."+m.Name] = m
		}
	}
	return profilesMap
}

func (p ProjectProfile) GetModelProfile(stage string, name string) *ModelProfile {
	for _, s := range p.Models.Stages {
		if s.Name == stage {
			for _, m := range s.Models {
				if m.Name == name {
					m.Stage = stage
					return m
				}
			}
		}
	}
	return &ModelProfile{
		Name:            name,
		Stage:           stage,
		Connection:      p.Connection,
		Materialization: MAT_TABLE,
	}
}
