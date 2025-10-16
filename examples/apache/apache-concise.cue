// Apache Web Server - Concise Syntax
// This demonstrates the optimized, terse syntax for OpenFroyo

package apache

// Workspace configuration
workspace: {
	name:    "apache-webserver"
	version: "1.0.0"
}

// Target selector
targets: labels: {
	role: "web"
	env:  "production"
}

// Option 1: Namespace-based syntax (MOST CONCISE)
// Package name is the key, state is inferred as "present" by default
linux: pkg: {
	apache2:              {}  // Install apache2 (state: present is default)
	"apache2-utils":      {}  // Install utilities
	"libapache2-mod-ssl": {}  // Install SSL module
}

// Option 2: With explicit state and version
linux: pkg: {
	nginx: {
		state:   "latest"  // Keep updated
		version: ""        // empty = latest
	}
	postgresql: {
		state:   "present"
		version: "14.5"  // Specific version
	}
	apache2: state: "absent"  // Remove package (single-line)
}

// Option 3: Shorthand for common cases
linux: pkg: [
	"nginx",        // string in array = install with defaults
	"postgresql",
	"redis-server",
]

// Option 4: Mixed syntax with states
linux: pkg: {
	install: ["nginx", "postgresql", "redis-server"]
	remove:  ["apache2", "lighttpd"]
	upgrade: ["openssl", "curl"]
}

// Option 5: Full control when needed (backwards compatible)
resources: {
	custom_pkg: {
		id:   "custom-nginx"
		type: "linux.pkg::package"
		config: {
			package:    "nginx"
			state:      "present"
			version:    "1.18.0"
			repository: "nginx-mainline"
			options:    ["--no-install-recommends"]
		}
		dependencies: [{
			resource_id: "openssl-package"
			type:        "require"
		}]
	}
}

// Metadata
metadata: {
	description: "Apache web server - concise syntax demo"
	author:      "OpenFroyo Team"
}
