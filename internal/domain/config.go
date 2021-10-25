package domain

type ViewConfig struct {
	Name     string
	Command  CommandConfig
	Sequence SequenceConfig
	Category CategoryConfig
}

type CategoryConfig struct {
	Slug  string
	Name  string
	Color string
}

type SequenceConfig struct {
	Slug  string
	Steps []StepConfig
}

type StepConfig struct {
	Name    string
	Command CommandConfig
}

type CommandConfig struct {
	Slug    string
	Command string
	Inputs  []CommandInputConfig
}

type CommandInputConfig struct {
	Input InputConfig
}

type InputConfig struct {
	Slug string
}
