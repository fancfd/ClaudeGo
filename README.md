# monitor

A system metrics monitoring tool that collects CPU, memory, disk, network, and process data every second and outputs to one or more destinations.

## Usage

```bash
go run . --console                        # print to stdout
go run . --file metrics.jsonl             # append JSON lines to a file
go run . --web                            # serve JSON at http://localhost:8080/metrics
go run . --console --file out.jsonl --web --port 9090  # combine modes
```

At least one output flag is required.

## Output modes

| Flag | Description |
|------|-------------|
| `--console` | Real-time top-like display, refreshed every second (top 10 processes) |
| `--file <path>` | Appends one JSON object per line to the specified file |
| `--web` | Serves the latest metrics as JSON at `GET /metrics` |
| `--port <port>` | Web server port (default: `8080`) |

### Console example

```
System Monitor — 2026-03-20 10:00:00

CPU:    [################                        ]  42.0%
Memory: [####################                    ]  51.3%  (8388 MB / 16384 MB)
Disk:   [########################                ]  60.1%  (150 GB free / 512 GB total)

Network: ↑ 1024 MB sent   ↓ 2048 MB recv

  PID     NAME                  CPU%      MEM%
  ------------------------------------------------
  1234    firefox               5.20%     3.10%
  ...
```

The display clears and redraws each second, similar to `top`.

## Alerts

Warnings are logged to stderr when thresholds are exceeded:

- CPU > 80%
- Memory > 90%
- Disk > 95%

## Build

```bash
go build
go test ./...
```

## Requirements

- Go 1.25+
