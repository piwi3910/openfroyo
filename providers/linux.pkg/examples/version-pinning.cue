// Example demonstrating version pinning and specific versions

package config

// Pin specific version of Docker
resource "docker_ce": {
	type: "linux.pkg::package"

	config: {
		package: "docker-ce"
		state:   "present"
		version: "5:24.0.5-1~ubuntu.22.04~jammy"
		manager: "apt"
		options: ["--allow-downgrades"]
	}

	annotations: {
		reason: "Pinned to specific version for production stability"
	}
}

// Pin Kubernetes components
resource "kubelet": {
	type: "linux.pkg::package"

	config: {
		package: "kubelet"
		state:   "present"
		version: "1.28.2-00"
		manager: "apt"
	}

	annotations: {
		reason: "Match Kubernetes cluster version"
	}
}

resource "kubeadm": {
	type: "linux.pkg::package"

	config: {
		package: "kubeadm"
		state:   "present"
		version: "1.28.2-00"
		manager: "apt"
	}

	depends_on: ["kubelet"]
}

resource "kubectl": {
	type: "linux.pkg::package"

	config: {
		package: "kubectl"
		state:   "present"
		version: "1.28.2-00"
		manager: "apt"
	}

	depends_on: ["kubelet"]
}

// Pin Java version
resource "openjdk_17": {
	type: "linux.pkg::package"

	config: {
		package: "openjdk-17-jdk"
		state:   "present"
		version: "17.0.8+7-1ubuntu1~22.04"
		manager: "apt"
	}

	annotations: {
		reason: "Application requires Java 17 specifically"
	}
}
