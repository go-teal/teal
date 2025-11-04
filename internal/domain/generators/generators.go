package generators

type Generator interface {
	RenderToFile() (error, bool) // Returns error and skipStatus (true if skipped)
	GetFileName() string
	GetFullPath() string
}
