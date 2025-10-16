package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("example.com", "testuser")

	if config.Host != "example.com" {
		t.Errorf("expected host 'example.com', got '%s'", config.Host)
	}

	if config.User != "testuser" {
		t.Errorf("expected user 'testuser', got '%s'", config.User)
	}

	if config.Port != 22 {
		t.Errorf("expected port 22, got %d", config.Port)
	}

	if config.AuthMethod != AuthMethodKey {
		t.Errorf("expected auth method 'key', got '%s'", config.AuthMethod)
	}

	if config.ConnectionTimeout != 30*time.Second {
		t.Errorf("expected connection timeout 30s, got %v", config.ConnectionTimeout)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		modifyFunc  func(*Config)
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			modifyFunc: func(c *Config) {
				c.AuthMethod = AuthMethodPassword
				c.Password = "secret"
			},
			expectError: false,
		},
		{
			name: "missing host",
			modifyFunc: func(c *Config) {
				c.Host = ""
			},
			expectError: true,
			errorMsg:    "host is required",
		},
		{
			name: "invalid port",
			modifyFunc: func(c *Config) {
				c.Port = 0
			},
			expectError: true,
			errorMsg:    "invalid port",
		},
		{
			name: "missing user",
			modifyFunc: func(c *Config) {
				c.User = ""
			},
			expectError: true,
			errorMsg:    "user is required",
		},
		{
			name: "password auth without password",
			modifyFunc: func(c *Config) {
				c.AuthMethod = AuthMethodPassword
				c.Password = ""
			},
			expectError: true,
			errorMsg:    "password is required",
		},
		{
			name: "key auth without key path",
			modifyFunc: func(c *Config) {
				c.AuthMethod = AuthMethodKey
				c.PrivateKeyPath = "/nonexistent/key"
			},
			expectError: true,
			errorMsg:    "private key file not found",
		},
		{
			name: "invalid connection timeout",
			modifyFunc: func(c *Config) {
				c.ConnectionTimeout = 0
			},
			expectError: true,
			errorMsg:    "connection timeout must be positive",
		},
		{
			name: "invalid command timeout",
			modifyFunc: func(c *Config) {
				c.CommandTimeout = 0
			},
			expectError: true,
			errorMsg:    "command timeout must be positive",
		},
		{
			name: "pooling enabled with invalid max size",
			modifyFunc: func(c *Config) {
				c.EnableConnectionPooling = true
				c.MaxPoolSize = 0
			},
			expectError: true,
			errorMsg:    "max pool size must be positive",
		},
		{
			name: "proxy with missing user",
			modifyFunc: func(c *Config) {
				c.ProxyHost = "proxy.example.com"
				c.ProxyUser = ""
			},
			expectError: true,
			errorMsg:    "proxy user is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig("example.com", "testuser")
			tt.modifyFunc(config)

			err := config.Validate()

			if tt.expectError && err == nil {
				t.Errorf("expected error containing '%s', got nil", tt.errorMsg)
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				// Check if error message contains expected text
				if len(err.Error()) == 0 {
					t.Errorf("expected error message containing '%s', got empty error", tt.errorMsg)
				}
			}
		})
	}
}

func TestConfigAddress(t *testing.T) {
	config := DefaultConfig("example.com", "testuser")
	config.Port = 2222

	expected := "example.com:2222"
	if address := config.Address(); address != expected {
		t.Errorf("expected address '%s', got '%s'", expected, address)
	}
}

func TestConfigProxyAddress(t *testing.T) {
	config := DefaultConfig("example.com", "testuser")
	config.ProxyHost = "proxy.example.com"
	config.ProxyPort = 2222

	expected := "proxy.example.com:2222"
	if address := config.ProxyAddress(); address != expected {
		t.Errorf("expected proxy address '%s', got '%s'", expected, address)
	}

	// Test with no proxy
	config.ProxyHost = ""
	if address := config.ProxyAddress(); address != "" {
		t.Errorf("expected empty proxy address, got '%s'", address)
	}
}

func TestConfigIsProxyEnabled(t *testing.T) {
	config := DefaultConfig("example.com", "testuser")

	if config.IsProxyEnabled() {
		t.Error("expected proxy to be disabled")
	}

	config.ProxyHost = "proxy.example.com"
	if !config.IsProxyEnabled() {
		t.Error("expected proxy to be enabled")
	}
}

func TestBuildSSHClientConfig(t *testing.T) {
	t.Run("password authentication", func(t *testing.T) {
		config := DefaultConfig("example.com", "testuser")
		config.AuthMethod = AuthMethodPassword
		config.Password = "secret"
		config.StrictHostKeyChecking = false

		clientConfig, err := config.BuildSSHClientConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if clientConfig.User != "testuser" {
			t.Errorf("expected user 'testuser', got '%s'", clientConfig.User)
		}

		if len(clientConfig.Auth) != 1 {
			t.Errorf("expected 1 auth method, got %d", len(clientConfig.Auth))
		}

		if clientConfig.Timeout != 30*time.Second {
			t.Errorf("expected timeout 30s, got %v", clientConfig.Timeout)
		}
	})

	t.Run("key authentication with valid key", func(t *testing.T) {
		// Create a temporary SSH key for testing
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "test_key")

		// Generate a valid ED25519 key using Go's crypto
		_, privKey, err := generateTestKeyForConfig()
		if err != nil {
			t.Fatalf("failed to generate test key: %v", err)
		}

		// Marshal the private key to OpenSSH format
		keyBytes, err := marshalED25519PrivateKey(privKey)
		if err != nil {
			t.Fatalf("failed to marshal key: %v", err)
		}

		if err := os.WriteFile(keyPath, keyBytes, 0600); err != nil {
			t.Fatalf("failed to write test key: %v", err)
		}

		config := DefaultConfig("example.com", "testuser")
		config.AuthMethod = AuthMethodKey
		config.PrivateKeyPath = keyPath
		config.StrictHostKeyChecking = false

		clientConfig, err := config.BuildSSHClientConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if clientConfig.User != "testuser" {
			t.Errorf("expected user 'testuser', got '%s'", clientConfig.User)
		}

		if len(clientConfig.Auth) != 1 {
			t.Errorf("expected 1 auth method, got %d", len(clientConfig.Auth))
		}
	})

	t.Run("agent authentication not implemented", func(t *testing.T) {
		config := DefaultConfig("example.com", "testuser")
		config.AuthMethod = AuthMethodAgent

		_, err := config.BuildSSHClientConfig()
		if err == nil {
			t.Error("expected error for agent auth, got nil")
		}
	})
}

// generateTestKeyForConfig generates an ED25519 key pair for testing.
func generateTestKeyForConfig() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return pubKey, privKey, nil
}

// marshalED25519PrivateKey marshals an ED25519 private key to PEM format.
func marshalED25519PrivateKey(privKey ed25519.PrivateKey) ([]byte, error) {
	pemBlock, err := ssh.MarshalPrivateKey(privKey, "")
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(pemBlock), nil
}
