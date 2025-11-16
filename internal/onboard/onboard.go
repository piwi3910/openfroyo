package onboard

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"

	sshclient "github.com/piwi3910/openfroyo/internal/ssh"
)

const (
	openfryoUser     = "openfroyo"
	sshKeyBits       = 4096
	defaultSSHKeyDir = ".ssh"
)

// OnboardHost sets up the openfroyo user on a remote host
func OnboardHost(host string, port int, user, password, inventoryFile string) error {
	fmt.Printf("🚀 Starting onboarding process for %s\n\n", host)

	// Step 1: Connect to remote host
	fmt.Printf("📡 Connecting to %s:%d as %s...\n", host, port, user)
	client, err := connectSSH(host, port, user, password)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()
	fmt.Println("✓ Connected successfully")

	// Step 2: Check if user is root or can sudo
	fmt.Println("\n🔑 Checking sudo privileges...")
	isRoot, canSudo, err := checkSudoAccess(client, password)
	if err != nil {
		return fmt.Errorf("failed to check sudo access: %w", err)
	}

	if !isRoot && !canSudo {
		return fmt.Errorf("user %s does not have root access or sudo privileges", user)
	}

	if isRoot {
		fmt.Println("✓ Running as root")
	} else {
		fmt.Println("✓ User has sudo privileges")
	}

	// Step 3: Create openfroyo user
	fmt.Printf("\n👤 Creating user '%s'...\n", openfryoUser)
	if err := createUser(client, isRoot, password); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	fmt.Printf("✓ User '%s' created successfully\n", openfryoUser)

	// Step 4: Configure passwordless sudo
	fmt.Printf("\n🔓 Configuring passwordless sudo for '%s'...\n", openfryoUser)
	if err := configureSudo(client, isRoot, password); err != nil {
		return fmt.Errorf("failed to configure sudo: %w", err)
	}
	fmt.Println("✓ Passwordless sudo configured")

	// Step 5: Generate SSH key pair
	fmt.Println("\n🔐 Generating SSH key pair...")
	privateKeyPath, publicKey, err := generateSSHKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate SSH keys: %w", err)
	}
	fmt.Printf("✓ SSH keys generated at %s\n", privateKeyPath)

	// Step 6: Copy public key to remote host
	fmt.Println("\n📤 Copying SSH public key to remote host...")
	if err := installSSHKey(client, publicKey, isRoot, password); err != nil {
		return fmt.Errorf("failed to install SSH key: %w", err)
	}
	fmt.Println("✓ SSH public key installed")

	// Step 7: Test SSH key authentication
	fmt.Println("\n🧪 Testing SSH key authentication...")
	if err := testSSHKeyAuth(host, port, privateKeyPath); err != nil {
		return fmt.Errorf("SSH key authentication test failed: %w", err)
	}
	fmt.Println("✓ SSH key authentication working")

	// Step 8: Update inventory file if provided
	if inventoryFile != "" {
		fmt.Printf("\n📝 Updating inventory file %s...\n", inventoryFile)
		if err := updateInventory(inventoryFile, host, port, privateKeyPath); err != nil {
			return fmt.Errorf("failed to update inventory: %w", err)
		}
		fmt.Println("✓ Inventory file updated")
	}

	fmt.Printf("\n✅ Onboarding completed successfully!\n\n")
	fmt.Printf("Host: %s\n", host)
	fmt.Printf("User: %s\n", openfryoUser)
	fmt.Printf("SSH Key: %s\n", privateKeyPath)
	if inventoryFile != "" {
		fmt.Printf("Inventory: %s (updated)\n", inventoryFile)
	}

	return nil
}

func connectSSH(host string, port int, user, password string) (*ssh.Client, error) {
	return sshclient.NewSimpleClient(host, port, user, password)
}

func checkSudoAccess(client *ssh.Client, password string) (isRoot bool, canSudo bool, err error) {
	// Check if running as root
	output, err := runCommand(client, "id -u")
	if err != nil {
		return false, false, err
	}

	if strings.TrimSpace(output) == "0" {
		return true, true, nil
	}

	// Check if can sudo with password
	cmd := fmt.Sprintf("echo '%s' | sudo -S whoami 2>&1", password)
	output, err = runCommand(client, cmd)

	// Check if output contains "root" even if there are warnings
	if strings.Contains(output, "root") {
		return false, true, nil
	}

	return false, false, nil
}

func createUser(client *ssh.Client, isRoot bool, sudoPassword string) error {
	// Check if user already exists
	_, err := runCommand(client, fmt.Sprintf("id -u %s 2>/dev/null", openfryoUser))
	if err == nil {
		fmt.Printf("  ℹ User '%s' already exists\n", openfryoUser)
		return nil
	}

	var cmd string
	if isRoot {
		cmd = fmt.Sprintf("useradd -m -s /bin/bash %s", openfryoUser)
	} else {
		cmd = fmt.Sprintf("echo '%s' | sudo -S useradd -m -s /bin/bash %s", sudoPassword, openfryoUser)
	}

	_, err = runCommand(client, cmd)
	return err
}

