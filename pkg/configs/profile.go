package configs

type MatType string

const (
	MAT_TABLE       MatType = "table"
	MAT_VIEW        MatType = "view"
	MAT_INCREMENTAL MatType = "incremental"
	MAT_CUSTOM      MatType = "custom"
	MAT_RAW         MatType = "raw"
)

type ProjectProfile struct {
	Version    string `yaml:"version"`
	Name       string `yaml:"name"`
	Connection string `yaml:"connection"`
	Models     struct {
		Stages []*struct {
			Name   string          `yaml:"name"`
			Models []*ModelProfile `yaml:"models"`
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
	Name             string         `yaml:"name"`
	Description      string         `yaml:"description"`
	Connection       string         `yaml:"connection"`
	Materialization  MatType        `yaml:"materialization"`
	PrimaryKeyFields []string       `yaml:"primary_key_fields"`
	Indexes          []*DBIndex     `yaml:"indexes"`
	IsDataFramed     bool           `yaml:"is_data_framed"`
	PersistInputs    bool           `yaml:"persist_inputs"`
	Stage            string         `yaml:"-"`
	Tests            []*TestProfile `yaml:"tests"`
	RawUpstreams     []string       `yaml:"raw_upstreams"`
}

type DBIndex struct {
	Name   string   `yaml:"name"`
	Unique bool     `yaml:"unique"`
	Fields []string `yaml:"fields"`
}

type TestProfile struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Connection  string `yaml:"connection"`
	Stage       string `yaml:"-"`
}

func (mp *ModelProfile) GetTempName() string {
	return "tmp_" + mp.Stage + "_" + mp.Name
}

func (p ProjectProfile) ToMap() map[string]*ModelProfile {
	profilesMap := make(map[string]*ModelProfile)
	for _, s := range p.Models.Stages {
		for _, m := range s.Models {
			if m.Stage == "" {
				m.Stage = s.Name
			}
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

func (p ProjectProfile) GetTestProfile(stage string, name string) *TestProfile {
	for _, s := range p.Models.Stages {
		if s.Name == stage {
			for _, m := range s.Models {
				for _, testProfile := range m.Tests {
					if testProfile.Name == name {
						testProfile.Stage = stage
						return testProfile
					}
				}
			}
		}
	}
	return &TestProfile{
		Name:       name,
		Stage:      stage,
		Connection: p.Connection,
	}
}
