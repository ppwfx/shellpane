package bootstrap

import (
	"github.com/pkg/errors"
)

func ValidateShellpaneConfig(config ShellpaneConfig) error {
	err := validateInputs(config.Inputs)
	if err != nil {
		return errors.Wrapf(err, "failed to validate inputs")
	}

	definedInputs := map[string]struct{}{}
	for i := range config.Inputs {
		definedInputs[config.Inputs[i].Slug] = struct{}{}
	}

	err = validateCommands(definedInputs, config.Commands)
	if err != nil {
		return errors.Wrapf(err, "failed to validate commands")
	}

	definedCommands := map[string]struct{}{}
	for i := range config.Commands {
		definedCommands[config.Commands[i].Slug] = struct{}{}
	}

	err = validateSequences(definedCommands, config.Sequences)
	if err != nil {
		return errors.Wrapf(err, "failed to validate sequences")
	}

	definedSequences := map[string]struct{}{}
	for i := range config.Sequences {
		definedSequences[config.Sequences[i].Slug] = struct{}{}
	}

	err = validateCategories(config.Categories)
	if err != nil {
		return errors.Wrapf(err, "failed to validate categories")
	}

	definedCategories := map[string]struct{}{}
	for i := range config.Categories {
		definedCategories[config.Categories[i].Slug] = struct{}{}
	}

	err = validateViews(definedCommands, definedSequences, definedCategories, config.Views)
	if err != nil {
		return errors.Wrapf(err, "failed to validate views")
	}

	definedViews := map[string]struct{}{}
	for i := range config.Views {
		definedViews[config.Views[i].Slug] = struct{}{}
	}

	err = validateRoles(definedCategories, definedViews, config.Roles)
	if err != nil {
		return errors.Wrapf(err, "failed to validate roles")
	}

	definedRoles := map[string]struct{}{}
	for i := range config.Roles {
		definedRoles[config.Roles[i].Slug] = struct{}{}
	}

	err = validateGroups(definedRoles, config.Groups)
	if err != nil {
		return errors.Wrapf(err, "failed to validate groups")
	}

	definedGroups := map[string]struct{}{}
	for i := range config.Groups {
		definedGroups[config.Groups[i].Slug] = struct{}{}
	}

	err = validateUsers(definedGroups, config.Users)
	if err != nil {
		return errors.Wrapf(err, "failed to validate users")
	}

	return nil
}

