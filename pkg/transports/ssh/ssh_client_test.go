package ssh

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// testSSHServer provides a minimal SSH server for testing.
type testSSHServer struct {
	listener net.Listener
	config   *ssh.ServerConfig
	addr     string
	done     chan struct{}
}

// newTestSSHServer creates a new test SSH server.
func newTestSSHServer(t *testing.T) *testSSHServer {
	// Generate a test host key
	_, privateKey, err := generateTestKey()
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "testuser" && string(pass) == "testpass" {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid credentials")
		},
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			// Accept any public key for testing
			return nil, nil
		},
	}

	config.AddHostKey(privateKey)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	server := &testSSHServer{
		listener: listener,
		config:   config,
		addr:     listener.Addr().String(),
		done:     make(chan struct{}),
	}

	go server.serve()

	return server
}

// serve handles incoming connections.
func (s *testSSHServer) serve() {
	for {
		select {
		case <-s.done:
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a single SSH connection.
func (s *testSSHServer) handleConnection(netConn net.Conn) {
	defer netConn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, s.config)
	if err != nil {
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		go s.handleChannel(channel, requests)
	}
}

// handleChannel handles a single SSH channel.
func (s *testSSHServer) handleChannel(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "exec":
			command := string(req.Payload[4:]) // Skip the length prefix

			if req.WantReply {
				req.Reply(true, nil)
			}

			// Handle specific test commands
			switch command {
			case "true":
				channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
			case "echo test":
				channel.Write([]byte("test\n"))
				channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
			case "echo error >&2":
				channel.Stderr().Write([]byte("error\n"))
				channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
			case "exit 1":
				channel.SendRequest("exit-status", false, []byte{0, 0, 0, 1})
			default:
				channel.Write([]byte("command: " + command + "\n"))
				channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
			}

			return

		case "subsystem":
			if string(req.Payload[4:]) == "sftp" {
				if req.WantReply {
					req.Reply(true, nil)
				}
				// SFTP subsystem would be handled here
				return
			}
			if req.WantReply {
				req.Reply(false, nil)
			}

		case "pty-req":
			if req.WantReply {
				req.Reply(true, nil)
			}

		case "shell":
			if req.WantReply {
				req.Reply(true, nil)
			}
			// Simple shell echo
			go io.Copy(channel, channel)
			return

		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// close shuts down the test server.
func (s *testSSHServer) close() {
	close(s.done)
	s.listener.Close()
}

// generateTestKey generates a test SSH key pair.
func generateTestKey() (ssh.PublicKey, ssh.Signer, error) {
	// Generate ED25519 key pair dynamically
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	signer, err := ssh.NewSignerFromKey(privKey)
	if err != nil {
		return nil, nil, err
	}

	publicKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, nil, err
	}

	return publicKey, signer, nil
}

func TestSSHClientConnect(t *testing.T) {
	server := newTestSSHServer(t)
	defer server.close()

	host, port := parseAddress(server.addr)

	config := DefaultConfig(host, "testuser")
	config.Port = port
	config.AuthMethod = AuthMethodPassword
	config.Password = "testpass"
	config.StrictHostKeyChecking = false
	config.ConnectionTimeout = 5 * time.Second

	client, err := NewSSHClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Disconnect()

	if !client.IsConnected() {
		t.Error("expected client to be connected")
	}

	info := client.GetConnectionInfo()
	if info.Host != host {
		t.Errorf("expected host '%s', got '%s'", host, info.Host)
	}
	if info.User != "testuser" {
		t.Errorf("expected user 'testuser', got '%s'", info.User)
	}
}

func TestSSHClientHealthCheck(t *testing.T) {
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

	if err := client.HealthCheck(ctx); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestSSHClientDisconnect(t *testing.T) {
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

	if err := client.Disconnect(); err != nil {
		t.Errorf("disconnect failed: %v", err)
	}

	if client.IsConnected() {
		t.Error("expected client to be disconnected")
	}
}

func TestSSHClientExecuteCommand(t *testing.T) {
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

	t.Run("successful command", func(t *testing.T) {
		stdout, stderr, err := client.ExecuteCommand(ctx, "echo test")
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		if stdout != "test" {
			t.Errorf("expected stdout 'test', got '%s'", stdout)
		}

		if stderr != "" {
			t.Errorf("expected empty stderr, got '%s'", stderr)
		}
	})

	t.Run("command with stderr", func(t *testing.T) {
		stdout, stderr, err := client.ExecuteCommand(ctx, "echo error >&2")
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		if stdout != "" {
			t.Errorf("expected empty stdout, got '%s'", stdout)
		}

		if stderr != "error" {
			t.Errorf("expected stderr 'error', got '%s'", stderr)
		}
	})
}

func TestSSHClientKeyBasedAuth(t *testing.T) {
	server := newTestSSHServer(t)
	defer server.close()

	host, port := parseAddress(server.addr)

	// Create a temporary key file
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key")

	// Generate a valid ED25519 key
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	// Marshal the key to PEM format
	pemBlock, err := ssh.MarshalPrivateKey(privKey, "")
	if err != nil {
		t.Fatalf("failed to marshal key: %v", err)
	}
	keyBytes := pem.EncodeToMemory(pemBlock)

	if err := os.WriteFile(keyPath, keyBytes, 0600); err != nil {
		t.Fatalf("failed to write key: %v", err)
	}

	config := DefaultConfig(host, "testuser")
	config.Port = port
	config.AuthMethod = AuthMethodKey
	config.PrivateKeyPath = keyPath
	config.StrictHostKeyChecking = false

	client, err := NewSSHClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect with key auth: %v", err)
	}
	defer client.Disconnect()

	if !client.IsConnected() {
		t.Error("expected client to be connected")
	}
}

// parseAddress splits an address into host and port.
func parseAddress(addr string) (string, int) {
	host, portStr, _ := net.SplitHostPort(addr)
	port := 0
	fmt.Sscanf(portStr, "%d", &port)
	return host, port
}
