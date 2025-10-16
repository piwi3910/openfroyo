// Complex dependency example
// This demonstrates different dependency types: require, notify, order

package config

workspace: {
	name:    "complex-deps"
	version: "1.0"

	providers: [
		{name: "linux.pkg", version: ">=1.0.0"},
		{name: "linux.service", version: ">=1.0.0"},
		{name: "linux.file", version: ">=1.0.0"},
		{name: "linux.firewall", version: ">=1.0.0"},
	]
}

resources: {
	// 1. Install PostgreSQL
	postgres_pkg: {
		type: "linux.pkg"
		name: "postgresql"

		config: {
			package: "postgresql-14"
			state:   "present"
		}

		target: {
			labels: {
				role: "database"
			}
		}

		labels: {
			component: "database"
		}
	}

	// 2. Configure PostgreSQL (requires package)
	postgres_config: {
		type: "linux.file"
		name: "postgresql-config"

		config: {
			path: "/etc/postgresql/14/main/postgresql.conf"
			content: """
				listen_addresses = '*'
				max_connections = 100
				shared_buffers = 256MB
				"""
			owner: "postgres"
			group: "postgres"
			mode:  "0644"
		}

		target: {
			labels: {
				role: "database"
			}
		}

		dependencies: [
			{
				resource_id: "postgres_pkg"
				type:        "require"
			},
		]

		labels: {
			component: "database"
		}
	}

	// 3. PostgreSQL HBA config (requires package, notifies service)
	postgres_hba: {
		type: "linux.file"
		name: "postgresql-hba"

		config: {
			path: "/etc/postgresql/14/main/pg_hba.conf"
			content: """
				local   all             postgres                                peer
				host    all             all             127.0.0.1/32            md5
				host    all             all             ::1/128                 md5
				host    all             all             10.0.0.0/8              md5
				"""
			owner: "postgres"
			group: "postgres"
			mode:  "0640"
		}

		target: {
			labels: {
				role: "database"
			}
		}

		dependencies: [
			{
				resource_id: "postgres_pkg"
				type:        "require"
			},
		]

		labels: {
			component: "database"
		}
	}

	// 4. PostgreSQL service (requires configs, will restart on config changes)
	postgres_service: {
		type: "linux.service"
		name: "postgresql"

		config: {
			name:    "postgresql"
			state:   "running"
			enabled: true
		}

		target: {
			labels: {
				role: "database"
			}
		}

		dependencies: [
			{
				resource_id: "postgres_pkg"
				type:        "require"
			},
			{
				resource_id: "postgres_config"
				type:        "notify" // Restart service if config changes
			},
			{
				resource_id: "postgres_hba"
				type:        "notify" // Restart service if HBA changes
			},
		]

		labels: {
			component: "database"
		}

		annotations: {
			restart_on_notify: "true"
		}
	}

	// 5. Firewall rule (order dependency - should run after service starts)
	postgres_firewall: {
		type: "linux.firewall"
		name: "postgresql-firewall"

		config: {
			rule:   "allow"
			port:   5432
			proto:  "tcp"
			source: "10.0.0.0/8"
		}

		target: {
			labels: {
				role: "database"
			}
		}

		dependencies: [
			{
				resource_id: "postgres_service"
				type:        "order" // Run after service, but service success not required
			},
		]

		labels: {
			component: "security"
		}
	}

	// 6. Application that depends on database
	app_service: {
		type: "linux.service"
		name: "myapp"

		config: {
			name:    "myapp"
			state:   "running"
			enabled: true
			env: {
				DB_HOST: "localhost"
				DB_PORT: "5432"
				DB_NAME: "myapp"
			}
		}

		target: {
			labels: {
				role: "app"
			}
		}

		dependencies: [
			{
				resource_id: "postgres_service"
				type:        "require" // App requires database to be running
			},
		]

		labels: {
			component: "application"
		}
	}

	// 7. Health check probe (order dependency on app)
	app_health_probe: {
		type: "probe.http"
		name: "app-health"

		config: {
			url:             "http://localhost:8080/health"
			method:          "GET"
			expected_status: 200
			timeout:         "5s"
			interval:        "30s"
		}

		target: {
			labels: {
				role: "app"
			}
		}

		dependencies: [
			{
				resource_id: "app_service"
				type:        "order" // Check after app starts
			},
		]

		labels: {
			component: "monitoring"
		}
	}
}

// Dependency graph visualization (comment only):
//
//   postgres_pkg
//      |
//      +-- (require) --> postgres_config
//      |                       |
//      +-- (require) ------    |
//      |                  |    |
//      +-- (require) --> postgres_hba
//                         |    |
//                         |    +-- (notify)
//                         |    |
//                         +-- (notify)
//                              |
//                              v
//                         postgres_service
//                              |
//                              +-- (order) --> postgres_firewall
//                              |
//                              +-- (require) --> app_service
//                                                     |
//                                                     +-- (order) --> app_health_probe
