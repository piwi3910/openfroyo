// Basic package installation example

package config

// Install NGINX web server
resource "nginx_package": {
	type: "linux.pkg::package"

	config: {
		package: "nginx"
		state:   "present"
	}
}

// Install PostgreSQL database
resource "postgresql_package": {
	type: "linux.pkg::package"

	config: {
		package: "postgresql-14"
		state:   "present"
	}
}

// Install Python 3 and pip
resource "python_package": {
	type: "linux.pkg::package"

	config: {
		package: "python3"
		state:   "latest"
	}
}

resource "python_pip": {
	type: "linux.pkg::package"

	config: {
		package: "python3-pip"
		state:   "present"
	}

	// Ensure python3 is installed first
	depends_on: ["python_package"]
}
