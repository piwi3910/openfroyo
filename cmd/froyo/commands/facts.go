package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/openfroyo/openfroyo/pkg/engine"
	"github.com/openfroyo/openfroyo/pkg/stores"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newFactsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "facts",
		Short: "Facts collection and management",
		Long: `Collect and manage typed facts about infrastructure.

Facts are structured data about systems, including:
  - OS and kernel information
  - Hardware specifications (CPU, memory, disk)
  - Network interfaces and configuration
  - Installed packages
  - Running services
  - Custom application facts`,
	}

	cmd.AddCommand(newFactsCollectCommand())
	cmd.AddCommand(newFactsListCommand())
	cmd.AddCommand(newFactsShowCommand())

	return cmd
}

func newFactsCollectCommand() *cobra.Command {
	var (
		selector  string
		targets   []string
		factTypes []string
		refresh   bool
	)

	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect facts from targets",
		Long: `Collect facts from target hosts.

Facts are gathered using various collectors:
  - os.basic: OS name, version, kernel
  - hw.cpu: CPU model, cores, architecture
  - hw.memory: RAM size, swap
  - hw.disk: Disk devices, capacity, usage
  - net.ifaces: Network interfaces, IPs, routes
  - pkg.manifest: Installed packages
  - svc.running: Active services

Facts are cached with TTL and can be refreshed on demand.`,
		Example: `  # Collect all facts from all hosts
  froyo facts collect

  # Collect facts from specific hosts
  froyo facts collect --target web1 --target web2

  # Collect specific fact types
  froyo facts collect --type os.basic --type hw.cpu

  # Collect using selector
  froyo facts collect --selector 'env=prod,role=web'

  # Force refresh cached facts
  froyo facts collect --refresh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("selector", selector).
				Strs("targets", targets).
				Strs("types", factTypes).
				Bool("refresh", refresh).
				Msg("Collecting facts")

			ctx := context.Background()

			// Load configuration
			dataDir := "./data"
			if configPath != "" {
				dataDir = filepath.Join(filepath.Dir(configPath), "data")
			}

			// Initialize store
			dbPath := filepath.Join(dataDir, "openfroyo.db")
			store, err := stores.NewSQLiteStore(stores.Config{
				Path: dbPath,
			})
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}

			if err := store.Init(ctx); err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}
			defer store.Close()

			// Create host registry and facts collector
			hostRegistry := engine.NewHostRegistry(store)
			factsCollector := engine.NewFactsCollector(store, hostRegistry)

			// Resolve target hosts
			var hostsToProcess []*engine.Host

			if len(targets) > 0 {
				// Explicit target list
				for _, targetID := range targets {
					host, err := hostRegistry.GetHost(ctx, targetID)
					if err != nil {
						log.Warn().Str("target", targetID).Err(err).Msg("Failed to get host")
						continue
					}
					hostsToProcess = append(hostsToProcess, host)
				}
			} else {
				// Use selector (default: all hosts)
				if selector == "" {
					selector = "all"
				}
				hostsToProcess, err = hostRegistry.SelectHosts(ctx, selector)
				if err != nil {
					return fmt.Errorf("failed to select hosts: %w", err)
				}
			}

			if len(hostsToProcess) == 0 {
				fmt.Println("No hosts found to collect facts from.")
				fmt.Println("\n💡 Tip: Onboard a host first with:")
				fmt.Println("   froyo onboard ssh --host <ip> --user root --password <pass>")
				return nil
			}

			fmt.Printf("Collecting facts from %d host(s)...\n\n", len(hostsToProcess))

			// Collect facts from each host
			successCount := 0
			for _, host := range hostsToProcess {
				fmt.Printf("📊 Collecting facts from %s (%s)...\n", host.Address, host.ID)

				result, err := factsCollector.CollectFacts(ctx, host.ID, factTypes, refresh)
				if err != nil {
					log.Error().Err(err).Str("host", host.Address).Msg("Failed to collect facts")
					fmt.Printf("   ❌ Error: %v\n\n", err)
					continue
				}

				fmt.Printf("   ✓ Collected %d fact types in %v\n", result.FactsCount, result.Duration)
				for factType := range result.Facts {
					fmt.Printf("     - %s\n", factType)
				}
				fmt.Println()
				successCount++
			}

			fmt.Printf("\n✅ Facts collection completed: %d/%d hosts successful\n", successCount, len(hostsToProcess))

			return nil
		},
	}

	cmd.Flags().StringVar(&selector, "selector", "", "target selector (label query)")
	cmd.Flags().StringSliceVarP(&targets, "target", "t", nil, "specific target hosts")
	cmd.Flags().StringSliceVar(&factTypes, "type", nil, "specific fact types to collect")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "force refresh cached facts")

	return cmd
}

func newFactsListCommand() *cobra.Command {
	var (
		target    string
		factType  string
		showStale bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List collected facts",
		Long: `List facts stored in the database.

Shows available facts with their age and validity status.`,
		Example: `  # List all facts
  froyo facts list

  # List facts for specific host
  froyo facts list --target web1

  # List specific fact type
  froyo facts list --type os.basic

  # Include stale facts
  froyo facts list --show-stale`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("target", target).
				Str("type", factType).
				Bool("show_stale", showStale).
				Msg("Listing facts")

			// TODO: Implement facts listing
			// - Query facts from database
			// - Filter by target and/or type
			// - Check TTL and mark stale facts
			// - Format output (table or JSON)

			fmt.Println("Not implemented yet: facts listing")
			fmt.Printf("Would list facts: target=%s, type=%s, show_stale=%v\n",
				target, factType, showStale)

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "filter by target host")
	cmd.Flags().StringVar(&factType, "type", "", "filter by fact type")
	cmd.Flags().BoolVar(&showStale, "show-stale", false, "include stale facts")

	return cmd
}

func newFactsShowCommand() *cobra.Command {
	var (
		target   string
		factType string
	)

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show detailed facts",
		Long: `Display detailed fact data for a target.

Shows the complete fact payload in JSON format.`,
		Example: `  # Show all facts for a host
  froyo facts show --target web1

  # Show specific fact type for a host
  froyo facts show --target web1 --type os.basic`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("target", target).
				Str("type", factType).
				Msg("Showing facts")

			ctx := context.Background()

			// Load configuration
			dataDir := "./data"
			if configPath != "" {
				dataDir = filepath.Join(filepath.Dir(configPath), "data")
			}

			// Initialize store
			dbPath := filepath.Join(dataDir, "openfroyo.db")
			store, err := stores.NewSQLiteStore(stores.Config{
				Path: dbPath,
			})
			if err != nil {
				return fmt.Errorf("failed to create store: %w", err)
			}

			if err := store.Init(ctx); err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}
			defer store.Close()

			// Create host registry and facts collector
			hostRegistry := engine.NewHostRegistry(store)
			factsCollector := engine.NewFactsCollector(store, hostRegistry)

			// Get host information
			host, err := hostRegistry.GetHost(ctx, target)
			if err != nil {
				return fmt.Errorf("host not found: %s", target)
			}

			fmt.Printf("Facts for host: %s (%s)\n\n", host.Address, host.ID)

			// Get facts
			var namespace *string
			if factType != "" {
				namespace = &factType
			}

			facts, err := factsCollector.GetFacts(ctx, target, namespace)
			if err != nil {
				return fmt.Errorf("failed to get facts: %w", err)
			}

			if len(facts) == 0 {
				fmt.Println("No facts available for this host.")
				fmt.Println("\n💡 Tip: Collect facts first with:")
				fmt.Printf("   froyo facts collect --target %s\n", target)
				return nil
			}

			// Pretty-print facts as JSON
			jsonData, err := json.MarshalIndent(facts, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal facts: %w", err)
			}

			fmt.Println(string(jsonData))

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "target host")
	cmd.Flags().StringVar(&factType, "type", "", "fact type")
	cmd.MarkFlagRequired("target")

	return cmd
}
