// Package handlers implements command handlers for the micro-runner.
package handlers

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

// ExecHandler handles shell command execution.
type ExecHandler struct{}

// Handle executes a shell command.
func (h *ExecHandler) Handle(ctx context.Context, params *protocol.ExecParams, eventCh chan<- *protocol.EventMessage) (*protocol.ExecResult, error) {
	if params.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	shell := params.Shell
	if shell == "" {
		shell = "/bin/sh"
	}

	// Build command
	var cmd *exec.Cmd

	// Handle sudo if requested
	if params.UseSudo {
		if params.SudoPassword != "" {
			// Use sudo with password via stdin
			if len(params.Args) > 0 {
				// sudo -S command args
				fullCmd := append([]string{"-S", params.Command}, params.Args...)
				cmd = exec.CommandContext(ctx, "sudo", fullCmd...)
			} else {
				// sudo -S sh -c "command"
				cmd = exec.CommandContext(ctx, "sudo", "-S", shell, "-c", params.Command)
			}
			// Set up stdin to provide password
			cmd.Stdin = bytes.NewBufferString(params.SudoPassword + "\n")
		} else {
			// NOPASSWD sudo
			if len(params.Args) > 0 {
				fullCmd := append([]string{params.Command}, params.Args...)
				cmd = exec.CommandContext(ctx, "sudo", fullCmd...)
			} else {
				cmd = exec.CommandContext(ctx, "sudo", shell, "-c", params.Command)
			}
		}
	} else {
		// Normal execution without sudo
		if len(params.Args) > 0 {
			cmd = exec.CommandContext(ctx, params.Command, params.Args...)
		} else {
			// If no args, run command through shell
			cmd = exec.CommandContext(ctx, shell, "-c", params.Command)
		}
	}

	// Set working directory
	if params.WorkDir != "" {
		cmd.Dir = params.WorkDir
	}

	// Set environment variables
	if len(params.Env) > 0 {
		env := make([]string, 0, len(params.Env))
		for k, v := range params.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	// Setup output capture
	var stdout, stderr bytes.Buffer
	if params.CaptureOut {
		cmd.Stdout = &stdout
	}
	if params.CaptureErr {
		cmd.Stderr = &stderr
	}

	// Execute command
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start).Seconds()

	result := &protocol.ExecResult{
		Duration: duration,
	}

	if params.CaptureOut {
		result.Stdout = stdout.String()
	}
	if params.CaptureErr {
		result.Stderr = stderr.String()
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute command: %w", err)
		}
	} else {
		result.ExitCode = 0
	}

	return result, nil
}
