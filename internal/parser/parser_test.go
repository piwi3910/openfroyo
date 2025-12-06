package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadStack(t *testing.T) {
	// Create a temporary stack file
	tmpDir := t.TempDir()
	stackPath := filepath.Join(tmpDir, "test.ofy")

	stackContent := `
name: Test Stack
inventory:
  - inventory/hosts.yml
defaults:
  timeout: 30
run:
  - name: Install packages
    module: generic/apt
    hosts:
      - web-01
    vars:
      package: nginx
      state: present
  - name: Configure service
    module: generic/service
    hosts:
      - "@group:web"
    vars:
      name: nginx
      state: started
`
	if err := os.WriteFile(stackPath, []byte(stackContent), 0644); err != nil {
		t.Fatalf("Failed to write test stack file: %v", err)
	}

	// Test loading the stack
	stack, err := LoadStack(stackPath)
	if err != nil {
		t.Fatalf("LoadStack failed: %v", err)
	}

	// Verify stack contents
	if stack.Name != "Test Stack" {
		t.Errorf("Expected stack name 'Test Stack', got '%s'", stack.Name)
	}

	if len(stack.Inventory) != 1 {
		t.Errorf("Expected 1 inventory file, got %d", len(stack.Inventory))
	}

	if stack.Defaults["timeout"] != 30 {
		t.Errorf("Expected default timeout 30, got %v", stack.Defaults["timeout"])
	}

	if len(stack.Run) != 2 {
		t.Errorf("Expected 2 run entries, got %d", len(stack.Run))
	}

	// Verify first run entry
	if stack.Run[0].Name != "Install packages" {
		t.Errorf("Expected first entry name 'Install packages', got '%s'", stack.Run[0].Name)
	}
	if stack.Run[0].Module != "generic/apt" {
		t.Errorf("Expected module 'generic/apt', got '%s'", stack.Run[0].Module)
	}

	// Verify second run entry hosts
	if len(stack.Run[1].Hosts) != 1 || stack.Run[1].Hosts[0] != "@group:web" {
		t.Errorf("Expected hosts ['@group:web'], got %v", stack.Run[1].Hosts)
	}
}

func TestLoadStackWithTasks(t *testing.T) {
	tmpDir := t.TempDir()
	stackPath := filepath.Join(tmpDir, "test-tasks.ofy")

	stackContent := `
name: Task Block Test
inventory:
  - inventory/hosts.yml
run:
  - name: Deploy application
    hosts:
      - web-01
    tasks:
      - name: Upload files
        module: generic/copy
        vars:
          src: /local/app
          dest: /opt/app
      - name: Restart service
        module: generic/service
        vars:
          name: app
          state: restarted
`
	if err := os.WriteFile(stackPath, []byte(stackContent), 0644); err != nil {
		t.Fatalf("Failed to write test stack file: %v", err)
	}

	stack, err := LoadStack(stackPath)
	if err != nil {
		t.Fatalf("LoadStack failed: %v", err)
	}

	if len(stack.Run) != 1 {
		t.Fatalf("Expected 1 run entry, got %d", len(stack.Run))
	}

	entry := stack.Run[0]
	if len(entry.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(entry.Tasks))
	}

	if entry.Tasks[0].Name != "Upload files" {
		t.Errorf("Expected task name 'Upload files', got '%s'", entry.Tasks[0].Name)
	}
	if entry.Tasks[0].Module != "generic/copy" {
		t.Errorf("Expected module 'generic/copy', got '%s'", entry.Tasks[0].Module)
	}
}

