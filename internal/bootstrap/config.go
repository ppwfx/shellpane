package bootstrap

import (
	"github.com/ppwfx/shellpane/internal/domain"
)

type ShellpaneConfig struct {
	Users      []UserConfig     `yaml:"users,omitempty"`
	Groups     []GroupConfig    `yaml:"groups,omitempty"`
	Roles      []RoleConfig     `yaml:"roles,omitempty"`
	Categories []CategoryConfig `yaml:"categories,omitempty"`
	Views      []ViewConfig     `yaml:"views,omitempty"`
	Sequences  []SequenceConfig `yaml:"sequences,omitempty"`
	Commands   []CommandConfig  `yaml:"commands,omitempty"`
	Inputs     []InputConfig    `yaml:"inputs,omitempty"`
}

type UserConfig struct {
	ID     string            `yaml:"id,omitempty"`
	Groups []UserGroupConfig `yaml:"groups,omitempty"`
}

type UserGroupConfig struct {
	GroupSlug string `yaml:"group,omitempty"`
}

type GroupConfig struct {
	Slug  string            `yaml:"slug,omitempty"`
	Roles []GroupRoleConfig `yaml:"roles,omitempty"`
}

type GroupRoleConfig struct {
	RoleSlug string `yaml:"role,omitempty"`
}

type RoleConfig struct {
	Slug       string               `yaml:"slug,omitempty"`
	Views      []RoleViewConfig     `yaml:"views,omitempty"`
	Categories []RoleCategoryConfig `yaml:"categories,omitempty"`
}

type RoleViewConfig struct {
	ViewSlug string `yaml:"view,omitempty"`
}

type RoleCategoryConfig struct {
	CategorySlug string `yaml:"category"`
}

type CategoryConfig struct {
	Slug  string `yaml:"slug,omitempty"`
	Name  string `yaml:"name,omitempty"`
	Color string `yaml:"color,omitempty"`
}

type ViewConfig struct {
	Slug         string            `yaml:"slug,omitempty"`
	Name         string            `yaml:"name,omitempty"`
	CommandSlug  string            `yaml:"command,omitempty"`
	SequenceSlug string            `yaml:"sequence,omitempty"`
	CategorySlug string            `yaml:"category,omitempty"`
	Execute      ViewExecuteConfig `yaml:"execute,omitempty"`
}

type ViewExecuteConfig struct {
	Auto bool `yaml:"auto,omitempty"`
}

type SequenceConfig struct {
	Slug  string       `yaml:"slug,omitempty"`
	Steps []StepConfig `yaml:"steps,omitempty"`
}

type StepConfig struct {
	Name        string `yaml:"name,omitempty"`
	CommandSlug string `yaml:"command,omitempty"`
}

type CommandConfig struct {
	Slug        string               `yaml:"slug,omitempty"`
	Command     string               `yaml:"command,omitempty"`
	Description string               `yaml:"description,omitempty"`
	Inputs      []CommandInputConfig `yaml:"inputs,omitempty"`
	Display     string               `yaml:"display,omitempty"`
}

type CommandInputConfig struct {
	InputSlug string `yaml:"input,omitempty"`
}

type InputConfig struct {
	Slug        string `yaml:"slug,omitempty"`
	Description string `yaml:"description,omitempty"`
}

