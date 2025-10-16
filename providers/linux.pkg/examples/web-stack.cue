// Complete web application stack example

package config

// Web Server
resource "nginx": {
	type: "linux.pkg::package"

	config: {
		package: "nginx"
		state:   "latest"
		manager: "apt"
	}

	labels: {
		tier:        "frontend"
		environment: "production"
	}
}

// Application Runtime
resource "nodejs": {
	type: "linux.pkg::package"

	config: {
		package: "nodejs"
		state:   "present"
		version: "18.17.0-1nodesource1"
		manager: "apt"
	}

	labels: {
		tier:        "application"
		environment: "production"
	}
}

// Database
resource "postgresql": {
	type: "linux.pkg::package"

	config: {
		package: "postgresql-14"
		state:   "present"
		manager: "apt"
	}

	labels: {
		tier:        "database"
		environment: "production"
	}
}

// Redis Cache
resource "redis": {
	type: "linux.pkg::package"

	config: {
		package: "redis-server"
		state:   "present"
		manager: "apt"
	}

	labels: {
		tier:        "cache"
		environment: "production"
	}
}

// SSL/TLS Support
resource "certbot": {
	type: "linux.pkg::package"

	config: {
		package: "certbot"
		state:   "latest"
		manager: "apt"
	}

	labels: {
		tier:        "security"
		environment: "production"
	}

	depends_on: ["nginx"]
}

// Monitoring Agent
resource "prometheus_node_exporter": {
	type: "linux.pkg::package"

	config: {
		package: "prometheus-node-exporter"
		state:   "present"
		manager: "apt"
	}

	labels: {
		tier:        "monitoring"
		environment: "production"
	}
}
