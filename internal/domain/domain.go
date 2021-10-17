package domain

type ViewSpec struct {
	Name  string
	Env   []EnvSpec
	Steps []Step
}

type Step struct {
	Name    string
	Command string
	Env     []EnvSpec
}

type StepOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type EnvSpec struct {
	Name string
}
