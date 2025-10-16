package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openfroyo/openfroyo/pkg/stores"
)

// Host represents a managed host in the inventory.
type Host struct {
	ID          string            `json:"id"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	User        string            `json:"user"`
	KeyPath     string            `json:"key_path"`
	Labels      map[string]string `json:"labels"`
	OnboardedAt time.Time         `json:"onboarded_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// HostRegistry manages the host inventory.
type HostRegistry struct {
	store stores.Store
}

// NewHostRegistry creates a new host registry.
func NewHostRegistry(store stores.Store) *HostRegistry {
	return &HostRegistry{
		store: store,
	}
}

// AddHost adds a new host to the registry.
func (r *HostRegistry) AddHost(ctx context.Context, host *Host) error {
	if host.ID == "" {
		host.ID = uuid.New().String()
	}

	now := time.Now()
	if host.CreatedAt.IsZero() {
		host.CreatedAt = now
	}
	host.UpdatedAt = now

	// Marshal host data
	hostData, err := json.Marshal(host)
	if err != nil {
		return fmt.Errorf("failed to marshal host data: %w", err)
	}

	// Store as fact
	fact := &stores.Fact{
		ID:        uuid.New().String(),
		TargetID:  host.ID,
		Namespace: "host.metadata",
		Key:       "info",
		Value:     string(hostData),
		TTL:       0, // No expiry
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := r.store.UpsertFact(ctx, fact); err != nil {
		return fmt.Errorf("failed to store host: %w", err)
	}

	// Store labels separately for easier querying
	if len(host.Labels) > 0 {
		labelsData, err := json.Marshal(host.Labels)
		if err != nil {
			return fmt.Errorf("failed to marshal labels: %w", err)
		}

		labelsFact := &stores.Fact{
			ID:        uuid.New().String(),
			TargetID:  host.ID,
			Namespace: "host.labels",
			Key:       "all",
			Value:     string(labelsData),
			TTL:       0,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := r.store.UpsertFact(ctx, labelsFact); err != nil {
			return fmt.Errorf("failed to store labels: %w", err)
		}
	}

	return nil
}

// GetHost retrieves a host by ID.
func (r *HostRegistry) GetHost(ctx context.Context, hostID string) (*Host, error) {
	fact, err := r.store.GetFact(ctx, hostID, "host.metadata", "info")
	if err != nil {
		return nil, fmt.Errorf("host not found: %s", hostID)
	}

	var host Host
	if err := json.Unmarshal([]byte(fact.Value), &host); err != nil {
		return nil, fmt.Errorf("failed to unmarshal host data: %w", err)
	}

	return &host, nil
}

// GetHostByAddress retrieves a host by address.
func (r *HostRegistry) GetHostByAddress(ctx context.Context, address string) (*Host, error) {
	// List all hosts and filter by address
	hosts, err := r.ListHosts(ctx)
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		if host.Address == address {
			return host, nil
		}
	}

	return nil, fmt.Errorf("host not found: %s", address)
}

// ListHosts lists all registered hosts.
func (r *HostRegistry) ListHosts(ctx context.Context) ([]*Host, error) {
	namespace := "host.metadata"
	facts, err := r.store.ListFacts(ctx, nil, &namespace, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts: %w", err)
	}

	hosts := make([]*Host, 0, len(facts))
	for _, fact := range facts {
		if fact.Key != "info" {
			continue
		}

		var host Host
		if err := json.Unmarshal([]byte(fact.Value), &host); err != nil {
			continue // Skip invalid entries
		}

		hosts = append(hosts, &host)
	}

	return hosts, nil
}

// SelectHosts selects hosts based on a selector.
// Selector format: "key1=value1,key2=value2" or "all" for all hosts.
func (r *HostRegistry) SelectHosts(ctx context.Context, selector string) ([]*Host, error) {
	if selector == "" || selector == "all" {
		return r.ListHosts(ctx)
	}

	// Parse selector into label map
	labels := parseSelector(selector)

	// Get all hosts
	allHosts, err := r.ListHosts(ctx)
	if err != nil {
		return nil, err
	}

	// Filter hosts by labels
	selectedHosts := make([]*Host, 0)
	for _, host := range allHosts {
		if matchesLabels(host.Labels, labels) {
			selectedHosts = append(selectedHosts, host)
		}
	}

	return selectedHosts, nil
}

// UpdateHost updates an existing host.
func (r *HostRegistry) UpdateHost(ctx context.Context, host *Host) error {
	host.UpdatedAt = time.Now()

	// Marshal host data
	hostData, err := json.Marshal(host)
	if err != nil {
		return fmt.Errorf("failed to marshal host data: %w", err)
	}

	// Update fact
	fact := &stores.Fact{
		ID:        uuid.New().String(),
		TargetID:  host.ID,
		Namespace: "host.metadata",
		Key:       "info",
		Value:     string(hostData),
		TTL:       0,
		UpdatedAt: host.UpdatedAt,
	}

	if err := r.store.UpsertFact(ctx, fact); err != nil {
		return fmt.Errorf("failed to update host: %w", err)
	}

	// Update labels
	if len(host.Labels) > 0 {
		labelsData, err := json.Marshal(host.Labels)
		if err != nil {
			return fmt.Errorf("failed to marshal labels: %w", err)
		}

		labelsFact := &stores.Fact{
			ID:        uuid.New().String(),
			TargetID:  host.ID,
			Namespace: "host.labels",
			Key:       "all",
			Value:     string(labelsData),
			TTL:       0,
			UpdatedAt: host.UpdatedAt,
		}

		if err := r.store.UpsertFact(ctx, labelsFact); err != nil {
			return fmt.Errorf("failed to update labels: %w", err)
		}
	}

	return nil
}

// DeleteHost removes a host from the registry.
func (r *HostRegistry) DeleteHost(ctx context.Context, hostID string) error {
	// Get all facts for this host
	targetID := hostID
	facts, err := r.store.ListFacts(ctx, &targetID, nil, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to list host facts: %w", err)
	}

	// Delete all facts
	for _, fact := range facts {
		if err := r.store.DeleteFact(ctx, fact.ID); err != nil {
			return fmt.Errorf("failed to delete fact: %w", err)
		}
	}

	return nil
}

// parseSelector parses a label selector string into a map.
// Format: "key1=value1,key2=value2"
func parseSelector(selector string) map[string]string {
	labels := make(map[string]string)

	if selector == "" || selector == "all" {
		return labels
	}

	pairs := strings.Split(selector, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			labels[key] = value
		}
	}

	return labels
}

// matchesLabels checks if host labels match the selector labels.
func matchesLabels(hostLabels, selectorLabels map[string]string) bool {
	if len(selectorLabels) == 0 {
		return true
	}

	for key, value := range selectorLabels {
		hostValue, ok := hostLabels[key]
		if !ok || hostValue != value {
			return false
		}
	}

	return true
}
