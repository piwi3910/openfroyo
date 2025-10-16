// Multi-environment configuration example
// This demonstrates environment-specific overrides using CUE

package config

// Base workspace configuration
workspace: {
	name:    "multi-env"
	version: "1.0"

	providers: [
		{name: "linux.pkg", version: ">=1.0.0"},
		{name: "linux.service", version: ">=1.0.0"},
	]
}

// Environment definitions
#Environment: {
	name:      string
	db_host:   string
	db_port:   int | *5432
	app_port:  int
	replicas:  int
	log_level: "debug" | "info" | "warning" | "error"
}

// Environment configurations
environments: {
	dev: #Environment & {
		name:      "development"
		db_host:   "localhost"
		app_port:  3000
		replicas:  1
		log_level: "debug"
	}

	staging: #Environment & {
		name:      "staging"
		db_host:   "staging-db.internal"
		app_port:  8080
		replicas:  2
		log_level: "info"
	}

	prod: #Environment & {
		name:      "production"
		db_host:   "prod-db.internal"
		app_port:  80
		replicas:  5
		log_level: "warning"
	}
}

// Select environment (can be overridden with tags)
selectedEnv: environments.dev

// Resource template
#AppResource: {
	_env: #Environment

	type: "linux.service"
	name: string

	config: {
		name:    string
		state:   "running"
		enabled: true
		env: {
			DB_HOST:   _env.db_host
			DB_PORT:   "\(_env.db_port)"
			APP_PORT:  "\(_env.app_port)"
			LOG_LEVEL: _env.log_level
		}
	}

	target: {
		labels: {
			env: _env.name
		}
	}
}

// Resources with environment-specific configuration
resources: {
	app_service: #AppResource & {
		_env: selectedEnv
		name: "myapp"
		config: name: "myapp"

		labels: {
			component: "application"
			managed:   "openfroyo"
		}
	}
}
