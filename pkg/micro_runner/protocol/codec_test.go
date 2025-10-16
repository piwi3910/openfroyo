package protocol

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestEncoder(t *testing.T) {
	tests := []struct {
		name    string
		msgType MessageType
		data    interface{}
		wantErr bool
	}{
		{
			name:    "encode ready message",
			msgType: MessageTypeReady,
			data: &ReadyMessage{
				Version:  "1.0.0",
				Platform: "linux",
				Arch:     "amd64",
				PID:      1234,
				Caps:     map[string]bool{"exec": true},
			},
			wantErr: false,
		},
		{
			name:    "encode event message",
			msgType: MessageTypeEvent,
			data: &EventMessage{
				CommandID: "cmd-123",
				Level:     "info",
				Message:   "Processing...",
			},
			wantErr: false,
		},
		{
			name:    "encode done message",
			msgType: MessageTypeDone,
			data: &DoneMessage{
				CommandID: "cmd-123",
				Duration:  1.5,
			},
			wantErr: false,
		},
		{
			name:    "encode error message",
			msgType: MessageTypeError,
			data: &ErrorMessage{
				CommandID: "cmd-123",
				Code:      "EXEC_FAILED",
				Message:   "Command execution failed",
				Retryable: false,
			},
			wantErr: false,
		},
		{
			name:    "encode exit message",
			msgType: MessageTypeExit,
			data: &ExitMessage{
				Reason:        "completed",
				ExitCode:      0,
				SelfDeleted:   true,
				CommandsTotal: 5,
			},
			wantErr: false,
		},
		{
			name:    "invalid message type",
			msgType: MessageType("INVALID"),
			data:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)

			err := enc.Encode(tt.msgType, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify output is valid JSON
				line := strings.TrimSpace(buf.String())
				var msg Message
				if err := json.Unmarshal([]byte(line), &msg); err != nil {
					t.Errorf("Output is not valid JSON: %v", err)
				}
				if msg.Type != tt.msgType {
					t.Errorf("Message type = %v, want %v", msg.Type, tt.msgType)
				}
			}
		})
	}
}

func TestDecoder(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		msgType MessageType
	}{
		{
			name:    "decode ready message",
			input:   `{"type":"READY","timestamp":"2024-01-01T00:00:00Z","data":{"version":"1.0.0","platform":"linux","arch":"amd64","pid":1234,"capabilities":{"exec":true}}}`,
			wantErr: false,
			msgType: MessageTypeReady,
		},
		{
			name:    "decode command message",
			input:   `{"type":"CMD","timestamp":"2024-01-01T00:00:00Z","data":{"id":"cmd-123","type":"exec","timeout":30,"params":{"command":"ls"}}}`,
			wantErr: false,
			msgType: MessageTypeCommand,
		},
		{
			name:    "decode event message",
			input:   `{"type":"EVENT","timestamp":"2024-01-01T00:00:00Z","data":{"command_id":"cmd-123","level":"info","message":"Processing"}}`,
			wantErr: false,
			msgType: MessageTypeEvent,
		},
		{
			name:    "invalid json",
			input:   `{invalid json`,
			wantErr: true,
		},
		{
			name:    "empty line",
			input:   ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := NewDecoder(strings.NewReader(tt.input + "\n"))
			msg, err := dec.Decode()

			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if msg.Type != tt.msgType {
					t.Errorf("Message type = %v, want %v", msg.Type, tt.msgType)
				}
			}
		})
	}
}

func TestDecodeCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		cmdType CommandType
	}{
		{
			name:    "valid exec command",
			input:   `{"type":"CMD","timestamp":"2024-01-01T00:00:00Z","data":{"id":"cmd-123","type":"exec","timeout":30,"params":{"command":"ls"}}}`,
			wantErr: false,
			cmdType: CommandTypeExec,
		},
		{
			name:    "valid file.write command",
			input:   `{"type":"CMD","timestamp":"2024-01-01T00:00:00Z","data":{"id":"cmd-124","type":"file.write","timeout":10,"params":{"path":"/tmp/test","content":"test","create":true}}}`,
			wantErr: false,
			cmdType: CommandTypeFileWrite,
		},
		{
			name:    "wrong message type",
			input:   `{"type":"EVENT","timestamp":"2024-01-01T00:00:00Z","data":{}}`,
			wantErr: true,
		},
		{
			name:    "missing command id",
			input:   `{"type":"CMD","timestamp":"2024-01-01T00:00:00Z","data":{"type":"exec","timeout":30,"params":{}}}`,
			wantErr: true,
		},
		{
			name:    "invalid timeout",
			input:   `{"type":"CMD","timestamp":"2024-01-01T00:00:00Z","data":{"id":"cmd-123","type":"exec","timeout":0,"params":{}}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := NewDecoder(strings.NewReader(tt.input + "\n"))
			cmd, err := dec.DecodeCommand()

			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cmd.Type != tt.cmdType {
					t.Errorf("Command type = %v, want %v", cmd.Type, tt.cmdType)
				}
			}
		})
	}
}

func TestParseParams(t *testing.T) {
	tests := []struct {
		name    string
		params  string
		target  interface{}
		wantErr bool
	}{
		{
			name:    "parse exec params",
			params:  `{"command":"ls","args":["-la"],"capture_out":true,"capture_err":true}`,
			target:  &ExecParams{},
			wantErr: false,
		},
		{
			name:    "parse file write params",
			params:  `{"path":"/tmp/test","content":"hello","create":true}`,
			target:  &FileWriteParams{},
			wantErr: false,
		},
		{
			name:    "parse pkg ensure params",
			params:  `{"name":"nginx","state":"present"}`,
			target:  &PkgEnsureParams{},
			wantErr: false,
		},
		{
			name:    "invalid json",
			params:  `{invalid}`,
			target:  &ExecParams{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseParams(json.RawMessage(tt.params), tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
