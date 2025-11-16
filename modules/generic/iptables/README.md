# Iptables Module

Manage iptables firewall rules on Linux systems. Automatically persists rules based on the distribution.

## Requirements

- iptables must be installed
- Root or sudo access (iptables commands require elevated privileges)
- For persistence: `netfilter-persistent` (Debian/Ubuntu) or standard iptables-save location

## Variables

- `chain` (optional): Chain name (default: "INPUT") - INPUT, OUTPUT, FORWARD, or custom chain
- `protocol` (optional): Protocol (tcp, udp, icmp, all, etc.)
- `destination_port` (optional): Destination port number or range
- `source` (optional): Source IP address or network (CIDR notation)
- `destination` (optional): Destination IP address or network (CIDR notation)
- `jump` (optional): Target action (default: "ACCEPT") - ACCEPT, DROP, REJECT, LOG, etc.
- `state` (optional): Rule state (default: "present") - "present" or "absent"
- `table` (optional): Table name (default: "filter") - filter, nat, mangle, raw

## States

- **present**: Add the rule if it doesn't exist
- **absent**: Remove the rule if it exists

## Examples

### Allow incoming HTTP traffic

```yaml
- name: Allow HTTP
  module: generic/iptables
  hosts:
    - web-server
  vars:
    chain: INPUT
    protocol: tcp
    destination_port: 80
    jump: ACCEPT
    state: present
```

### Allow SSH from specific network

```yaml
- name: Allow SSH from admin network
  module: generic/iptables
  hosts:
    - "@group:servers"
  vars:
    chain: INPUT
    protocol: tcp
    destination_port: 22
    source: 192.168.1.0/24
    jump: ACCEPT
    state: present
```

### Block traffic from specific IP

```yaml
- name: Block malicious IP
  module: generic/iptables
  hosts:
    - web-server
  vars:
    chain: INPUT
    source: 10.0.0.50
    jump: DROP
    state: present
```

### Allow outbound HTTPS

```yaml
- name: Allow outbound HTTPS
  module: generic/iptables
  hosts:
    - app-server
  vars:
    chain: OUTPUT
    protocol: tcp
    destination_port: 443
    jump: ACCEPT
    state: present
```

### Remove a rule

```yaml
- name: Remove old firewall rule
  module: generic/iptables
  hosts:
    - server-01
  vars:
    chain: INPUT
    protocol: tcp
    destination_port: 8080
    jump: ACCEPT
    state: absent
```

### NAT rule for port forwarding

```yaml
- name: Forward port 80 to 8080
  module: generic/iptables
  hosts:
    - gateway
  vars:
    table: nat
    chain: PREROUTING
    protocol: tcp
    destination_port: 80
    jump: REDIRECT
    state: present
```

## How It Works

1. **Verification**: Checks if iptables is installed
2. **Rule Check**: Verifies if the rule already exists using `iptables -C`
3. **Add/Remove**: Adds or removes the rule based on state and current configuration
4. **Persistence**: Automatically saves rules using the appropriate method:
   - Debian/Ubuntu: `netfilter-persistent save` or `/etc/iptables/rules.v4`
   - RHEL/CentOS: `/etc/sysconfig/iptables`
5. **Result**: Returns success/failure status and output

## Common Chains

- **INPUT**: Rules for incoming traffic to the system
- **OUTPUT**: Rules for outgoing traffic from the system
- **FORWARD**: Rules for traffic being routed through the system

## Common Tables

- **filter**: Default table for packet filtering (INPUT, OUTPUT, FORWARD chains)
- **nat**: Network Address Translation (PREROUTING, POSTROUTING, OUTPUT chains)
- **mangle**: Packet alteration (all chains)
- **raw**: Connection tracking exemptions (PREROUTING, OUTPUT chains)

## Common Targets

- **ACCEPT**: Allow the packet
- **DROP**: Silently discard the packet
- **REJECT**: Reject the packet and send an error response
- **LOG**: Log the packet and continue processing
- **REDIRECT**: Redirect to a local port (NAT table)
- **MASQUERADE**: Dynamic source NAT (NAT table)

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
make build
```

This produces `wasm/iptables.wasm`.
