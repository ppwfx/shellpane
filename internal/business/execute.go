package business

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"

	"github.com/ppwfx/shellpane/internal/utils/errutil"
)

const (
	FormatRaw = "raw"
)

type ExecuteCommandRequest struct {
	Slug   string
	Inputs []InputValue
	Format string
}

type InputValue struct {
	Name  string
	Value string
}

type ExecuteCommandResponse struct {
	errutil.Response
	Output CommandOutput
}

type CommandOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func (h Handler) ExecuteCommand(ctx context.Context, req ExecuteCommandRequest) (ExecuteCommandResponse, error) {
	userID := UserID(ctx)
	if userID != "" {
		allowedCommands := h.opts.Repository.GetUserAllowedCommands()[userID]

		_, ok := allowedCommands[req.Slug]
		if !ok {
			return ExecuteCommandResponse{}, errutil.Unauthorized(errors.Errorf("user is not allowed to execute command user=%v command=%v", userID, req.Slug))
		}
	}

	command, ok := h.opts.Repository.GetCommandConfig(req.Slug)
	if !ok {
		return ExecuteCommandResponse{}, errors.Wrapf(errutil.NotFound(errutil.Nil(), "Command", req.Slug), "failed to find command slug=%v", req.Slug)
	}

	//err := validateExecuteCommandRequest(command, req)
	//if err != nil {
	//	return ExecuteCommandResponse{}, errors.Wrapf(err, "failed to validate request")
	//}

	o, err := executeCommand(ctx, command.Command, req.Inputs)
	if err != nil {
		return ExecuteCommandResponse{}, errors.Wrapf(err, "failed to execute command")
	}

	return ExecuteCommandResponse{Output: o}, nil
}

func executeCommand(ctx context.Context, command string, env []InputValue) (CommandOutput, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
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

	err := cmd.Run()
	var exitErr *exec.ExitError
	var exitCode int
	switch {
	case err != nil && errors.As(err, &exitErr):
		stat, ok := exitErr.Sys().(syscall.WaitStatus)
		if !ok {
			return CommandOutput{}, errors.Wrapf(errutil.Unknown(errors.Errorf("can't cast exit error=%v to syscall.WaitStatus", exitErr)), "failed to get exit code")
		}

		exitCode = stat.ExitStatus()
	case err != nil:
		return CommandOutput{}, errors.Wrapf(errutil.Unknown(err), "failed to run command")
	}

	o := CommandOutput{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}

	return o, nil
}
