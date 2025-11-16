# OpenFroyo MVP Specification (Agentless + WASM + Stack-Based Execution)

## Overview
OpenFroyo is an agentless, Go‑based automation and orchestration framework inspired by Ansible and Terraform, using WebAssembly (WASM) as its execution engine for remote actions.  
This MVP describes:

- Project structure  
- Inventory model (hosts + groups)  
- Stack execution model  
- Task blocks  
- Module structure  
- WASM execution model  
- SSH-based push execution (no agents)

---

# 1. Project Structure

```
openfroyo-project/
  inventory/
    hosts.ofy.yml
    groups.ofy.yml
  stacks/
    web_stack.ofy.yml
  modules/
    webserver/
      module.ofy.yml
      defaults.ofy.yml
      vars.ofy.yml
      wasm/
        install_nginx.wasm
        write_nginx_config.wasm
    postgres/
      module.ofy.yml
      wasm/
  modules/system/
    wasm/
      service_control.wasm
```

---

# 2. Inventory Files

## 2.1 `inventory/hosts.ofy.yml`

Defines SSH connection details for each host.

```yaml
hosts:
  web-01:
    hostname: "10.0.0.11"
    port: 22
    user: "ubuntu"
    auth:
      method: "ssh_key"
      key_path: "~/.ssh/id_rsa"
    become:
      enabled: true
      method: "sudo"

  web-02:
    hostname: "10.0.0.12"
    user: "ubuntu"
    auth:
      method: "ssh_agent"

  db-01:
    hostname: "10.0.1.10"
    user: "postgres"
    auth:
      method: "ssh_key"
      key_path: "~/.ssh/postgres_rsa"
```

---

## 2.2 `inventory/groups.ofy.yml`

Defines host groups.

```yaml
groups:
  web:
    hosts: ["web-01", "web-02"]

  db:
    hosts: ["db-01"]

  all:
    children: ["web", "db"]
```

---

# 3. Stack File Format (`stacks/*.ofy.yml`)

A stack describes **orchestration** by defining:

- Inventory files  
- Default execution parameters  
- An ordered `run:` list of modules and task blocks  

## Example: `stacks/web_stack.ofy.yml`

```yaml
project: "my-web-stack"

inventory:
  hosts_file: "inventory/hosts.ofy.yml"
  groups_file: "inventory/groups.ofy.yml"

defaults:
  hosts: "@group:web"
  strategy: "parallel"
  max_parallel: 2

run:
  - name: "Webserver base"
    module: "webserver"
    vars:
      domain: "staging.example.com"

  - name: "Database layer"
    module: "postgres"
    hosts: "@group:db"
    vars:
      db_name: "myapp"
      db_user: "myapp"
      db_password: "supersecret"

  - name: "Deploy app code"
    hosts: "@group:web"
    strategy: "parallel"
    task:
      - name: "Upload app"
        run_wasm:
          module: "modules/deployer/wasm/upload_app.wasm"
          vars:
            repo: "gh:org/myapp"
            version: "v1.0.0"

      - name: "Restart nginx"
        run_wasm:
          module: "modules/system/wasm/service_control.wasm"
          vars:
            name: "nginx"
            state: "restarted"

  - name: "Harden OS"
    module: "os_hardening"
    hosts: "@group:all"
```

---

# 4. `run:` Entry Types

Each item in `run:` can be:

### 4.1 Module Invocation
```yaml
- name: "Webserver base"
  module: "webserver"
  vars:
    domain: "staging.example.com"
```

### 4.2 Task Block
A task block is a list of steps executed on specified hosts.

```yaml
- name: "Deploy app code"
  hosts: "@group:web"
  task:
    - name: "Upload app"
      run_wasm:
        module: "modules/deployer/wasm/upload_app.wasm"
        vars: { ... }

    - name: "Restart nginx"
      run_wasm:
        module: "modules/system/wasm/service_control.wasm"
        vars: { ... }
```

---

# 5. Module Format (`modules/<name>/module.ofy.yml`)

Modules are reusable units similar to Ansible roles.

```yaml
module:
  name: "webserver"
  description: "Provision nginx webservers"

  default_hosts: "@group:web"

  steps:
    - name: "Install nginx"
      run_wasm:
        module: "./wasm/install_nginx.wasm"
        vars:
          package: "{{ var.package | default('nginx') }}"

    - name: "Write nginx config"
      run_wasm:
        module: "./wasm/write_nginx_config.wasm"
        vars:
          path: "{{ var.config_path | default('/etc/nginx/sites-enabled/default') }}"
          domain: "{{ var.domain | default('example.com') }}"
```

### Optional: `defaults.ofy.yml`
```yaml
defaults:
  package: "nginx"
  config_path: "/etc/nginx/sites-enabled/default"
  domain: "example.com"
```

### Optional: `vars.ofy.yml`
Project-override variables.

```yaml
vars:
  domain: "prod.example.com"
```

---

# 6. WASM Execution Model

### 6.1 Runner Flow (over SSH)

For each step:

1. SSH to host  
2. Ensure `froyo-runner` binary exists  
3. Upload WASM module if missing or version changed  
4. Execute runner:

```
froyo-runner   --module <path>.wasm   --input-base64 "<JSON>"   --timeout <sec>
```

5. Read JSON output  
6. Mark step as: `ok` / `changed` / `failed`

---

# 7. WASM Module Contract

### 7.1 WASM Input (JSON)
```json
{
  "vars": { ... },
  "context": {
    "host": "web-01",
    "task_name": "Install nginx"
  }
}
```

### 7.2 WASM Output (JSON)
```json
{
  "status": "changed",
  "message": "nginx installed",
  "facts": {
    "nginx_version": "1.25.0"
  }
}
```

### 7.3 Allowed Status Values
- `"ok"`
- `"changed"`
- `"failed"`

---

# 8. Host API (Exposed to WASM)

- `host_log(level, msg)`
- `host_exec(cmd, args, timeout)`
- `host_file_read(path)`
- `host_file_write(path, contents, mode)`
- `host_file_stat(path)`
- `host_http_request(method, url, headers, body)`
- `host_env_get(name)`

---

# 9. Execution Behaviors

### 9.1 Parallelism
Controlled globally or per-run entry.

### 9.2 Strategy
- `parallel`
- `serial`

### 9.3 Failure Modes
- `continue_on_error: false` (default)
- `max_fail_percentage: X`

---

# 10. CLI Usage

### Run a stack
```
froyo apply stacks/web_stack.ofy.yml
```

### Dry run
```
froyo plan stacks/web_stack.ofy.yml
```

### Run until a step
```
froyo apply stacks/web_stack.ofy.yml --until "Database layer"
```

### Run starting from a step
```
froyo apply stacks/web_stack.ofy.yml --from "Deploy app code"
```

---

# 11. MVP Goals

### Implement:
- Stack parser
- Inventory loader
- Ordered `run:` execution engine
- SSH executor
- WASM runner (Go)
- Host API
- Module loader
- Task block handler

### Out of scope for MVP:
- Conditionals (`when:`)
- Loops (`loop:`)
- Pull/agent mode
- Diff mode for resources
- Providers (Terraform-style)

---

# END OF DOCUMENT
