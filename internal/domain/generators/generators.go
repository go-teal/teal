package generators

type Generator interface {
	RenderToFile() error
	GetFileName() string
	GetFullPath() string
}