func configureSudo(client *ssh.Client, isRoot bool, sudoPassword string) error {
	sudoersContent := fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL", openfryoUser)
	sudoersFile := fmt.Sprintf("/etc/sudoers.d/%s", openfryoUser)

	var cmd string
	if isRoot {
		cmd = fmt.Sprintf("echo '%s' > %s && chmod 0440 %s", sudoersContent, sudoersFile, sudoersFile)
	} else {
		cmd = fmt.Sprintf("echo '%s' | sudo -S sh -c \"echo '%s' > %s && chmod 0440 %s\"",
			sudoPassword, sudoersContent, sudoersFile, sudoersFile)
	}

	_, err := runCommand(client, cmd)
	return err
}

func generateSSHKeyPair() (privateKeyPath string, publicKey string, err error) {
	// Create .ssh directory in user's home
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	sshDir := filepath.Join(homeDir, defaultSSHKeyDir)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return "", "", err
	}

	privateKeyPath = filepath.Join(sshDir, "openfroyo_rsa")
	publicKeyPath := privateKeyPath + ".pub"

	// Check if key already exists
	if _, err := os.Stat(privateKeyPath); err == nil {
		fmt.Printf("  ℹ SSH key already exists at %s\n", privateKeyPath)
		// Read existing public key
		pubKeyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to read existing public key: %w", err)
		}
		return privateKeyPath, string(pubKeyBytes), nil
	}

	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, sshKeyBits)
	if err != nil {
		return "", "", err
	}

	// Encode private key to PEM
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write private key to file
	privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", err
	}
	defer privateKeyFile.Close()

	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return "", "", err
	}

	// Generate public key
	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	publicKey = string(ssh.MarshalAuthorizedKey(publicRsaKey))

	// Write public key to file
	if err := os.WriteFile(publicKeyPath, []byte(publicKey), 0644); err != nil {
		return "", "", err
	}

	return privateKeyPath, publicKey, nil
}

func installSSHKey(client *ssh.Client, publicKey string, isRoot bool, sudoPassword string) error {
	// Create .ssh directory for openfroyo user
	var cmd string
	if isRoot {
		cmd = fmt.Sprintf("mkdir -p /home/%s/.ssh && chmod 700 /home/%s/.ssh", openfryoUser, openfryoUser)
	} else {
		cmd = fmt.Sprintf("echo '%s' | sudo -S sh -c 'mkdir -p /home/%s/.ssh && chmod 700 /home/%s/.ssh'",
			sudoPassword, openfryoUser, openfryoUser)
	}

	if _, err := runCommand(client, cmd); err != nil {
		return err
	}

	// Add public key to authorized_keys
	publicKey = strings.TrimSpace(publicKey)
	authorizedKeysFile := fmt.Sprintf("/home/%s/.ssh/authorized_keys", openfryoUser)

	if isRoot {
		cmd = fmt.Sprintf("echo '%s' >> %s && chmod 600 %s && chown -R %s:%s /home/%s/.ssh",
			publicKey, authorizedKeysFile, authorizedKeysFile, openfryoUser, openfryoUser, openfryoUser)
	} else {
		cmd = fmt.Sprintf("echo '%s' | sudo -S sh -c \"echo '%s' >> %s && chmod 600 %s && chown -R %s:%s /home/%s/.ssh\"",
			sudoPassword, publicKey, authorizedKeysFile, authorizedKeysFile, openfryoUser, openfryoUser, openfryoUser)
	}

	_, err := runCommand(client, cmd)
	return err
}

func testSSHKeyAuth(host string, port int, privateKeyPath string) error {
	// Read private key
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Connect with key
	config := &ssh.ClientConfig{
		User: openfryoUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to connect with SSH key: %w", err)
	}
	defer client.Close()

	// Test sudo access
	output, err := runCommand(client, "sudo whoami")
	if err != nil {
		return fmt.Errorf("sudo test failed: %w", err)
	}

	if strings.TrimSpace(output) != "root" {
		return fmt.Errorf("unexpected sudo output: %s", output)
	}

	return nil
}

func updateInventory(inventoryFile, host string, port int, privateKeyPath string) error {
	// Structure for inventory file
	type InventoryHost struct {
		SSHHost    string `yaml:"ssh_host"`
		SSHPort    int    `yaml:"ssh_port"`
		SSHUser    string `yaml:"ssh_user"`
		SSHKeyFile string `yaml:"ssh_key_file"`
	}

	type Inventory struct {
		Hosts map[string]InventoryHost `yaml:"hosts"`
	}

	var inventory Inventory

	// Read existing inventory if it exists
	if _, err := os.Stat(inventoryFile); err == nil {
		data, err := os.ReadFile(inventoryFile)
		if err != nil {
			return fmt.Errorf("failed to read inventory file: %w", err)
		}

		if err := yaml.Unmarshal(data, &inventory); err != nil {
			return fmt.Errorf("failed to parse inventory file: %w", err)
		}
	} else {
		// Create new inventory structure
		inventory.Hosts = make(map[string]InventoryHost)
	}

	// Generate host name (use IP or hostname)
	hostName := strings.ReplaceAll(host, ".", "-")

	// Add or update host entry
	inventory.Hosts[hostName] = InventoryHost{
		SSHHost:    host,
		SSHPort:    port,
		SSHUser:    openfryoUser,
		SSHKeyFile: privateKeyPath,
	}

	// Write back to file
	data, err := yaml.Marshal(&inventory)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %w", err)
	}

	// Create directory if needed
	dir := filepath.Dir(inventoryFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(inventoryFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write inventory file: %w", err)
	}

	return nil
}

func runCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	return string(output), err
}