func generateConfigs(conf ShellpaneConfig) (
	map[string]domain.UserConfig,
	[]domain.ViewConfig,
	[]domain.CategoryConfig,
	map[string]domain.CommandConfig,
	map[string]map[string]struct{},
	map[string]map[string]struct{},
	map[string]map[string]struct{},
) {
	inputsM := map[string]domain.InputConfig{}
	for _, i := range conf.Inputs {
		inputsM[i.Slug] = domain.InputConfig{
			Slug:        i.Slug,
			Description: i.Description,
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
			Slug:        c.Slug,
			Command:     c.Command,
			Description: c.Description,
			Inputs:      commandInputs,
			Display:     c.Display,
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

	categoriesM := map[string]domain.CategoryConfig{}
	for _, c := range conf.Categories {
		category := domain.CategoryConfig{
			Slug:  c.Slug,
			Name:  c.Name,
			Color: c.Color,
		}

		categoriesM[c.Slug] = category
	}

	views := []domain.ViewConfig{}
	viewsM := map[string]domain.ViewConfig{}
	for _, v := range conf.Views {
		view := domain.ViewConfig{
			Slug: v.Slug,
			Name: v.Name,
			Execute: domain.ViewExecuteConfig{
				Auto: v.Execute.Auto,
			},
			Command:  commandsM[v.CommandSlug],
			Sequence: processesM[v.SequenceSlug],
			Category: categoriesM[v.CategorySlug],
		}

		c := categoriesM[v.CategorySlug]
		c.Views = append(categoriesM[v.CategorySlug].Views, view)
		categoriesM[v.CategorySlug] = c

		viewsM[v.Slug] = view
		views = append(views, view)
	}

	categories := []domain.CategoryConfig{}
	for _, c := range conf.Categories {
		categories = append(categories, categoriesM[c.Slug])
	}

	rolesM := map[string]domain.RoleConfig{}
	for _, r := range conf.Roles {
		var views []domain.ViewConfig
		for _, v := range r.Views {
			views = append(views, viewsM[v.ViewSlug])
		}

		var categories []domain.CategoryConfig
		for _, c := range r.Categories {
			categories = append(categories, categoriesM[c.CategorySlug])
		}

		rolesM[r.Slug] = domain.RoleConfig{
			Slug:       r.Slug,
			Views:      views,
			Categories: categories,
		}
	}

	groupsM := map[string]domain.GroupConfig{}
	for _, r := range conf.Groups {
		var roles []domain.RoleConfig
		for _, r := range r.Roles {
			roles = append(roles, rolesM[r.RoleSlug])
		}

		groupsM[r.Slug] = domain.GroupConfig{
			Slug:  r.Slug,
			Roles: roles,
		}
	}

	usersM := map[string]domain.UserConfig{}
	for _, u := range conf.Users {
		var groups []domain.GroupConfig
		for _, g := range u.Groups {
			groups = append(groups, groupsM[g.GroupSlug])
		}

		usersM[u.ID] = domain.UserConfig{
			ID:     u.ID,
			Groups: groups,
		}
	}

	allowedCommands := map[string]map[string]struct{}{}
	for userID, user := range usersM {
		allowedCommands[userID] = map[string]struct{}{}
		for i := range user.Groups {
			for ii := range user.Groups[i].Roles {
				for _, v := range user.Groups[i].Roles[ii].Views {
					allowedCommands[userID][v.Command.Slug] = struct{}{}

					for _, s := range v.Sequence.Steps {
						allowedCommands[userID][s.Command.Slug] = struct{}{}
					}
				}

				for _, c := range user.Groups[i].Roles[ii].Categories {
					for _, v := range c.Views {
						allowedCommands[userID][v.Command.Slug] = struct{}{}

						for _, s := range v.Sequence.Steps {
							allowedCommands[userID][s.Command.Slug] = struct{}{}
						}
					}
				}
			}
		}
	}

	allowedViews := map[string]map[string]struct{}{}
	for userID, user := range usersM {
		allowedViews[userID] = map[string]struct{}{}
		for i := range user.Groups {
			for ii := range user.Groups[i].Roles {
				for _, v := range user.Groups[i].Roles[ii].Views {
					allowedViews[userID][v.Slug] = struct{}{}
				}

				for _, c := range user.Groups[i].Roles[ii].Categories {
					for _, v := range c.Views {
						allowedViews[userID][v.Slug] = struct{}{}
					}
				}
			}
		}
	}

	allowedCategories := map[string]map[string]struct{}{}
	for userID, user := range usersM {
		allowedCategories[userID] = map[string]struct{}{}
		for i := range user.Groups {
			for ii := range user.Groups[i].Roles {
				for _, c := range user.Groups[i].Roles[ii].Categories {
					allowedCategories[userID][c.Slug] = struct{}{}
				}

				for _, v := range user.Groups[i].Roles[ii].Views {
					allowedCategories[userID][v.Category.Slug] = struct{}{}
				}
			}
		}
	}

	return usersM, views, categories, commandsM, allowedCategories, allowedViews, allowedCommands
}
