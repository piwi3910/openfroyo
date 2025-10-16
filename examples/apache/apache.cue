// Apache Web Server Demo Configuration
// This demonstrates a complete web server setup using OpenFroyo

package apache

import "strings"

// Workspace configuration
workspace: {
	name:    "apache-webserver"
	version: "1.0.0"
	providers: [{
		name:    "linux.pkg"
		version: ">=1.0.0"
	}]
}

// Target selector - apply to web servers
targets: {
	labels: {
		role: "web"
		env:  "production"
	}
}

// Resources to manage
resources: {
	// 1. Install Apache package
	apache_pkg: {
		id:   "apache2-package"
		type: "linux.pkg::package"
		name: "apache2"
		config: {
			package: "apache2"
			state:   "present"
		}
		labels: {
			component: "webserver"
			tier:      "frontend"
		}
		target: targets
	}

	// 2. Install Apache utilities
	apache_utils: {
		id:   "apache2-utils"
		type: "linux.pkg::package"
		name: "apache2-utils"
		config: {
			package: "apache2-utils"
			state:   "present"
		}
		dependencies: [{
			resource_id: "apache2-package"
			type:        "require"
		}]
		labels: {
			component: "webserver"
			tier:      "frontend"
		}
		target: targets
	}

	// 3. Install mod_ssl for HTTPS
	mod_ssl: {
		id:   "apache2-ssl"
		type: "linux.pkg::package"
		name: "mod_ssl"
		config: {
			package: "libapache2-mod-ssl"
			state:   "present"
		}
		dependencies: [{
			resource_id: "apache2-package"
			type:        "require"
		}]
		labels: {
			component: "webserver"
			tier:      "frontend"
			security:  "tls"
		}
		target: targets
	}
}

// Metadata for documentation
metadata: {
	description: "Apache web server with SSL support"
	author:      "OpenFroyo Team"
	tags: ["web", "apache", "demo"]
	documentation: """
		This configuration installs Apache HTTP Server with:
		- Apache 2.4 base package
		- Apache utilities (ab, htpasswd, etc.)
		- mod_ssl for HTTPS support

		All packages installed on hosts with labels:
		- role=web
		- env=production
		"""
}
