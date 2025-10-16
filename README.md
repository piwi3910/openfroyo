# 🧊 OpenFroyo

**Next-generation Infrastructure-as-Code orchestration engine**

OpenFroyo combines the declarative state management of Terraform with the procedural capabilities of Ansible, modernized with typed configurations, WASM-based providers, and a secure agentless architecture.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)]()

## ✨ Key Features

- **🔒 Secure by Design** - WASM-based provider sandbox with capability-based security
- **📝 Typed Configurations** - CUE for schema validation + Starlark for scripting
- **🚀 Agentless** - Ephemeral micro-runner for complex operations, no persistent agents
- **📊 Observable** - Built-in OpenTelemetry tracing, Prometheus metrics, structured logging
- **🔄 Drift Detection** - Automatic detection and optional auto-remediation
- **📜 Policy-Driven** - OPA integration for governance and compliance
- **⚡ Parallel Execution** - DAG-based scheduler with configurable concurrency
- **💾 Simple Deployment** - Single binary with SQLite backend for easy start

## 🚀 Quick Start

### Installation

```bash
# Download binary
curl -L https://github.com/openfroyo/openfroyo/releases/latest/download/froyo-$(uname -s)-$(uname -m) -o froyo
chmod +x froyo
sudo mv froyo /usr/local/bin/

# Or build from source
git clone https://github.com/openfroyo/openfroyo
cd openfroyo
make build
sudo cp bin/froyo /usr/local/bin/
```

### Your First Configuration

Create `webserver.cue`:

```cue
workspace: {
    name: "webserver"
    version: "1.0.0"
    providers: [{name: "linux.pkg", version: ">=1.0.0"}]
}

resources: {
    nginx: {
        type: "linux.pkg::package"
        name: "nginx"
        config: {
            package: "nginx"
            state: "present"
        }
        target: {
            labels: {env: "prod", role: "web"}
        }
    }
}
```

### Initialize and Apply

```bash
# Initialize workspace
froyo init --solo

# Validate configuration
froyo validate webserver.cue

# Generate plan
froyo plan --out plan.json webserver.cue

# Apply configuration
froyo apply --plan plan.json
```

## 📚 Documentation

- **[Architecture Guide](docs/architecture.md)** - System design and components
- **[CLI Reference](cmd/froyo/README.md)** - Complete command documentation
- **[Configuration Guide](docs/configuration.md)** - CUE and Starlark usage
- **[Provider Development](docs/provider-development.md)** - Building WASM providers
- **[Policy Guide](docs/policies/README.md)** - OPA policy management
- **[Examples](examples/)** - Ready-to-use configurations

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    OpenFroyo CLI                             │
└──────────────────┬──────────────────────────────────────────┘
                   │
         ┌─────────┴─────────┐
         ▼                   ▼
   ┌──────────┐      ┌──────────────┐
   │   CUE    │      │  Starlark    │
   │  Parser  │      │  Evaluator   │
   └─────┬────┘      └──────┬───────┘
         │                  │
         └────────┬─────────┘
                  ▼
         ┌────────────────┐
         │  OPA Policy    │
         │    Engine      │
         └────────┬───────┘
                  ▼
         ┌────────────────┐
         │   DAG Planner  │
         │  & Scheduler   │
         └────────┬───────┘
                  ▼
         ┌────────────────┐
         │ WASM Provider  │
         │     Host       │
         └────────┬───────┘
                  │
         ┌────────┴────────┐
         ▼                 ▼
   ┌──────────┐      ┌──────────┐
   │   SSH    │      │  Micro-  │
   │Transport │──────│  Runner  │
   └──────────┘      └──────────┘
```

## 🔧 Core Components

### 6-Phase Execution Workflow

1. **Evaluate** - Parse CUE configs, execute Starlark, validate policies
2. **Discover** - Collect facts from targets via SSH/micro-runner
3. **Plan** - Compute diffs, build dependency DAG
4. **Apply** - Execute plan units in parallel
5. **Post-Apply** - Run handlers, smoke tests
6. **Drift** - Detect and optionally remediate drift

### WASM Provider System

Providers run in a secure WebAssembly sandbox with:
- Capability-based access control
- Memory and timeout limits
- JSON schema validation
- OCI image packaging

### Micro-Runner Protocol

Ephemeral binary for complex operations:
- JSON-over-stdio communication
- Self-deletes after execution (10min TTL)
- Supports: exec, file ops, package management, service control
- <10MB static binary

## 📦 Available Providers

- **linux.pkg** - Package management (apt/dnf/yum/zypper)
- **linux.service** - Systemd service management (coming soon)
- **linux.file** - File and directory operations (coming soon)
- **probe.http** - HTTP health checks (coming soon)

## 🎯 Use Cases

- **Infrastructure Provisioning** - Configure servers, install packages
- **Configuration Management** - Manage files, services, users
- **Compliance Enforcement** - Policy-driven infrastructure
- **Drift Remediation** - Detect and fix configuration drift
- **Multi-Environment Deployment** - Same configs across dev/staging/prod

## 🔐 Security

- **WASM Sandbox** - Providers run in isolated WebAssembly environment
- **Capability System** - Explicit permission for file, network, execution access
- **Agentless** - No persistent agents reduce attack surface
- **SSH Key Auth** - Secure communication with targets
- **Policy Enforcement** - OPA policies block non-compliant changes
- **Audit Logging** - Complete audit trail of all operations

## 📊 Observability

Built-in telemetry with:
- **OpenTelemetry** - Distributed tracing (OTLP, Jaeger, stdout)
- **Prometheus** - Metrics endpoint at `/metrics`
- **Structured Logging** - JSON or console output (zerolog)
- **Event System** - Real-time event streaming

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run linter
make lint

# Run all checks
make check
```

## 🏗️ Building

```bash
# Build CLI
make build

# Build micro-runner
make build-runner

# Build all binaries
make build-all

# Cross-compile for all platforms
make build-cross
```

## 📈 Roadmap

### MVP (Current)
- ✅ Core engine with DAG planner
- ✅ WASM provider host
- ✅ CUE + Starlark configuration
- ✅ SQLite state backend
- ✅ SSH transport + micro-runner
- ✅ OPA policy engine
- ✅ Telemetry infrastructure
- ✅ linux.pkg provider

### Future
- 🔜 Additional providers (file, service, cloud)
- 🔜 Web Console UI
- 🔜 Distributed deployment (Postgres, S3, NATS)
- 🔜 Pull-mode agents for edge/offline
- 🔜 Multi-tenant RBAC
- 🔜 SSO/OIDC integration
- 🔜 Advanced analytics and dashboards

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📄 License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

OpenFroyo builds on these excellent projects:
- [CUE](https://cuelang.org/) - Configuration language
- [Starlark](https://github.com/google/starlark-go) - Embedded scripting
- [Wazero](https://wazero.io/) - WebAssembly runtime
- [OPA](https://www.openpolicyagent.org/) - Policy engine
- [OpenTelemetry](https://opentelemetry.io/) - Observability
- [Cobra](https://cobra.dev/) - CLI framework

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/openfroyo/openfroyo/issues)
- **Discussions**: [GitHub Discussions](https://github.com/openfroyo/openfroyo/discussions)
- **Documentation**: [docs/](docs/)

---

**Built with ❤️ by the OpenFroyo team**
