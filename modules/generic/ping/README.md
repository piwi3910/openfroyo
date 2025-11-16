# ping Module

Test network connectivity using ICMP ping requests.

## Purpose

The ping module sends ICMP echo requests to test network connectivity and measure round-trip time (RTT) to remote hosts. It supports both IPv4 and IPv6 and can optionally bind to a specific network interface.

## Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| host | string | Yes | 8.8.8.8 | Hostname or IP address to ping |
| count | int | No | 4 | Number of ping packets to send |
| timeout | int | No | 5 | Timeout in seconds for each ping |
| interface | string | No | "" | Network interface to use (e.g., eth0, wlan0) |

## Examples

### Basic ping
```yaml
- task: Check Google DNS connectivity
  module: ping
  vars:
    host: 8.8.8.8
```

### Custom packet count and timeout
```yaml
- task: Ping with more packets
  module: ping
  vars:
    host: example.com
    count: 10
    timeout: 3
```

### Ping via specific interface
```yaml
- task: Test connectivity via eth0
  module: ping
  vars:
    host: 192.168.1.1
    interface: eth0
```

### IPv6 ping
```yaml
- task: Ping IPv6 address
  module: ping
  vars:
    host: 2001:4860:4860::8888
    count: 5
```

## Output Facts

The module returns the following facts:

- `exit_code`: Exit code of the ping command (0 = success)
- `packets_transmitted`: Number of packets sent
- `packets_received`: Number of packets received
- `packet_loss_percent`: Percentage of packets lost
- `rtt_min`: Minimum round-trip time in milliseconds (if available)
- `rtt_avg`: Average round-trip time in milliseconds (if available)
- `rtt_max`: Maximum round-trip time in milliseconds (if available)
- `output`: Raw output from the ping command

## Implementation Details

This module uses the `shell_exec` pattern to execute ping commands on remote hosts. The ping command varies slightly by platform but generally uses:

```bash
ping -c {count} -W {timeout} {host}
```

With interface binding:
```bash
ping -c {count} -W {timeout} -I {interface} {host}
```

The module parses the ping output to extract statistics including packet loss and RTT values.

## Platform Compatibility

- **Linux**: Full support (IPv4 and IPv6)
- **macOS**: Full support (IPv4 and IPv6)
- **BSD**: Full support (IPv4 and IPv6)

Note: The module requires the `ping` command to be available on the target system (standard on all Unix-like systems).

## Return Status

- **ok**: Ping command executed successfully and host is reachable
- **failed**:
  - Required variable 'host' is missing or empty
  - Invalid count or timeout values
  - Ping command failed (100% packet loss or command error)

The actual success/failure is determined by the exit code and packet loss percentage in the command output.
