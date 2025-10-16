package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Encoder writes protocol messages to an io.Writer.
type Encoder struct {
	w *bufio.Writer
}

// NewEncoder creates a new protocol encoder.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: bufio.NewWriter(w),
	}
}

// Encode writes a message to the output stream.
func (e *Encoder) Encode(msgType MessageType, data interface{}) error {
	if err := msgType.Validate(); err != nil {
		return fmt.Errorf("invalid message type: %w", err)
	}

	var dataBytes []byte
	var err error
	if data != nil {
		dataBytes, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
	}

	msg := Message{
		Type:      msgType,
		Timestamp: time.Now().UTC(),
		Data:      dataBytes,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if _, err := e.w.Write(msgBytes); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := e.w.WriteByte('\n'); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	if err := e.w.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	return nil
}

// EncodeReady sends a READY message.
func (e *Encoder) EncodeReady(ready *ReadyMessage) error {
	return e.Encode(MessageTypeReady, ready)
}

// EncodeEvent sends an EVENT message.
func (e *Encoder) EncodeEvent(event *EventMessage) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}
	return e.Encode(MessageTypeEvent, event)
}

// EncodeDone sends a DONE message.
func (e *Encoder) EncodeDone(done *DoneMessage) error {
	return e.Encode(MessageTypeDone, done)
}

// EncodeError sends an ERROR message.
func (e *Encoder) EncodeError(err *ErrorMessage) error {
	return e.Encode(MessageTypeError, err)
}

// EncodeExit sends an EXIT message.
func (e *Encoder) EncodeExit(exit *ExitMessage) error {
	return e.Encode(MessageTypeExit, exit)
}

// Decoder reads protocol messages from an io.Reader.
type Decoder struct {
	r *bufio.Scanner
}

// NewDecoder creates a new protocol decoder.
func NewDecoder(r io.Reader) *Decoder {
	scanner := bufio.NewScanner(r)
	// Set a large buffer for potentially large commands
	const maxCapacity = 10 * 1024 * 1024 // 10 MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	return &Decoder{
		r: scanner,
	}
}

// Decode reads the next message from the input stream.
func (d *Decoder) Decode() (*Message, error) {
	if !d.r.Scan() {
		if err := d.r.Err(); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		return nil, io.EOF
	}

	line := d.r.Bytes()
	if len(line) == 0 {
		return nil, fmt.Errorf("empty line")
	}

	var msg Message
	if err := json.Unmarshal(line, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if err := msg.Type.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	return &msg, nil
}

// DecodeCommand decodes a command message.
func (d *Decoder) DecodeCommand() (*CommandMessage, error) {
	msg, err := d.Decode()
	if err != nil {
		return nil, err
	}

	if msg.Type != MessageTypeCommand {
		return nil, fmt.Errorf("expected CMD message, got %s", msg.Type)
	}

	var cmd CommandMessage
	if err := json.Unmarshal(msg.Data, &cmd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command: %w", err)
	}

	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	return &cmd, nil
}

// ParseParams parses command parameters into a specific type.
func ParseParams(params json.RawMessage, target interface{}) error {
	if err := json.Unmarshal(params, target); err != nil {
		return fmt.Errorf("failed to parse params: %w", err)
	}
	return nil
}
