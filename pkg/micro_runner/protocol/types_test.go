package protocol

import (
	"testing"
)

func TestMessageTypeValidate(t *testing.T) {
	tests := []struct {
		name    string
		msgType MessageType
		wantErr bool
	}{
		{"valid READY", MessageTypeReady, false},
		{"valid CMD", MessageTypeCommand, false},
		{"valid EVENT", MessageTypeEvent, false},
		{"valid DONE", MessageTypeDone, false},
		{"valid ERROR", MessageTypeError, false},
		{"valid EXIT", MessageTypeExit, false},
		{"invalid type", MessageType("INVALID"), true},
		{"empty type", MessageType(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msgType.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MessageType.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandTypeValidate(t *testing.T) {
	tests := []struct {
		name    string
		cmdType CommandType
		wantErr bool
	}{
		{"valid exec", CommandTypeExec, false},
		{"valid file.write", CommandTypeFileWrite, false},
		{"valid file.read", CommandTypeFileRead, false},
		{"valid pkg.ensure", CommandTypePkgEnsure, false},
		{"valid service.reload", CommandTypeServiceReload, false},
		{"valid sudoers.ensure", CommandTypeSudoersEnsure, false},
		{"valid sshd.harden", CommandTypeSSHDHarden, false},
		{"invalid type", CommandType("invalid"), true},
		{"empty type", CommandType(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmdType.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CommandType.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandMessageValidate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *CommandMessage
		wantErr bool
	}{
		{
			name: "valid command",
			cmd: &CommandMessage{
				ID:      "cmd-123",
				Type:    CommandTypeExec,
				Timeout: 30,
				Params:  []byte(`{"command":"ls"}`),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			cmd: &CommandMessage{
				Type:    CommandTypeExec,
				Timeout: 30,
				Params:  []byte(`{}`),
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			cmd: &CommandMessage{
				ID:      "cmd-123",
				Type:    CommandType("invalid"),
				Timeout: 30,
				Params:  []byte(`{}`),
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			cmd: &CommandMessage{
				ID:      "cmd-123",
				Type:    CommandTypeExec,
				Timeout: 0,
				Params:  []byte(`{}`),
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			cmd: &CommandMessage{
				ID:      "cmd-123",
				Type:    CommandTypeExec,
				Timeout: -1,
				Params:  []byte(`{}`),
			},
			wantErr: true,
		},
		{
			name: "empty params",
			cmd: &CommandMessage{
				ID:      "cmd-123",
				Type:    CommandTypeExec,
				Timeout: 30,
				Params:  []byte{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CommandMessage.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventMessageValidate(t *testing.T) {
	tests := []struct {
		name    string
		evt     *EventMessage
		wantErr bool
	}{
		{
			name: "valid event",
			evt: &EventMessage{
				CommandID: "cmd-123",
				Level:     "info",
				Message:   "Processing",
			},
			wantErr: false,
		},
		{
			name: "valid event with progress",
			evt: &EventMessage{
				CommandID: "cmd-123",
				Level:     "info",
				Message:   "Downloading",
				Progress: &ProgressInfo{
					Current: 50,
					Total:   100,
					Unit:    "bytes",
				},
			},
			wantErr: false,
		},
		{
			name: "missing command ID",
			evt: &EventMessage{
				Level:   "info",
				Message: "Processing",
			},
			wantErr: true,
		},
		{
			name: "invalid level",
			evt: &EventMessage{
				CommandID: "cmd-123",
				Level:     "invalid",
				Message:   "Processing",
			},
			wantErr: true,
		},
		{
			name: "empty level defaults to info",
			evt: &EventMessage{
				CommandID: "cmd-123",
				Message:   "Processing",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.evt.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("EventMessage.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
