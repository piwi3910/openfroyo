// Configuration with Starlark integration
// This demonstrates using Starlark for procedural resource generation

package config

workspace: {
	name:    "starlark-demo"
	version: "1.0"

	providers: [
		{name: "linux.pkg", version: ">=1.0.0"},
		{name: "linux.service", version: ">=1.0.0"},
	]

	// Variables accessible in Starlark
	variables: {
		server_count:  3
		base_port:     8080
		domain_suffix: "example.com"
	}
}

// Starlark function to generate multiple servers
// This would be evaluated by the StarlarkEvaluator
#GenerateServers: """
	def generate_servers(count, base_port, domain_suffix):
	    servers = {}
	    for i in range(count):
	        server_id = "app_server_" + str(i)
	        servers[server_id] = {
	            "type": "linux.service",
	            "name": "app-server-" + str(i),
	            "config": {
	                "name": "app-" + str(i),
	                "state": "running",
	                "enabled": True,
	                "env": {
	                    "SERVER_ID": str(i),
	                    "SERVER_PORT": str(base_port + i),
	                    "SERVER_NAME": "server-" + str(i) + "." + domain_suffix,
	                }
	            },
	            "target": {
	                "labels": {
	                    "role": "app",
	                    "instance": str(i),
	                }
	            },
	            "labels": {
	                "component": "application",
	                "server_id": str(i),
	            }
	        }
	    return servers

	# Execute and return generated servers
	result = generate_servers(server_count, base_port, domain_suffix)
	"""

// Static resources
resources: {
	// Load balancer configuration
	load_balancer: {
		type: "linux.pkg"
		name: "haproxy"

		config: {
			package: "haproxy"
			state:   "present"
			version: "latest"
		}

		target: {
			labels: {
				role: "loadbalancer"
			}
		}

		labels: {
			component: "loadbalancer"
			managed:   "openfroyo"
		}
	}

	lb_service: {
		type: "linux.service"
		name: "haproxy"

		config: {
			name:    "haproxy"
			state:   "running"
			enabled: true
		}

		target: {
			labels: {
				role: "loadbalancer"
			}
		}

		dependencies: [
			{
				resource_id: "load_balancer"
				type:        "require"
			},
		]

		labels: {
			component: "loadbalancer"
			managed:   "openfroyo"
		}
	}

	// The generated app servers would be merged here
	// by the Starlark evaluator processing #GenerateServers
}
