package domain

type UserConfig struct {
	ID     string
	Groups []GroupConfig
}

type GroupConfig struct {
	Slug  string
	Roles []RoleConfig
}

type RoleConfig struct {
	Slug       string
	Views      []ViewConfig
	Categories []CategoryConfig
}

type ViewConfig struct {
	Slug     string
	Name     string
	Execute  ViewExecuteConfig
	Command  CommandConfig
	Sequence SequenceConfig
	Category CategoryConfig
}

type ViewExecuteConfig struct {
	Auto bool
}

type CategoryConfig struct {
	Slug  string
	Name  string
	Color string
	Views []ViewConfig
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
	Slug        string
	Command     string
	Description string
	Inputs      []CommandInputConfig
}

type CommandInputConfig struct {
	Input InputConfig
}

type InputConfig struct {
	Slug        string
	Description string
}
