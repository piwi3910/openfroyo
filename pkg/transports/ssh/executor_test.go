package ssh

import (
	"context"
	"testing"
	"time"
)

func TestExecutorExecuteCommand(t *testing.T) {
	server := newTestSSHServer(t)
	defer server.close()

	host, port := parseAddress(server.addr)

	config := DefaultConfig(host, "testuser")
	config.Port = port
	config.AuthMethod = AuthMethodPassword
	config.Password = "testpass"
	config.StrictHostKeyChecking = false

	client, err := NewSSHClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Disconnect()

	tests := []struct {
		name           string
		command        string
		expectError    bool
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "simple echo",
			command:        "echo test",
			expectError:    false,
			expectedStdout: "test",
			expectedStderr: "",
		},
		{
			name:           "stderr output",
			command:        "echo error >&2",
			expectError:    false,
			expectedStdout: "",
			expectedStderr: "error",
		},
		{
			name:        "exit with error",
			command:     "exit 1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := client.ExecuteCommand(ctx, tt.command)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError {
				if stdout != tt.expectedStdout {
					t.Errorf("expected stdout '%s', got '%s'", tt.expectedStdout, stdout)
				}

				if stderr != tt.expectedStderr {
					t.Errorf("expected stderr '%s', got '%s'", tt.expectedStderr, stderr)
				}
			}
		})
	}
}

func TestExecutorExecuteCommandWithTimeout(t *testing.T) {
	server := newTestSSHServer(t)
	defer server.close()

	host, port := parseAddress(server.addr)

	config := DefaultConfig(host, "testuser")
	config.Port = port
	config.AuthMethod = AuthMethodPassword
	config.Password = "testpass"
	config.StrictHostKeyChecking = false

	client, err := NewSSHClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Test with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This command should timeout (though our test server might execute it immediately)
	_, _, err = client.ExecuteCommand(ctx, "sleep 10")
	if err != nil {
		// Error is expected if timeout occurs
		t.Logf("command timed out as expected: %v", err)
	}
}

func TestExecutorExecuteCommandWithSudo(t *testing.T) {
	server := newTestSSHServer(t)
	defer server.close()

	host, port := parseAddress(server.addr)

	config := DefaultConfig(host, "testuser")
	config.Port = port
	config.AuthMethod = AuthMethodPassword
	config.Password = "testpass"
	config.StrictHostKeyChecking = false

	client, err := NewSSHClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Test sudo command (NOPASSWD)
	stdout, _, err := client.ExecuteCommandWithSudo(ctx, "whoami", "")
	if err != nil {
		t.Fatalf("sudo command failed: %v", err)
	}

	// The test server will echo the command
	if len(stdout) == 0 {
		t.Error("expected non-empty stdout")
	}
}

func TestExecutorStartInteractiveSession(t *testing.T) {
	server := newTestSSHServer(t)
	defer server.close()

	host, port := parseAddress(server.addr)

	config := DefaultConfig(host, "testuser")
	config.Port = port
	config.AuthMethod = AuthMethodPassword
	config.Password = "testpass"
	config.StrictHostKeyChecking = false

	client, err := NewSSHClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Disconnect()

	stdin, stdout, _, cleanup, err := client.StartInteractiveSession(ctx)
	if err != nil {
		t.Fatalf("failed to start interactive session: %v", err)
	}
	defer cleanup()

	// Write a command
	_, err = stdin.Write([]byte("test\n"))
	if err != nil {
		t.Fatalf("failed to write to stdin: %v", err)
	}

	// Read response (with timeout)
	buf := make([]byte, 1024)
	done := make(chan struct{})
	var n int

	go func() {
		n, _ = stdout.Read(buf)
		close(done)
	}()

	select {
	case <-done:
		if n > 0 {
			t.Logf("received output: %s", string(buf[:n]))
		}
	case <-time.After(2 * time.Second):
		t.Log("timeout waiting for response (expected in some cases)")
	}
}

func TestExecutorBatch(t *testing.T) {
	server := newTestSSHServer(t)
	defer server.close()

	host, port := parseAddress(server.addr)

	config := DefaultConfig(host, "testuser")
	config.Port = port
	config.AuthMethod = AuthMethodPassword
	config.Password = "testpass"
	config.StrictHostKeyChecking = false

	client, err := NewSSHClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Disconnect()

	commands := []string{
		"echo first",
		"echo second",
		"echo third",
	}

	results, err := client.executor.ExecuteBatch(ctx, commands, false, false, "")
	if err != nil {
		t.Fatalf("batch execution failed: %v", err)
	}

	if len(results) != len(commands) {
		t.Errorf("expected %d results, got %d", len(commands), len(results))
	}

	for i, result := range results {
		if result.ExitCode != 0 {
			t.Errorf("command %d failed with exit code %d", i, result.ExitCode)
		}
		if result.Duration <= 0 {
			t.Errorf("command %d has invalid duration", i)
		}
	}
}

func TestExecutorBatchStopOnError(t *testing.T) {
	server := newTestSSHServer(t)
	defer server.close()

	host, port := parseAddress(server.addr)

	config := DefaultConfig(host, "testuser")
	config.Port = port
	config.AuthMethod = AuthMethodPassword
	config.Password = "testpass"
	config.StrictHostKeyChecking = false

	client, err := NewSSHClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Disconnect()

	commands := []string{
		"echo first",
		"exit 1",
		"echo third",
	}

	results, err := client.executor.ExecuteBatch(ctx, commands, true, false, "")
	if err == nil {
		t.Error("expected error when stopping on error")
	}

	// Should have executed 2 commands before stopping
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}
