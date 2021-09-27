package business

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

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

type GetViewOutputRequest struct {
	Name   string
	Format string
	Env    []EnvValue
}

type GetViewOutputResponse struct {
	errutil.Response
	Output domain.ViewOutput
}

func (h Handler) GetViewOutput(ctx context.Context, req GetViewOutputRequest) (GetViewOutputResponse, error) {
	spec, ok := h.opts.Repository.GetViewSpecByName(req.Name)
	if !ok {
		return GetViewOutputResponse{}, errors.Wrapf(errutil.NotFound(errutil.Nil(), "ViewSpec", req.Name), "failed to find ViewSpec by name=%v", req.Name)
	}

	err := validateGetViewOutputRequest(spec, req)
	if err != nil {
		return GetViewOutputResponse{}, errors.Wrapf(err, "failed to validate request")
	}

	v, err := generateViewOutput(ctx, spec, req.Env)
	if err != nil {
		return GetViewOutputResponse{}, errors.Wrapf(err, "failed to generate view")
	}

	return GetViewOutputResponse{Output: v}, nil
}

func generateViewOutput(ctx context.Context, s domain.ViewSpec, env []EnvValue) (domain.ViewOutput, error) {
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

	o := domain.ViewOutput{}

	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err != nil && errors.As(err, &exitErr):
		stat, ok := exitErr.Sys().(syscall.WaitStatus)
		if !ok {
			return domain.ViewOutput{}, errors.Wrapf(errutil.Unknown(errors.Errorf("can't cast exit error=%v to syscall.WaitStatus", exitErr)), "failed to get exit code")
		}

		o.ExitCode = stat.ExitStatus()
	case err != nil:
		return domain.ViewOutput{}, errors.Wrapf(errutil.Unknown(err), "failed to run command")
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
