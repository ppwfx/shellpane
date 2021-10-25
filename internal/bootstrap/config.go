package bootstrap

import "github.com/ppwfx/shellpane/internal/domain"

type ShellpaneConfig struct {
	Categories []CategoryConfig
	Views      []ViewConfig
	Sequences  []SequenceConfig
	Commands   []CommandConfig
	Inputs     []InputConfig
}

type CategoryConfig struct {
	Slug  string
	Name  string
	Color string
}

type ViewConfig struct {
	Name         string
	CommandSlug  string `yaml:"command"`
	SequenceSlug string `yaml:"sequence"`
	CategorySlug string `yaml:"category"`
}

type SequenceConfig struct {
	Slug  string
	Steps []StepConfig
}

type StepConfig struct {
	Name        string
	CommandSlug string `yaml:"command"`
}

type CommandConfig struct {
	Slug    string
	Command string
	Inputs  []CommandInputConfig
}

type CommandInputConfig struct {
	InputSlug string `yaml:"input"`
}

type InputConfig struct {
	Slug string
}

func generateConfigs(conf ShellpaneConfig) ([]domain.ViewConfig, []domain.CategoryConfig, map[string]domain.CommandConfig) {
	inputsM := map[string]domain.InputConfig{}
	for _, i := range conf.Inputs {
		inputsM[i.Slug] = domain.InputConfig{
			Slug: i.Slug,
		}
	}

	commandsM := map[string]domain.CommandConfig{}
	for _, c := range conf.Commands {
		var commandInputs []domain.CommandInputConfig
		for _, i := range c.Inputs {
			commandInputs = append(commandInputs, domain.CommandInputConfig{
				Input: inputsM[i.InputSlug],
			})
		}

		commandsM[c.Slug] = domain.CommandConfig{
			Slug:    c.Slug,
			Command: c.Command,
			Inputs:  commandInputs,
		}
	}

	processesM := map[string]domain.SequenceConfig{}
	for _, p := range conf.Sequences {
		var steps []domain.StepConfig
		for _, s := range p.Steps {
			steps = append(steps, domain.StepConfig{
				Name:    s.Name,
				Command: commandsM[s.CommandSlug],
			})
		}

		processesM[p.Slug] = domain.SequenceConfig{
			Slug:  p.Slug,
			Steps: steps,
		}
	}

	categories := []domain.CategoryConfig{}
	categoriesM := map[string]domain.CategoryConfig{}
	for _, c := range conf.Categories {
		category := domain.CategoryConfig{
			Slug:  c.Slug,
			Name:  c.Name,
			Color: c.Color,
		}

		categoriesM[c.Slug] = category
		categories = append(categories, category)
	}

	views := []domain.ViewConfig{}
	for _, v := range conf.Views {
		views = append(views, domain.ViewConfig{
			Name:     v.Name,
			Command:  commandsM[v.CommandSlug],
			Sequence: processesM[v.SequenceSlug],
			Category: categoriesM[v.CategorySlug],
		})
	}

	return views, categories, commandsM
}
