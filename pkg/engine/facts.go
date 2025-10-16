package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openfroyo/openfroyo/pkg/stores"
	"github.com/openfroyo/openfroyo/pkg/transports/ssh"
	"github.com/rs/zerolog/log"
)

// FactsCollector collects facts from hosts.
type FactsCollector struct {
	store        stores.Store
	hostRegistry *HostRegistry
	defaultTTL   int
}

// FactsCollectionResult contains the result of a facts collection operation.
type FactsCollectionResult struct {
	TargetID     string         `json:"target_id"`
	Host         string         `json:"host"`
	FactsCount   int            `json:"facts_count"`
	CollectedAt  time.Time      `json:"collected_at"`
	Duration     time.Duration  `json:"duration"`
	Facts        map[string]any `json:"facts"`
}

// OSFacts contains OS information.
type OSFacts struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Kernel   string `json:"kernel"`
	Arch     string `json:"arch"`
	Hostname string `json:"hostname"`
}

// CPUFacts contains CPU information.
type CPUFacts struct {
	Model    string `json:"model"`
	Cores    int    `json:"cores"`
	Threads  int    `json:"threads"`
	Arch     string `json:"arch"`
	Vendor   string `json:"vendor"`
}

// MemoryFacts contains memory information.
type MemoryFacts struct {
	TotalMB    int64 `json:"total_mb"`
	AvailableMB int64 `json:"available_mb"`
	SwapTotalMB int64 `json:"swap_total_mb"`
	SwapFreeMB  int64 `json:"swap_free_mb"`
}

// DiskFacts contains disk information.
type DiskFacts struct {
	Devices []DiskDevice `json:"devices"`
}

// DiskDevice represents a disk device.
type DiskDevice struct {
	Device     string `json:"device"`
	MountPoint string `json:"mount_point"`
	FSType     string `json:"fs_type"`
	TotalGB    int64  `json:"total_gb"`
	UsedGB     int64  `json:"used_gb"`
	AvailableGB int64  `json:"available_gb"`
	UsePercent int    `json:"use_percent"`
}

// NetworkFacts contains network information.
type NetworkFacts struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

// NetworkInterface represents a network interface.
type NetworkInterface struct {
	Name       string   `json:"name"`
	IPAddresses []string `json:"ip_addresses"`
	MACAddress string   `json:"mac_address"`
	State      string   `json:"state"`
}

// PackageFacts contains package information.
type PackageFacts struct {
	Manager   string   `json:"manager"`
	Packages  []string `json:"packages"`
	Count     int      `json:"count"`
}

// NewFactsCollector creates a new facts collector.
func NewFactsCollector(store stores.Store, hostRegistry *HostRegistry) *FactsCollector {
	return &FactsCollector{
		store:        store,
		hostRegistry: hostRegistry,
		defaultTTL:   3600, // 1 hour default TTL
	}
}

// CollectFacts collects facts from a host.
func (c *FactsCollector) CollectFacts(ctx context.Context, hostID string, factTypes []string, refresh bool) (*FactsCollectionResult, error) {
	startTime := time.Now()

	// Get host information
	host, err := c.hostRegistry.GetHost(ctx, hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	log.Info().
		Str("host_id", hostID).
		Str("host", host.Address).
		Strs("fact_types", factTypes).
		Bool("refresh", refresh).
		Msg("Collecting facts")

	// Check cache if not refreshing
	if !refresh {
		// Check if facts exist and are still valid
		// For simplicity, we'll always collect for now
	}

	// Connect to host
	sshConfig := &ssh.Config{
		Host:                  host.Address,
		Port:                  host.Port,
		User:                  host.User,
		AuthMethod:            ssh.AuthMethodKey,
		PrivateKeyPath:        host.KeyPath,
		StrictHostKeyChecking: false,
		ConnectionTimeout:     30 * time.Second,
		CommandTimeout:        2 * time.Minute,
	}

	transport, err := ssh.NewSSHClient(sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client: %w", err)
	}

	if err := transport.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to host: %w", err)
	}
	defer transport.Disconnect()

	// Collect all facts
	allFacts := make(map[string]any)
	factsCount := 0

	// Determine which fact types to collect
	if len(factTypes) == 0 {
		factTypes = []string{"os.basic", "hw.cpu", "hw.memory", "hw.disk", "net.ifaces", "pkg.manifest"}
	}

	for _, factType := range factTypes {
		var factData any
		var err error

		switch factType {
		case "os.basic":
			factData, err = c.collectOSFacts(ctx, transport)
		case "hw.cpu":
			factData, err = c.collectCPUFacts(ctx, transport)
		case "hw.memory":
			factData, err = c.collectMemoryFacts(ctx, transport)
		case "hw.disk":
			factData, err = c.collectDiskFacts(ctx, transport)
		case "net.ifaces":
			factData, err = c.collectNetworkFacts(ctx, transport)
		case "pkg.manifest":
			factData, err = c.collectPackageFacts(ctx, transport)
		default:
			log.Warn().Str("type", factType).Msg("Unknown fact type")
			continue
		}

		if err != nil {
			log.Error().Err(err).Str("type", factType).Msg("Failed to collect fact")
			continue
		}

		allFacts[factType] = factData
		factsCount++

		// Store fact
		if err := c.storeFact(ctx, hostID, factType, factData); err != nil {
			log.Error().Err(err).Str("type", factType).Msg("Failed to store fact")
		}
	}

	duration := time.Since(startTime)

	log.Info().
		Str("host_id", hostID).
		Int("facts_count", factsCount).
		Dur("duration", duration).
		Msg("Facts collection completed")

	return &FactsCollectionResult{
		TargetID:    hostID,
		Host:        host.Address,
		FactsCount:  factsCount,
		CollectedAt: time.Now(),
		Duration:    duration,
		Facts:       allFacts,
	}, nil
}