func validateUsers(definedGroups map[string]struct{}, users []UserConfig) error {
	for i := range users {
		err := validateUser(definedGroups, users[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate user id=%v", users[i].ID)
		}
	}

	seenIDs := map[string]struct{}{}
	for i := range users {
		_, seen := seenIDs[users[i].ID]
		if seen {
			return errors.Errorf("duplicate id=%v", users[i].ID)
		}
		seenIDs[users[i].ID] = struct{}{}
	}

	return nil
}

func validateUser(definedGroups map[string]struct{}, user UserConfig) error {
	if user.ID == "" {
		return errors.New("id is empty")
	}

	for i := range user.Groups {
		_, defined := definedGroups[user.Groups[i].GroupSlug]
		if !defined {
			return errors.Errorf("undefined group=%v", user.Groups[i].GroupSlug)
		}
	}

	return nil
}

func validateGroups(definedRoles map[string]struct{}, groups []GroupConfig) error {
	for i := range groups {
		err := validateGroup(definedRoles, groups[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate group slug=%v", groups[i].Slug)
		}
	}

	seenSlugs := map[string]struct{}{}
	for i := range groups {
		_, seen := seenSlugs[groups[i].Slug]
		if seen {
			return errors.Errorf("duplicate slug=%v", groups[i].Slug)
		}
		seenSlugs[groups[i].Slug] = struct{}{}
	}

	return nil
}

func validateGroup(definedRoles map[string]struct{}, group GroupConfig) error {
	if group.Slug == "" {
		return errors.New("slug is empty")
	}

	for i := range group.Roles {
		_, defined := definedRoles[group.Roles[i].RoleSlug]
		if !defined {
			return errors.Errorf("undefined role=%v", group.Roles[i].RoleSlug)
		}
	}

	return nil
}

func validateRoles(definedCategories map[string]struct{}, definedViews map[string]struct{}, roles []RoleConfig) error {
	for i := range roles {
		err := validateRole(definedCategories, definedViews, roles[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate role slug=%v", roles[i].Slug)
		}
	}

	seenSlugs := map[string]struct{}{}
	for i := range roles {
		_, seen := seenSlugs[roles[i].Slug]
		if seen {
			return errors.Errorf("duplicate slug=%v", roles[i].Slug)
		}
		seenSlugs[roles[i].Slug] = struct{}{}
	}

	return nil
}

func validateRole(definedCategories map[string]struct{}, definedViews map[string]struct{}, role RoleConfig) error {
	if role.Slug == "" {
		return errors.New("slug is empty")
	}

	for i := range role.Views {
		_, defined := definedViews[role.Views[i].ViewSlug]
		if !defined {
			return errors.Errorf("undefined view=%v", role.Views[i].ViewSlug)
		}
	}

	for i := range role.Categories {
		_, defined := definedCategories[role.Categories[i].CategorySlug]
		if !defined {
			return errors.Errorf("undefined category=%v", role.Categories[i].CategorySlug)
		}
	}

	return nil
}

func validateViews(definedCommands map[string]struct{}, definedSequences map[string]struct{}, definedCategories map[string]struct{}, views []ViewConfig) error {
	for i := range views {
		err := validateView(definedCommands, definedSequences, definedCategories, views[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate view name=%v", views[i].Name)
		}
	}

	seenNames := map[string]struct{}{}
	for i := range views {
		_, seen := seenNames[views[i].Name]
		if seen {
			return errors.Errorf("duplicate name=%v", views[i].Name)
		}
		seenNames[views[i].Name] = struct{}{}
	}

	seenSlugs := map[string]struct{}{}
	for i := range views {
		_, seen := seenSlugs[views[i].Slug]
		if seen {
			return errors.Errorf("duplicate slug=%v", views[i].Slug)
		}
		seenSlugs[views[i].Slug] = struct{}{}
	}

	return nil
}

func validateView(definedCommands map[string]struct{}, definedSequences map[string]struct{}, definedCategories map[string]struct{}, view ViewConfig) error {
	if view.Name == "" {
		return errors.New("name is empty")
	}

	if view.Slug == "" {
		return errors.New("slug is empty")
	}

	switch {
	case view.CommandSlug != "" && view.SequenceSlug != "":
		return errors.New("command and process set")
	case view.CommandSlug == "" && view.SequenceSlug == "":
		return errors.New("no command and no process set")
	case view.CommandSlug != "":
		_, defined := definedCommands[view.CommandSlug]
		if !defined {
			return errors.Errorf("undefined command=%v", view.CommandSlug)
		}
	case view.SequenceSlug != "":
		_, defined := definedSequences[view.SequenceSlug]
		if !defined {
			return errors.Errorf("undefined process=%v", view.SequenceSlug)
		}
	}

	_, defined := definedCategories[view.CategorySlug]
	if !defined {
		return errors.Errorf("undefined category=%v", view.CategorySlug)
	}

	return nil
}

func validateCategories(categories []CategoryConfig) error {
	for i := range categories {
		err := validateCategory(categories[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate process slug=%v", categories[i].Slug)
		}
	}

	seenSlugs := map[string]struct{}{}
	for i := range categories {
		_, seen := seenSlugs[categories[i].Slug]
		if seen {
			return errors.Errorf("duplicate slug=%v", categories[i].Slug)
		}
		seenSlugs[categories[i].Slug] = struct{}{}
	}

	return nil
}

func validateCategory(category CategoryConfig) error {
	if category.Slug == "" {
		return errors.New("slug is empty")
	}

	if category.Name == "" {
		return errors.New("name is empty")
	}

	if category.Color == "" {
		return errors.New("color is empty")
	}

	return nil
}

func validateSequences(definedCommands map[string]struct{}, processes []SequenceConfig) error {
	for i := range processes {
		err := validateSequence(definedCommands, processes[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate process slug=%v", processes[i].Slug)
		}
	}

	seenSlugs := map[string]struct{}{}
	for i := range processes {
		_, seen := seenSlugs[processes[i].Slug]
		if seen {
			return errors.Errorf("duplicate slug=%v", processes[i].Slug)
		}
		seenSlugs[processes[i].Slug] = struct{}{}
	}

	return nil
}

func validateSequence(definedCommands map[string]struct{}, process SequenceConfig) error {
	if process.Slug == "" {
		return errors.New("slug is empty")
	}

	if len(process.Steps) == 0 {
		return errors.New("no steps defined")
	}

	err := validateSteps(definedCommands, process.Steps)
	if err != nil {
		return errors.Wrapf(err, "failed to validate steps")
	}

	return nil
}

func validateSteps(definedCommands map[string]struct{}, steps []StepConfig) error {
	for i := range steps {
		err := validateStep(definedCommands, steps[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate step name=%v", steps[i].Name)
		}
	}

	return nil
}

func validateStep(definedCommands map[string]struct{}, step StepConfig) error {
	if step.Name == "" {
		return errors.New("name is empty")
	}

	_, defined := definedCommands[step.CommandSlug]
	if !defined {
		return errors.Errorf("undefined command=%v", step.CommandSlug)
	}

	return nil
}

func validateCommands(definedInputs map[string]struct{}, commands []CommandConfig) error {
	for i := range commands {
		err := validateCommand(definedInputs, commands[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate command slug=%v", commands[i].Slug)
		}
	}

	seenSlugs := map[string]struct{}{}
	for i := range commands {
		_, seen := seenSlugs[commands[i].Slug]
		if seen {
			return errors.Errorf("duplicate slug=%v", commands[i].Slug)
		}
		seenSlugs[commands[i].Slug] = struct{}{}
	}

	return nil
}

func validateCommand(definedInputs map[string]struct{}, command CommandConfig) error {
	if command.Slug == "" {
		return errors.New("slug is empty")
	}

	if command.Command == "" {
		return errors.New("command is empty")
	}

	for i := range command.Inputs {
		_, defined := definedInputs[command.Inputs[i].InputSlug]
		if !defined {
			return errors.Errorf("undefined input slug=%v", command.Inputs[i].InputSlug)
		}
	}

	seenInputs := map[string]struct{}{}
	for i := range command.Inputs {
		_, seen := seenInputs[command.Inputs[i].InputSlug]
		if seen {
			return errors.Errorf("duplicate input slug=%v", command.Inputs[i].InputSlug)
		}
		seenInputs[command.Inputs[i].InputSlug] = struct{}{}
	}

	return nil
}

func validateInputs(inputs []InputConfig) error {
	for i := range inputs {
		err := validateInput(inputs[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate input slug=%v", inputs[i].Slug)
		}
	}

	seenSlugs := map[string]struct{}{}
	for i := range inputs {
		_, seen := seenSlugs[inputs[i].Slug]
		if seen {
			return errors.Errorf("duplicate slug=%v", inputs[i].Slug)
		}
		seenSlugs[inputs[i].Slug] = struct{}{}
	}

	return nil
}

func validateInput(input InputConfig) error {
	if input.Slug == "" {
		return errors.New("slug is empty")
	}

	return nil
}
