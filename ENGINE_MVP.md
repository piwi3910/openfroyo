# 🧊 **OpenFroyo – Infrastructure Orchestration Engine**

### *Declarative, typed, and modular automation for infrastructure and configuration management.*

---

## 📘 1. Overview

**OpenFroyo** is a next-generation Infrastructure-as-Code engine that combines the *declarative state and planning model of Terraform* with the *procedural and configuration capabilities of Ansible* — but modernized:

- Typed configs via **CUE**  
- Light procedural scripting via **Starlark**  
- WASM-based provider system (secure, portable)  
- Ephemeral **micro-runner** for complex local ops  
- Optional agents for pull mode  
- Drift detection, state, and policy enforcement  

---

## 🧱 2. MVP Scope — *What we build now*

| Category | Implement now | Deferred to later |
|-----------|----------------|------------------|
| **Core runtime** | ✅ Go-based planner, DAG engine, scheduler | ➖ Multi-cluster controller |
| **Providers** | ✅ WASM host, linux.pkg/file/service/firewall, probe.http | ➖ Cloud (AWS/GCP), network NOS, Windows |
| **Micro-runner** | ✅ Standard JSON/stdio self-deleting runner | ➖ Persistent agent |
| **Persistence** | ✅ Solo profile: SQLite + FS + embedded queue | ➖ Cluster profile: Postgres, S3, NATS |
| **CLI** | ✅ `froyo` commands (plan, apply, run, drift, onboard) | ➖ Console UI / REST auth |
| **Facts system** | ✅ os.basic, hw.cpu, net.ifaces, pkg.manifest | ➖ cloud.*, k8s.*, network.* |
| **Policy** | ✅ Local OPA enforcement | ➖ Centralized policy service |
| **Auth** | ✅ Local keypair + optional password onboarding | ➖ OIDC / RBAC / multi-user |
| **Observability** | ✅ Logs + OTel traces (stdout/exporter) | ➖ Metrics service + dashboards |
| **GUI** | ❌ *(explicitly excluded for MVP)* | ➖ Web Console (“OpenFroyo Console”) |

---

## ⚙️ 3. Tech Stack (Go-only core)

| Layer | Technology | Role |
|--------|-------------|------|
| Runtime | **Go** | single language for CLI, controller, worker, and micro-runner |
| Plugins | **WASM/WASI** via Wasmtime | safe provider execution |
| Config | **CUE** + **Starlark** | declarative + minimal scripting |
| Queue | Embedded **Badger/Pebble** (“FroyoQueue”) | at-least-once delivery |
| Persistence | **SQLite** (WAL) + FS blobs | local, simple, upgradeable |
| Secrets | **age** keypair | envelope encryption |
| Policy | **OPA (rego)** | policy-as-code enforcement |
| Telemetry | **OpenTelemetry** + Zerolog | structured logs, traces |

---

## 🧩 4. Execution & Orchestration Engine

1. **Evaluate**
   - Parse CUE configs, run Starlark helpers.
   - Validate schema + policy.

2. **Discover**
   - Collect facts via API/SSH/WinRM/micro-runner.

3. **Plan**
   - Compute diffs (`Desired vs. Actual`).
   - Build DAG of Plan Units (PUs).
   - Persist plan & graph (JSON + DOT).

4. **Apply**
   - Execute DAG in parallel respecting `require`/`notify`.
   - Each PU executes provider ops inside WASM sandbox.
   - Complex local ops delegated to **micro-runner**.
   - Update state, log events, record artifacts.

5. **Post-Apply**
   - Trigger handlers/actions (e.g., reload service).
   - Run smoke tests.

6. **Drift**
   - Periodically compare Actual vs. State.
   - Auto-reconcile or open change requests (policy driven).

---

## 🧠 5. Micro-Runner Standard (Agentless+)

- Tiny static Go binary (<10 MB).  
- Runs via SSH/WinRM, communicates **JSON over stdio**.  
- Frames: `READY → CMD → EVENT → DONE/ERROR → EXIT`.  
- Self-deletes; leaves no persistent agent.  
- Commands supported: `exec`, `file.write`, `pkg.ensure`, `sudoers.ensure`, `sshd.harden`, `service.reload`.  
- Verified signature (cosign) + SHA256 before exec.  
- TTL 10 min, per-task timeouts, idempotency keys.

This is **the default path** for any “complex local” operation.

---

## 🔌 6. Plugin / Provider System

**Format:** OCI image  
```
/plugin/plugin.wasm
/plugin/manifest.yaml
/schemas/*.json
/LICENSE, /SBOM.json
```

**Contract:**
```go
Init()  error
Read()  (Actual, error)
Plan()  (Ops, error)
Apply() (Result, error)
Destroy() (Result, error)
```

**Capabilities:** declared in manifest → enforced by host (e.g., `net:outbound`, `fs:temp`, `exec:micro-runner`).  
**Schemas:** JSON Schema for inputs/outputs; validated before execution.  
**Errors:** classified (`transient`, `throttled`, `conflict`, `permanent`).  
**Examples:**  
- `linux.pkg` → ensure package present/absent  
- `linux.service` → enable/start/stop service  
- `probe.http` → check URL 200 OK  

