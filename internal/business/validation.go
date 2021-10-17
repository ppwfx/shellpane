package business

import (
	"github.com/pkg/errors"
	"github.com/ppwfx/shellpane/internal/domain"
	"github.com/ppwfx/shellpane/internal/utils/errutil"
	"regexp"
)



func ValidateViewSpecs(specs []domain.ViewSpec) error {
	err := validateViewSpecNames(specs)
	if err != nil {
		return errors.Wrapf(err, "failed to validate spec names")
	}

	for i := range specs {
		err := validateViewSpec(specs[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate spec with name=%v", specs[i].Name)
		}
	}

	return nil
}

func validateViewSpecNames(specs []domain.ViewSpec) error {
	seenNames := map[string]struct{}{}
	for i := range specs {
		if specs[i].Name == "" {
			return errutil.Invalid(errors.Errorf("spec[%v] name is empty", i))
		}

		_, ok := seenNames[specs[i].Name]
		if ok {
			return errutil.Invalid(errors.Errorf("duplicate name=%v", specs[i].Name))
		}

		seenNames[specs[i].Name] = struct{}{}
	}

	return nil
}

func validateViewSpec(spec domain.ViewSpec) error {
	err := validateEnvSpecNames(spec.Env)
	if err != nil {
		return errors.Wrapf(err, "failed to validate env names")
	}

	if spec.Command == "" {
		return errutil.Invalid(errors.New("command is empty"))
	}

	for i := range spec.Env {
		if spec.Env[0].Name == "" {
			return errutil.Invalid(errors.Errorf("env[%v] name is empty", i))
		}
	}

	return nil
}

func validateEnvSpecNames(specs []domain.EnvSpec) error {
	seenNames := map[string]struct{}{}
	for i := range specs {
		if specs[i].Name == "" {
			return errutil.Invalid(errors.Errorf("spec[%v] name is empty", i))
		}

		_, ok := seenNames[specs[i].Name]
		if ok {
			return errutil.Invalid(errors.Errorf("duplicate name=%v", specs[i].Name))
		}

		seenNames[specs[i].Name] = struct{}{}
	}

	return nil
}

func validateGetViewOutputRequest(spec domain.ViewSpec, req GetViewOutputRequest) error {
	if req.Name == "" {
		return errutil.Invalid(errors.New("name is empty"))
	}

	switch req.Format {
	case "", FormatRaw:
		break
	default:
		return errutil.Invalid(errors.Errorf("invalid format=%v", req.Format))
	}

	if req.Name == "" {
		return errutil.Invalid(errors.New("name is empty"))
	}

	validEnvNames := map[string]struct{}{}
	for i := range spec.Env {
		validEnvNames[spec.Env[i].Name] = struct{}{}
	}

	for i := range req.Env {
		_, ok := validEnvNames[req.Env[i].Name]
		if !ok {
			return errutil.Invalid(errors.Errorf("invalid env-name=%v", req.Env[i].Name))
		}
	}


	for i := range spec.Env {
		err := validate(spec.Env[i], req.Env[i])
		if err != nil {
		    return errors.Wrapf(err, "input not valid %v=%v", spec.Env[i], req.Env[i])
		}
	}

	return nil
}

func validate(spec domain.EnvSpec, val EnvValue) error {
	re, err := regexp.Compile(spec.Validator)
	if err != nil {
		return errutil.Invalid(errors.Wrapf(err, "invalid regular expression=%v", spec.Validator))
	}
	ok := re.MatchString(val.Value)
	if !ok {
		return errutil.Invalid(errors.Errorf("invalid env-value %v=%v", val.Name, val.Value))
	}

	return nil
}