# cronwatch

Lightweight daemon that monitors cron job execution times and alerts on drift or failures.

---

## Installation

```bash
go install github.com/yourname/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronwatch.git && cd cronwatch && go build -o cronwatch .
```

---

## Usage

Define your monitored jobs in a `cronwatch.yaml` config file:

```yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    tolerance: 5m
    alert:
      email: ops@example.com

  - name: hourly-sync
    schedule: "0 * * * *"
    tolerance: 2m
    alert:
      webhook: https://hooks.example.com/alert
```

Start the daemon:

```bash
cronwatch --config cronwatch.yaml
```

Wrap your existing cron commands to report execution:

```bash
# In your crontab
0 2 * * * cronwatch run --job daily-backup -- /usr/local/bin/backup.sh
```

cronwatch will alert you if a job:
- Fails to execute within the expected window (drift)
- Exits with a non-zero status code
- Exceeds its expected runtime

---

## Configuration Options

| Field       | Description                              | Default |
|-------------|------------------------------------------|---------|
| `schedule`  | Standard cron expression                 | —       |
| `tolerance` | Allowed drift before alerting            | `1m`    |
| `alert`     | Notification target (email, webhook, etc)| —       |

---

## License

MIT © 2024 yourname