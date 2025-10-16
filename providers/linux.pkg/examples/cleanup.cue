// Example demonstrating package removal and cleanup

package config

// Remove Apache if it exists (switching to NGINX)
resource "remove_apache": {
	type: "linux.pkg::package"

	config: {
		package: "apache2"
		state:   "absent"
		manager: "apt"
	}

	annotations: {
		reason: "Migrating from Apache to NGINX"
	}
}

// Remove unused packages
resource "remove_sendmail": {
	type: "linux.pkg::package"

	config: {
		package: "sendmail"
		state:   "absent"
	}

	annotations: {
		reason: "Using external mail service"
	}
}

resource "remove_old_python": {
	type: "linux.pkg::package"

	config: {
		package: "python2"
		state:   "absent"
	}

	annotations: {
		reason: "Python 2 end of life"
	}
}

// Remove development tools from production
resource "remove_gcc": {
	type: "linux.pkg::package"

	config: {
		package: "gcc"
		state:   "absent"
	}

	annotations: {
		reason: "Development tools not needed in production"
	}

	labels: {
		environment: "production"
	}
}

resource "remove_make": {
	type: "linux.pkg::package"

	config: {
		package: "make"
		state:   "absent"
	}

	labels: {
		environment: "production"
	}
}
