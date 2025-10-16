// Example supporting multiple distributions with conditional logic

package config

// Common packages across all distributions
resource "git": {
	type: "linux.pkg::package"

	config: {
		package: "git"
		state:   "latest"
	}
}

resource "curl": {
	type: "linux.pkg::package"

	config: {
		package: "curl"
		state:   "present"
	}
}

resource "vim": {
	type: "linux.pkg::package"

	config: {
		package: "vim"
		state:   "present"
	}
}

// Debian/Ubuntu specific (apt)
resource "build_essential_deb": {
	type: "linux.pkg::package"

	config: {
		package: "build-essential"
		state:   "present"
		manager: "apt"
	}

	labels: {
		distro_family: "debian"
	}
}

// RHEL/CentOS specific (dnf/yum)
resource "development_tools_rhel": {
	type: "linux.pkg::package"

	config: {
		package: "gcc"
		state:   "present"
		manager: "dnf"
	}

	labels: {
		distro_family: "rhel"
	}
}

resource "development_tools_make_rhel": {
	type: "linux.pkg::package"

	config: {
		package: "make"
		state:   "present"
		manager: "dnf"
	}

	labels: {
		distro_family: "rhel"
	}

	depends_on: ["development_tools_rhel"]
}

// openSUSE specific (zypper)
resource "patterns_devel_base_suse": {
	type: "linux.pkg::package"

	config: {
		package: "patterns-devel-base"
		state:   "present"
		manager: "zypper"
	}

	labels: {
		distro_family: "suse"
	}
}

// Container runtimes - different package names per distro
resource "docker_ubuntu": {
	type: "linux.pkg::package"

	config: {
		package: "docker.io"
		state:   "present"
		manager: "apt"
	}

	labels: {
		distro:  "ubuntu"
		service: "container-runtime"
	}
}

resource "docker_rhel": {
	type: "linux.pkg::package"

	config: {
		package: "docker"
		state:   "present"
		manager: "dnf"
		options: ["--enablerepo=docker-ce-stable"]
	}

	labels: {
		distro:  "rhel"
		service: "container-runtime"
	}
}
