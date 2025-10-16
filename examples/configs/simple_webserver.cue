// Simple web server configuration example
// This demonstrates a basic Apache web server setup with dependencies

package config

// Workspace configuration
workspace: {
	name:    "simple-webserver"
	version: "1.0"

	providers: [
		{
			name:    "linux.pkg"
			version: ">=1.0.0"
		},
		{
			name:    "linux.service"
			version: ">=1.0.0"
		},
		{
			name:    "linux.file"
			version: ">=1.0.0"
		},
	]

	backend: {
		type: "solo"
		path: "./data"
	}

	policy: {
		enabled:      true
		mode:         "enforcing"
		on_violation: "fail"
	}
}

// Resource definitions
resources: {
	// Install Apache package
	apache_pkg: {
		type: "linux.pkg"
		name: "apache2"

		config: {
			package: "apache2"
			state:   "present"
			version: "latest"
		}

		target: {
			labels: {
				env:  "prod"
				role: "web"
			}
		}

		labels: {
			component: "webserver"
			managed:   "openfroyo"
		}

		annotations: {
			description: "Apache HTTP Server package"
			owner:       "ops-team"
		}
	}

	// Configure Apache
	apache_config: {
		type: "linux.file"
		name: "apache-config"

		config: {
			path:    "/etc/apache2/sites-available/default.conf"
			content: """
				<VirtualHost *:80>
				    ServerName example.com
				    DocumentRoot /var/www/html

				    <Directory /var/www/html>
				        Options -Indexes +FollowSymLinks
				        AllowOverride All
				        Require all granted
				    </Directory>

				    ErrorLog ${APACHE_LOG_DIR}/error.log
				    CustomLog ${APACHE_LOG_DIR}/access.log combined
				</VirtualHost>
				"""
			owner:       "root"
			group:       "root"
			mode:        "0644"
			create_dirs: true
		}

		target: {
			labels: {
				env:  "prod"
				role: "web"
			}
		}

		dependencies: [
			{
				resource_id: "apache_pkg"
				type:        "require"
			},
		]

		labels: {
			component: "webserver"
			managed:   "openfroyo"
		}
	}

	// Create document root
	document_root: {
		type: "linux.file"
		name: "document-root"

		config: {
			path:  "/var/www/html"
			state: "directory"
			owner: "www-data"
			group: "www-data"
			mode:  "0755"
		}

		target: {
			labels: {
				env:  "prod"
				role: "web"
			}
		}

		dependencies: [
			{
				resource_id: "apache_pkg"
				type:        "require"
			},
		]

		labels: {
			component: "webserver"
			managed:   "openfroyo"
		}
	}

	// Manage Apache service
	apache_service: {
		type: "linux.service"
		name: "apache2"

		config: {
			name:    "apache2"
			state:   "running"
			enabled: true
		}

		target: {
			labels: {
				env:  "prod"
				role: "web"
			}
		}

		dependencies: [
			{
				resource_id: "apache_pkg"
				type:        "require"
			},
			{
				resource_id: "apache_config"
				type:        "notify"
			},
		]

		labels: {
			component: "webserver"
			managed:   "openfroyo"
		}

		annotations: {
			restart_on_change: "true"
		}
	}
}
