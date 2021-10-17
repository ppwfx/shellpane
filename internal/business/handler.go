package business

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	"github.com/ppwfx/shellpane/internal/domain"
	"github.com/ppwfx/shellpane/internal/persistence"
	"github.com/ppwfx/shellpane/internal/utils/errutil"
)

type HandlerConfig struct {
}

type HandlerOpts struct {
	Config     HandlerConfig
	Repository persistence.Repository
}

type Handler struct {
	opts HandlerOpts
}

func NewHandler(opts HandlerOpts) Handler {
	return Handler{
		opts: opts,
	}
}

type EnvValue struct {
	Name  string
	Value string
}

const (
	FormatRaw = "raw"
)

type GetStepOutputRequest struct {
	ViewName string
	ViewEnv  []EnvValue
	StepName string
	StepEnv  []EnvValue
	Format   string
}

type GetStepOutputResponse struct {
	errutil.Response
	Output domain.StepOutput
}

func (h Handler) GetStepOutput(ctx context.Context, req GetStepOutputRequest) (GetStepOutputResponse, error) {
	view, ok := h.opts.Repository.GetViewSpec(req.ViewName)
	if !ok {
		return GetStepOutputResponse{}, errors.Wrapf(errutil.NotFound(errutil.Nil(), "View", req.ViewName), "failed to find View with view name=%v", req.ViewName)
	}

	step, ok := getStep(view, req.StepName)
	if !ok {
		return GetStepOutputResponse{}, errors.Wrapf(errutil.NotFound(errutil.Nil(), "Step", req.StepName), "failed to find Step with view name=%v and step name=%v", req.ViewName, req.StepName)
	}

	err := validateGetStepOutputRequest(view, step, req)
	if err != nil {
		return GetStepOutputResponse{}, errors.Wrapf(err, "failed to validate request")
	}

	env := append(req.ViewEnv, req.StepEnv...)
	spew.Dump(env)

	v, err := generateStepOutput(ctx, step, env)
	if err != nil {
		return GetStepOutputResponse{}, errors.Wrapf(err, "failed to generate view")
	}

	return GetStepOutputResponse{Output: v}, nil
}

func getStep(viewSpec domain.ViewSpec, name string) (domain.Step, bool) {
	for i := range viewSpec.Steps {
		if viewSpec.Steps[i].Name == name {
			return viewSpec.Steps[i], true
		}
	}

	return domain.Step{}, false
}

func generateStepOutput(ctx context.Context, s domain.Step, env []EnvValue) (domain.StepOutput, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", s.Command)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	var envStrings []string
	for _, e := range env {
		envStrings = append(envStrings, fmt.Sprintf("%v=%v", e.Name, e.Value))
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, envStrings...)

	o := domain.StepOutput{}

	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err != nil && errors.As(err, &exitErr):
		stat, ok := exitErr.Sys().(syscall.WaitStatus)
		if !ok {
			return domain.StepOutput{}, errors.Wrapf(errutil.Unknown(errors.Errorf("can't cast exit error=%v to syscall.WaitStatus", exitErr)), "failed to get exit code")
		}

		o.ExitCode = stat.ExitStatus()
	case err != nil:
		return domain.StepOutput{}, errors.Wrapf(errutil.Unknown(err), "failed to run command")
	}

	o.Stdout = stdout.String()
	o.Stderr = stderr.String()

	return o, nil
}

type GetViewSpecsRequest struct {
}

type GetViewSpecsResponse struct {
	errutil.Response
	Specs []domain.ViewSpec
}

func (h Handler) GetViewSpecs(ctx context.Context, req GetViewSpecsRequest) (GetViewSpecsResponse, error) {
	return GetViewSpecsResponse{Specs: h.opts.Repository.GetViewSpecs()}, nil
}
