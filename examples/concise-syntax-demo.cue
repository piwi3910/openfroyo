// Concise Syntax Demo for OpenFroyo
// This demonstrates the new terse syntax for defining resources

package demo

// Workspace configuration
workspace: {
	name:    "concise-demo"
	version: "1.0.0"
}

// =============================================================================
// CONCISE SYNTAX - Package Management
// =============================================================================

// Basic package installation (defaults to state: "present")
linux: pkg: {
	nginx:       {}  // Install nginx with defaults
	postgresql:  {}  // Install postgresql with defaults
	redis:       {}  // Install redis with defaults
}

// Package installation with specific versions
linux: pkg: {
	python3: {
		state:   "present"
		version: "3.11"
	}
	nodejs: {
		state:   "latest"  // Always keep updated
	}
	apache2: {
		state: "absent"  // Ensure removed
	}
}

// =============================================================================
// COMPARISON: Verbose vs Concise
// =============================================================================

// OLD VERBOSE SYNTAX (7 lines per resource):
resources: old_way: {
	id:   "nginx-package"
	type: "linux.pkg::package"
	name: "nginx"
	config: {
		package: "nginx"
		state:   "present"
	}
}

// NEW CONCISE SYNTAX (1 line per resource):
linux: pkg: nginx: {}  // Same result!

// =============================================================================
// MIXED USAGE - Both syntaxes work together
// =============================================================================

// Use concise syntax for simple cases
linux: pkg: {
	curl:  {}
	wget:  {}
	htop:  {}
	tmux:  {}
}

// Use verbose syntax when you need full control
resources: custom_postgresql: {
	id:   "postgresql-custom"
	type: "linux.pkg::package"
	name: "postgresql"
	config: {
		package:    "postgresql-14"
		state:      "present"
		version:    "14.5-1ubuntu1"
		repository: "postgresql-main"
		options:    ["--no-install-recommends"]
	}
	labels: {
		component: "database"
		tier:      "backend"
	}
	dependencies: [{
		resource_id: "linux-pkg-postgresql"
		type:        "require"
	}]
}

// Metadata
metadata: {
	description: "Demo showing concise vs verbose syntax"
	author:      "OpenFroyo Team"
	tags: ["demo", "syntax", "comparison"]
}