// collectOSFacts collects OS information.
func (c *FactsCollector) collectOSFacts(ctx context.Context, transport *ssh.SSHClient) (*OSFacts, error) {
	facts := &OSFacts{}

	// Get OS name and version
	stdout, _, err := transport.ExecuteCommand(ctx, "cat /etc/os-release 2>/dev/null || cat /etc/lsb-release 2>/dev/null")
	if err == nil {
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "NAME=") {
				facts.Name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
			} else if strings.HasPrefix(line, "VERSION=") {
				facts.Version = strings.Trim(strings.TrimPrefix(line, "VERSION="), "\"")
			}
		}
	}

	// Get kernel version
	stdout, _, err = transport.ExecuteCommand(ctx, "uname -r")
	if err == nil {
		facts.Kernel = strings.TrimSpace(stdout)
	}

	// Get architecture
	stdout, _, err = transport.ExecuteCommand(ctx, "uname -m")
	if err == nil {
		facts.Arch = strings.TrimSpace(stdout)
	}

	// Get hostname
	stdout, _, err = transport.ExecuteCommand(ctx, "hostname")
	if err == nil {
		facts.Hostname = strings.TrimSpace(stdout)
	}

	return facts, nil
}

// collectCPUFacts collects CPU information.
func (c *FactsCollector) collectCPUFacts(ctx context.Context, transport *ssh.SSHClient) (*CPUFacts, error) {
	facts := &CPUFacts{}

	// Get CPU info from /proc/cpuinfo
	stdout, _, err := transport.ExecuteCommand(ctx, "cat /proc/cpuinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}

	lines := strings.Split(stdout, "\n")
	coreCount := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				facts.Model = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "processor") {
			coreCount++
		} else if strings.HasPrefix(line, "vendor_id") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				facts.Vendor = strings.TrimSpace(parts[1])
			}
		}
	}

	facts.Cores = coreCount
	facts.Threads = coreCount // Simplified - actual thread count may differ

	// Get architecture
	stdout, _, err = transport.ExecuteCommand(ctx, "uname -m")
	if err == nil {
		facts.Arch = strings.TrimSpace(stdout)
	}

	return facts, nil
}

// collectMemoryFacts collects memory information.
func (c *FactsCollector) collectMemoryFacts(ctx context.Context, transport *ssh.SSHClient) (*MemoryFacts, error) {
	facts := &MemoryFacts{}

	// Get memory info from /proc/meminfo
	stdout, _, err := transport.ExecuteCommand(ctx, "cat /proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/meminfo: %w", err)
	}

	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, _ := strconv.ParseInt(fields[1], 10, 64)

		switch fields[0] {
		case "MemTotal:":
			facts.TotalMB = value / 1024
		case "MemAvailable:":
			facts.AvailableMB = value / 1024
		case "SwapTotal:":
			facts.SwapTotalMB = value / 1024
		case "SwapFree:":
			facts.SwapFreeMB = value / 1024
		}
	}

	return facts, nil
}

