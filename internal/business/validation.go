package business

import (
	"github.com/pkg/errors"

	"github.com/ppwfx/shellpane/internal/domain"
	"github.com/ppwfx/shellpane/internal/utils/errutil"
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
	err := validateSteps(spec.Steps)
	if err != nil {
		return errors.Wrapf(err, "failed to validate steps")
	}

	err = validateEnvSpecNames(spec.Env)
	if err != nil {
		return errors.Wrapf(err, "failed to validate env names")
	}

	for i := range spec.Env {
		if spec.Env[0].Name == "" {
			return errutil.Invalid(errors.Errorf("env[%v] name is empty", i))
		}
	}

	return nil
}

func validateSteps(steps []domain.Step) error {
	for i := range steps {
		err := validateStep(steps[i])
		if err != nil {
			return errors.Wrapf(err, "failed to validate step with name=%v", steps[i].Name)
		}
	}

	return nil
}

func validateStep(step domain.Step) error {
	err := validateEnvSpecNames(step.Env)
	if err != nil {
		return errors.Wrapf(err, "failed to validate env names")
	}

	if step.Command == "" {
		return errutil.Invalid(errors.New("command is empty"))
	}

	for i := range step.Env {
		if step.Env[0].Name == "" {
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

func validateGetStepOutputRequest(view domain.ViewSpec, step domain.Step, req GetStepOutputRequest) error {
	if req.ViewName == "" {
		return errutil.Invalid(errors.New("view name is empty"))
	}

	if req.StepName == "" {
		return errutil.Invalid(errors.New("step name is empty"))
	}

	switch req.Format {
	case "", FormatRaw:
		break
	default:
		return errutil.Invalid(errors.Errorf("invalid format=%v", req.Format))
	}

	validViewEnvNames := map[string]struct{}{}
	for i := range view.Env {
		validViewEnvNames[view.Env[i].Name] = struct{}{}
	}

	for i := range req.ViewEnv {
		_, ok := validViewEnvNames[req.ViewEnv[i].Name]
		if !ok {
			return errutil.Invalid(errors.Errorf("invalid view env name=%v", req.ViewEnv[i].Name))
		}
	}

	validStepEnvNames := map[string]struct{}{}
	for i := range step.Env {
		validStepEnvNames[step.Env[i].Name] = struct{}{}
	}

	for i := range req.StepEnv {
		_, ok := validStepEnvNames[req.StepEnv[i].Name]
		if !ok {
			return errutil.Invalid(errors.Errorf("invalid step env name=%v", req.StepEnv[i].Name))
		}
	}

	return nil
}