func TestLoadInventory(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts.yml")

	hostsContent := `
hosts:
  web-01:
    ssh_host: 192.168.1.10
    ssh_port: 22
    ssh_user: admin
    ssh_key_file: ~/.ssh/id_rsa
  web-02:
    ssh_host: 192.168.1.11
    ssh_port: 2222
    ssh_user: root
    ssh_password: secret
  agent-01:
    mode: agent
    agent_id: agent-abc123
    datacenter: us-east-1
`
	if err := os.WriteFile(hostsPath, []byte(hostsContent), 0644); err != nil {
		t.Fatalf("Failed to write test hosts file: %v", err)
	}

	groupsPath := filepath.Join(tmpDir, "groups.yml")
	groupsContent := `
groups:
  web:
    hosts:
      - web-01
      - web-02
  all:
    hosts:
      - web-01
      - web-02
      - agent-01
`
	if err := os.WriteFile(groupsPath, []byte(groupsContent), 0644); err != nil {
		t.Fatalf("Failed to write test groups file: %v", err)
	}

	inventory, err := LoadInventory([]string{hostsPath, groupsPath})
	if err != nil {
		t.Fatalf("LoadInventory failed: %v", err)
	}

	// Verify hosts
	if len(inventory.Hosts) != 3 {
		t.Errorf("Expected 3 hosts, got %d", len(inventory.Hosts))
	}

	// Verify web-01
	host, err := inventory.GetHost("web-01")
	if err != nil {
		t.Errorf("Failed to get host web-01: %v", err)
	}
	if host.SSHHost != "192.168.1.10" {
		t.Errorf("Expected SSH host 192.168.1.10, got %s", host.SSHHost)
	}
	if host.SSHPort != 22 {
		t.Errorf("Expected SSH port 22, got %d", host.SSHPort)
	}

	// Verify web-02
	host, err = inventory.GetHost("web-02")
	if err != nil {
		t.Errorf("Failed to get host web-02: %v", err)
	}
	if host.SSHPort != 2222 {
		t.Errorf("Expected SSH port 2222, got %d", host.SSHPort)
	}

	// Verify agent host
	host, err = inventory.GetHost("agent-01")
	if err != nil {
		t.Errorf("Failed to get host agent-01: %v", err)
	}
	if host.Mode != "agent" {
		t.Errorf("Expected mode 'agent', got '%s'", host.Mode)
	}
	if host.AgentID != "agent-abc123" {
		t.Errorf("Expected agent_id 'agent-abc123', got '%s'", host.AgentID)
	}

	// Verify groups
	if len(inventory.Groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(inventory.Groups))
	}
}

func TestResolveHosts(t *testing.T) {
	inventory := &Inventory{
		Hosts: map[string]Host{
			"web-01": {SSHHost: "192.168.1.10", SSHPort: 22},
			"web-02": {SSHHost: "192.168.1.11", SSHPort: 22},
			"db-01":  {SSHHost: "192.168.1.20", SSHPort: 22},
		},
		Groups: map[string]Group{
			"web": {Hosts: []string{"web-01", "web-02"}},
			"db":  {Hosts: []string{"db-01"}},
			"all": {Hosts: []string{"web-01", "web-02", "db-01"}},
		},
	}

	// Test direct host reference
	hosts, err := inventory.ResolveHosts([]string{"web-01"})
	if err != nil {
		t.Errorf("ResolveHosts failed: %v", err)
	}
	if len(hosts) != 1 || hosts[0] != "web-01" {
		t.Errorf("Expected [web-01], got %v", hosts)
	}

	// Test group reference
	hosts, err = inventory.ResolveHosts([]string{"@group:web"})
	if err != nil {
		t.Errorf("ResolveHosts failed: %v", err)
	}
	if len(hosts) != 2 {
		t.Errorf("Expected 2 hosts, got %d", len(hosts))
	}

	// Test mixed references
	hosts, err = inventory.ResolveHosts([]string{"@group:web", "db-01"})
	if err != nil {
		t.Errorf("ResolveHosts failed: %v", err)
	}
	if len(hosts) != 3 {
		t.Errorf("Expected 3 hosts, got %d", len(hosts))
	}

	// Test invalid group
	_, err = inventory.ResolveHosts([]string{"@group:nonexistent"})
	if err == nil {
		t.Error("Expected error for nonexistent group")
	}

	// Test invalid host
	_, err = inventory.ResolveHosts([]string{"nonexistent-host"})
	if err == nil {
		t.Error("Expected error for nonexistent host")
	}
}

func TestGetHost(t *testing.T) {
	inventory := &Inventory{
		Hosts: map[string]Host{
			"web-01": {SSHHost: "192.168.1.10", SSHPort: 22, SSHUser: "admin"},
		},
	}

	// Test existing host
	host, err := inventory.GetHost("web-01")
	if err != nil {
		t.Errorf("GetHost failed: %v", err)
	}
	if host.SSHHost != "192.168.1.10" {
		t.Errorf("Expected SSH host 192.168.1.10, got %s", host.SSHHost)
	}

	// Test non-existing host
	_, err = inventory.GetHost("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent host")
	}
}
