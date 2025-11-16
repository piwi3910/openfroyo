# Cron Module

Manage cron jobs on Unix-like systems.

## Description

The cron module allows you to add, remove, or modify cron jobs for any user. It supports both traditional cron schedule syntax and special time strings.

## Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `name` | string | Yes | - | Unique name for the cron job (used as identifier) |
| `job` | string | Yes* | - | Command to execute (*required when state=present) |
| `state` | string | No | "present" | "present" to add/update, "absent" to remove |
| `user` | string | No | "root" | User whose crontab to modify |
| `minute` | string | No | "*" | Minute (0-59 or */N) |
| `hour` | string | No | "*" | Hour (0-23 or */N) |
| `day` | string | No | "*" | Day of month (1-31) |
| `month` | string | No | "*" | Month (1-12) |
| `weekday` | string | No | "*" | Day of week (0-7, 0=Sunday) |
| `special_time` | string | No | - | Special time: "reboot", "yearly", "monthly", "weekly", "daily", "hourly" |
| `disabled` | bool | No | false | Comment out the job (keeps it but disabled) |

## Examples

### Daily backup at 2 AM

```yaml
- task: Daily database backup
  module: cron
  vars:
    name: daily-db-backup
    job: /usr/local/bin/backup-database.sh
    minute: "0"
    hour: "2"
    user: root
```

### Hourly log cleanup (using special_time)

```yaml
- task: Hourly log cleanup
  module: cron
  vars:
    name: log-cleanup
    job: /usr/local/bin/cleanup-logs.sh
    special_time: hourly
    user: root
```

### Every 5 minutes monitoring

```yaml
- task: Monitoring check
  module: cron
  vars:
    name: monitoring-check
    job: /usr/local/bin/check-services.sh
    minute: "*/5"
    user: monitoring
```

### Weekly system update

```yaml
- task: Weekly update
  module: cron
  vars:
    name: weekly-update
    job: apt-get update && apt-get upgrade -y
    special_time: weekly
    user: root
```

### Remove a cron job

```yaml
- task: Remove old backup
  module: cron
  vars:
    name: old-backup-job
    state: absent
    user: root
```

### Disabled job (present but commented)

```yaml
- task: Temporarily disabled maintenance
  module: cron
  vars:
    name: disabled-maintenance
    job: /usr/local/bin/maintenance.sh
    minute: "0"
    hour: "3"
    disabled: true
    user: root
```

### Reboot task

```yaml
- task: Run on reboot
  module: cron
  vars:
    name: startup-script
    job: /usr/local/bin/startup.sh
    special_time: reboot
    user: root
```

### Complex schedule - First day of month

```yaml
- task: Monthly report
  module: cron
  vars:
    name: monthly-report
    job: /usr/local/bin/generate-report.sh
    minute: "0"
    hour: "0"
    day: "1"
    user: reports
```

## Schedule Syntax

### Traditional Cron Schedule

```
minute hour day month weekday
  |     |    |    |      |
  |     |    |    |      +---- Day of week (0-7, 0=Sunday)
  |     |    |    +----------- Month (1-12)
  |     |    +---------------- Day of month (1-31)
  |     +--------------------- Hour (0-23)
  +--------------------------- Minute (0-59)
```

Use `*` for "any value" or `*/N` for "every N units".

### Special Time Strings

| Value | Equivalent | Description |
|-------|------------|-------------|
| `@reboot` | - | Run once at startup |
| `@yearly` | `0 0 1 1 *` | Once a year, midnight, Jan 1 |
| `@annually` | `0 0 1 1 *` | Same as @yearly |
| `@monthly` | `0 0 1 * *` | Once a month, midnight, first of month |
| `@weekly` | `0 0 * * 0` | Once a week, midnight on Sunday |
| `@daily` | `0 0 * * *` | Once a day, midnight |
| `@midnight` | `0 0 * * *` | Same as @daily |
| `@hourly` | `0 * * * *` | Once an hour, beginning of hour |

## Implementation

This module uses the shell_exec pattern to manage cron jobs. It:

1. Uses the job `name` as a unique identifier (stored as comment)
2. Preserves other cron jobs in the user's crontab
3. Handles missing crontabs gracefully
4. Updates existing jobs when name matches
5. Removes jobs cleanly (both comment and cron line)
6. Supports disabling jobs by commenting them out
7. Shows the full crontab after operations

## Job Identification

Jobs are identified by their `name` variable, which is stored as a comment line (`# name`) immediately before the cron entry. This allows:

- Safe updates without affecting other jobs
- Clean removal of specific jobs
- Multiple jobs with the same command but different schedules

## Requirements

- `cron` and `crontab` must be installed
- User must have permission to modify the target user's crontab
- For other users' crontabs, sudo access may be required

## Return Values

The module returns a shell_exec command that performs the requested cron operations. The executor handles the actual execution and reports success/failure based on exit codes.

## Notes

- Cron job names must be unique per user
- Job commands should use absolute paths
- Environment variables in cron jobs may differ from interactive shells
- Cron output is typically emailed to the user unless redirected
