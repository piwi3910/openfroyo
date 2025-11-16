package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadInventory loads inventory from YAML files
func LoadInventory(paths []string) (*Inventory, error) {
	inventory := &Inventory{
		Hosts:  make(map[string]Host),
		Groups: make(map[string]Group),
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read inventory file %s: %w", path, err)
		}

		var inv Inventory
		if err := yaml.Unmarshal(data, &inv); err != nil {
			return nil, fmt.Errorf("failed to parse inventory file %s: %w", path, err)
		}

		// Merge hosts
		for name, host := range inv.Hosts {
			inventory.Hosts[name] = host
		}

		// Merge groups
		for name, group := range inv.Groups {
			inventory.Groups[name] = group
		}
	}

	// Set defaults for hosts
	for name, host := range inventory.Hosts {
		if host.SSHPort == 0 {
			host.SSHPort = 22
		}
		if host.SSHUser == "" {
			host.SSHUser = "root"
		}
		inventory.Hosts[name] = host
	}

	return inventory, nil
}

// ResolveHosts resolves a list of host/group references to actual hosts
func (inv *Inventory) ResolveHosts(refs []string) ([]string, error) {
	hostSet := make(map[string]bool)

	for _, ref := range refs {
		// Check if it's a group reference (@group:name)
		if len(ref) > 7 && ref[:7] == "@group:" {
			groupName := ref[7:]
			group, exists := inv.Groups[groupName]
			if !exists {
				return nil, fmt.Errorf("group not found: %s", groupName)
			}
			for _, hostName := range group.Hosts {
				hostSet[hostName] = true
			}
		} else {
			// It's a direct host reference
			if _, exists := inv.Hosts[ref]; !exists {
				return nil, fmt.Errorf("host not found: %s", ref)
			}
			hostSet[ref] = true
		}
	}

	// Convert set to slice
	hosts := make([]string, 0, len(hostSet))
	for host := range hostSet {
		hosts = append(hosts, host)
	}

	return hosts, nil
}

// GetHost retrieves a host by name
func (inv *Inventory) GetHost(name string) (Host, error) {
	host, exists := inv.Hosts[name]
	if !exists {
		return Host{}, fmt.Errorf("host not found: %s", name)
	}
	return host, nil
}