---

## 🗄️ 7. Persistence Layer

### Solo profile (MVP)
- **Metadata & state:** SQLite (WAL) → `./data/openfroyo.db`  
- **Blobs:** local FS → `./data/blobs/`  
- **Events/Audit:** SQLite tables (append-only).  
- **Queue:** embedded FroyoQueue (Badger).  
- **Secrets:** `age` keypair in `./data/keys/agekey`.  

### Cluster profile (later)
- **Postgres**, **S3/MinIO**, **NATS JetStream**, **Vault/KMS**, **Redis**.

### Tables (Solo)
- `runs`, `plan_units`, `events`, `resource_state`, `facts`, `audit`.  
- Advisory locks via SQLite file locks.  
- Backup: `froyo backup` → hot-copy + tar blobs.  
- Restore: `froyo restore`.  

---

## 🔐 8. Onboarding Workflow (MVP)

Command:
```bash
froyo onboard ssh --host 10.0.0.42 --user root --password s3cr3t \
  --key default-ed25519 --create-user froyo \
  --sudo 'NOPASSWD: /usr/bin/systemctl,/usr/bin/apt,/usr/bin/dnf' \
  --lock-down 'disable_password_auth' \
  --labels 'env=dev,role=web'
```

Steps:
1. SSH connect using password.  
2. Upload and run **micro-runner**.  
3. Runner creates user `froyo`, sets authorized_keys, adds sudoers rule.  
4. Optionally disables password auth, restarts sshd safely.  
5. Controller records host fingerprint + key handle → target registered.  
6. Runner self-deletes.  
7. Subsequent runs connect with key-based SSH.

**Config-driven onboarding** (`onboarding.yaml`) supported for batch use.  
**Rollback** available (`froyo onboard rollback --host …`).  

---

## 🧩 9. CLI Reference (Engine)

| Command | Description |
|----------|--------------|
| `froyo init --solo` | Initialize workspace + keys + config |
| `froyo validate` | Validate CUE schema + policies |
| `froyo facts collect --selector …` | Gather typed facts |
| `froyo plan --out plan.json` | Build plan |
| `froyo apply --plan plan.json` | Execute plan |
| `froyo run <action>` | Run action/runbook |
| `froyo drift detect` | Detect config drift |
| `froyo onboard ssh …` | Onboard new host |
| `froyo backup / restore` | Export or restore data |
| `froyo dev up` | Run controller + worker locally |

---

## 🧩 10. Future Modules (post-MVP roadmap)

| Area | Planned Feature |
|------|-----------------|
| **GUI** | “OpenFroyo Console” – web UI + API + dashboards |
| **Distributed mode** | External Postgres, S3, NATS, Vault |
| **Multi-tenant RBAC** | Orgs, workspaces, users, roles |
| **SSO / OIDC** | OAuth2, token policies |
| **Advanced providers** | AWS, GCP, K8s, network NOS, Windows |
| **Agents (pull mode)** | Persistent lightweight agent for edge/offline |
| **Blueprints** | Higher-order modules (K8s clusters, LAMP stacks) |
| **Workflow orchestration** | Change approvals, canaries, CR gates |
| **Analytics** | Drift trends, success rates, MTTR metrics |
| **Console integrations** | Webhooks, GitHub/GitLab CI, ticketing systems |

---

## 🧩 11. Repo Layout (engine-first)

```
openfroyo/
├── cmd/
│   └── froyo/             # controller/worker/cli
├── pkg/
│   ├── engine/            # planner, DAG, scheduler
│   ├── stores/            # sqlite, fs, queue adapters
│   ├── providers/host/    # wasm host runtime
│   ├── transports/        # ssh, winrm, api
│   ├── micro_runner/      # runner client/protocol
│   ├── policy/            # OPA integration
│   ├── telemetry/         # OTel/logging
│   └── api/               # gRPC/REST services
├── providers/             # wasm providers (linux.pkg, etc.)
├── examples/              # apache demo, facts schemas
└── docs/                  # design docs, md guides
```

---

## 🔄 12. Development Milestones

**M1 — Core Engine (weeks 1-3)**  
- CLI, parser, SQLite store, DAG planner.  
- WASM runtime + linux providers.  
- SSH + micro-runner flow.  
- Apache module demo.

**M2 — Reliability & Policy (weeks 4-6)**  
- Retries, idempotency, drift detection.  
- OPA enforcement.  
- Facts TTL + caching.  
- Logs + OTel traces.

**M3 — Polish (weeks 7-8)**  
- Provider SDK + docs.  
- Backup/export/import.  
- Error classes, structured events.  
- Package binaries for all OSes.

**Later:** Console UI, distributed backend, SSO, analytics.

---

## ✅ 13. Philosophy

> *Start local, scale later.*  
> A single binary (`froyo`) can configure your lab.  
> The same engine can later back a full control plane.  
> All while keeping your playbooks, facts, and providers unchanged.
