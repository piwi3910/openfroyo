// Package main implements the OpenFroyo micro-runner binary.
// This is a minimal, self-contained, static binary that executes
// commands received via JSON-over-stdio and self-deletes on exit.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/openfroyo/openfroyo/pkg/micro_runner/handlers"
	"github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

const (
	version = "1.0.0"
	ttl     = 10 * time.Minute
)

type runner struct {
	encoder      *protocol.Encoder
	decoder      *protocol.Decoder
	execPath     string
	commandCount int
	startTime    time.Time
}

func main() {
	r := &runner{
		encoder:   protocol.NewEncoder(os.Stdout),
		decoder:   protocol.NewDecoder(os.Stdin),
		startTime: time.Now(),
	}

	// Get executable path for self-delete
	var err error
	r.execPath, err = os.Executable()
	if err != nil {
		r.sendErrorAndExit("INIT_FAILED", fmt.Sprintf("failed to get executable path: %v", err), 1)
		return
	}

	// Send READY message
	if err := r.sendReady(); err != nil {
		r.sendErrorAndExit("READY_FAILED", fmt.Sprintf("failed to send ready: %v", err), 1)
		return
	}

	// Main command loop with TTL timeout
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()

	exitCode := 0
	reason := "completed"

	for {
		select {
		case <-ctx.Done():
			reason = "ttl_expired"
			exitCode = 0
			goto exit
		default:
			// Try to read next command
			if err := r.processNextCommand(ctx); err != nil {
				if err.Error() == "EOF" {
					reason = "stdin_closed"
					exitCode = 0
				} else {
					reason = "error"
					exitCode = 1
				}
				goto exit
			}
		}
	}

exit:
	r.exit(reason, exitCode)
}

func (r *runner) sendReady() error {
	ready := &protocol.ReadyMessage{
		Version:  version,
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
		PID:      os.Getpid(),
		Caps: map[string]bool{
			"exec":            true,
			"file.write":      true,
			"file.read":       true,
			"pkg.ensure":      true,
			"service.reload":  true,
			"sudoers.ensure":  true,
			"sshd.harden":     true,
		},
		Metadata: map[string]string{
			"ttl": ttl.String(),
		},
	}

	return r.encoder.EncodeReady(ready)
}

func (r *runner) processNextCommand(ctx context.Context) error {
	cmd, err := r.decoder.DecodeCommand()
	if err != nil {
		return err
	}

	r.commandCount++

	// Create command context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(cmd.Timeout)*time.Second)
	defer cancel()

	// Event channel for progress updates
	eventCh := make(chan *protocol.EventMessage, 10)
	defer close(eventCh)

	// Start goroutine to send events
	go func() {
		for evt := range eventCh {
			r.encoder.EncodeEvent(evt)
		}
	}()

	// Execute command
	start := time.Now()
	result, err := r.handleCommand(cmdCtx, cmd, eventCh)
	duration := time.Since(start).Seconds()

	if err != nil {
		// Send error message
		errMsg := &protocol.ErrorMessage{
			CommandID: cmd.ID,
			Code:      "EXEC_FAILED",
			Message:   err.Error(),
			Retryable: false,
		}
		return r.encoder.EncodeError(errMsg)
	}

	// Send done message
	doneMsg := &protocol.DoneMessage{
		CommandID: cmd.ID,
		Result:    result,
		Duration:  duration,
	}
	return r.encoder.EncodeDone(doneMsg)
}

func (r *runner) handleCommand(ctx context.Context, cmd *protocol.CommandMessage, eventCh chan<- *protocol.EventMessage) (json.RawMessage, error) {
	switch cmd.Type {
	case protocol.CommandTypeExec:
		var params protocol.ExecParams
		if err := protocol.ParseParams(cmd.Params, &params); err != nil {
			return nil, err
		}
		handler := &handlers.ExecHandler{}
		result, err := handler.Handle(ctx, &params, eventCh)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)

	case protocol.CommandTypeFileWrite:
		var params protocol.FileWriteParams
		if err := protocol.ParseParams(cmd.Params, &params); err != nil {
			return nil, err
		}
		handler := &handlers.FileWriteHandler{}
		result, err := handler.Handle(ctx, &params, eventCh)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)

	case protocol.CommandTypeFileRead:
		var params protocol.FileReadParams
		if err := protocol.ParseParams(cmd.Params, &params); err != nil {
			return nil, err
		}
		handler := &handlers.FileReadHandler{}
		result, err := handler.Handle(ctx, &params, eventCh)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)

	case protocol.CommandTypePkgEnsure:
		var params protocol.PkgEnsureParams
		if err := protocol.ParseParams(cmd.Params, &params); err != nil {
			return nil, err
		}
		handler := &handlers.PkgEnsureHandler{}
		result, err := handler.Handle(ctx, &params, eventCh)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)

	case protocol.CommandTypeServiceReload:
		var params protocol.ServiceReloadParams
		if err := protocol.ParseParams(cmd.Params, &params); err != nil {
			return nil, err
		}
		handler := &handlers.ServiceReloadHandler{}
		result, err := handler.Handle(ctx, &params, eventCh)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)

	case protocol.CommandTypeSudoersEnsure:
		var params protocol.SudoersEnsureParams
		if err := protocol.ParseParams(cmd.Params, &params); err != nil {
			return nil, err
		}
		handler := &handlers.SudoersEnsureHandler{}
		result, err := handler.Handle(ctx, &params, eventCh)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)

	case protocol.CommandTypeSSHDHarden:
		var params protocol.SSHDHardenParams
		if err := protocol.ParseParams(cmd.Params, &params); err != nil {
			return nil, err
		}
		handler := &handlers.SSHDHardenHandler{}
		result, err := handler.Handle(ctx, &params, eventCh)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)

	default:
		return nil, fmt.Errorf("unsupported command type: %s", cmd.Type)
	}
}

func (r *runner) exit(reason string, exitCode int) {
	// Send EXIT message
	exitMsg := &protocol.ExitMessage{
		Reason:        reason,
		ExitCode:      exitCode,
		CommandsTotal: r.commandCount,
		SelfDeleted:   false,
	}

	// Attempt self-delete
	if err := os.Remove(r.execPath); err == nil {
		exitMsg.SelfDeleted = true
	}

	r.encoder.EncodeExit(exitMsg)
	os.Exit(exitCode)
}

func (r *runner) sendErrorAndExit(code, message string, exitCode int) {
	errMsg := &protocol.ErrorMessage{
		Code:      code,
		Message:   message,
		Retryable: false,
	}
	r.encoder.EncodeError(errMsg)
	os.Exit(exitCode)
}