// collectDiskFacts collects disk information.
func (c *FactsCollector) collectDiskFacts(ctx context.Context, transport *ssh.SSHClient) (*DiskFacts, error) {
	facts := &DiskFacts{
		Devices: make([]DiskDevice, 0),
	}

	// Get disk usage with df
	stdout, _, err := transport.ExecuteCommand(ctx, "df -BG -T | grep '^/'")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}

	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		device := DiskDevice{
			Device:     fields[0],
			FSType:     fields[1],
			MountPoint: fields[6],
		}

		// Parse sizes (remove 'G' suffix)
		if total, err := strconv.ParseInt(strings.TrimSuffix(fields[2], "G"), 10, 64); err == nil {
			device.TotalGB = total
		}
		if used, err := strconv.ParseInt(strings.TrimSuffix(fields[3], "G"), 10, 64); err == nil {
			device.UsedGB = used
		}
		if available, err := strconv.ParseInt(strings.TrimSuffix(fields[4], "G"), 10, 64); err == nil {
			device.AvailableGB = available
		}

		// Parse percentage
		percentStr := strings.TrimSuffix(fields[5], "%")
		if percent, err := strconv.Atoi(percentStr); err == nil {
			device.UsePercent = percent
		}

		facts.Devices = append(facts.Devices, device)
	}

	return facts, nil
}

// collectNetworkFacts collects network interface information.
func (c *FactsCollector) collectNetworkFacts(ctx context.Context, transport *ssh.SSHClient) (*NetworkFacts, error) {
	facts := &NetworkFacts{
		Interfaces: make([]NetworkInterface, 0),
	}

	// Get interface list with ip command
	stdout, _, err := transport.ExecuteCommand(ctx, "ip -o addr show")
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Parse ip addr output
	lines := strings.Split(stdout, "\n")
	interfaceMap := make(map[string]*NetworkInterface)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		ifName := fields[1]
		if strings.HasSuffix(ifName, ":") {
			ifName = strings.TrimSuffix(ifName, ":")
		}

		// Skip loopback
		if ifName == "lo" {
			continue
		}

		// Get or create interface entry
		iface, ok := interfaceMap[ifName]
		if !ok {
			iface = &NetworkInterface{
				Name:        ifName,
				IPAddresses: make([]string, 0),
				State:       "UP",
			}
			interfaceMap[ifName] = iface
		}

		// Extract IP address
		for i, field := range fields {
			if field == "inet" || field == "inet6" {
				if i+1 < len(fields) {
					ipAddr := strings.Split(fields[i+1], "/")[0]
					iface.IPAddresses = append(iface.IPAddresses, ipAddr)
				}
			}
		}
	}

	// Get MAC addresses
	for ifName, iface := range interfaceMap {
		stdout, _, err := transport.ExecuteCommand(ctx, fmt.Sprintf("cat /sys/class/net/%s/address 2>/dev/null", ifName))
		if err == nil {
			iface.MACAddress = strings.TrimSpace(stdout)
		}

		facts.Interfaces = append(facts.Interfaces, *iface)
	}

	return facts, nil
}

// collectPackageFacts collects installed package information.
func (c *FactsCollector) collectPackageFacts(ctx context.Context, transport *ssh.SSHClient) (*PackageFacts, error) {
	facts := &PackageFacts{
		Packages: make([]string, 0),
	}

	// Detect package manager and list packages
	managers := []struct {
		name    string
		command string
		pattern *regexp.Regexp
	}{
		{
			name:    "dpkg",
			command: "dpkg -l | grep '^ii' | awk '{print $2}'",
			pattern: nil,
		},
		{
			name:    "rpm",
			command: "rpm -qa",
			pattern: nil,
		},
	}

	for _, mgr := range managers {
		stdout, _, err := transport.ExecuteCommand(ctx, mgr.command)
		if err == nil && stdout != "" {
			facts.Manager = mgr.name
			lines := strings.Split(stdout, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					facts.Packages = append(facts.Packages, line)
				}
			}
			facts.Count = len(facts.Packages)
			break
		}
	}

	return facts, nil
}

// storeFact stores a fact in the database.
func (c *FactsCollector) storeFact(ctx context.Context, targetID string, namespace string, data any) error {
	// Marshal fact data to JSON
	valueBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal fact data: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(c.defaultTTL) * time.Second)

	fact := &stores.Fact{
		ID:        uuid.New().String(),
		TargetID:  targetID,
		Namespace: namespace,
		Key:       "data",
		Value:     string(valueBytes),
		TTL:       c.defaultTTL,
		ExpiresAt: &expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := c.store.UpsertFact(ctx, fact); err != nil {
		return fmt.Errorf("failed to store fact: %w", err)
	}

	return nil
}

// GetFacts retrieves cached facts for a host.
func (c *FactsCollector) GetFacts(ctx context.Context, hostID string, namespace *string) (map[string]any, error) {
	facts, err := c.store.ListFacts(ctx, &hostID, namespace, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list facts: %w", err)
	}

	result := make(map[string]any)
	for _, fact := range facts {
		if fact.Namespace == "host.metadata" || fact.Namespace == "host.labels" {
			continue // Skip host metadata
		}

		var data any
		if err := json.Unmarshal([]byte(fact.Value), &data); err != nil {
			continue // Skip invalid data
		}

		result[fact.Namespace] = data
	}

	return result, nil
}
