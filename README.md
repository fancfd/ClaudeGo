# monitor

A system metrics monitoring tool that collects CPU, memory, disk, network, and process data every second and outputs to one or more destinations.

## Usage

```bash
go run . --console                                     # 打印到终端
go run . --file metrics.jsonl                          # 追加写入 JSON Lines 文件
go run . --web                                         # 在 http://localhost:8080/metrics 提供 JSON 接口
go run . --console --file out.jsonl --web --port 9090  # 多种模式同时启用
```

至少需要指定一种输出模式，否则程序退出并报错。

## Output Flags

| Flag | 说明 |
|------|------|
| `--console` | 实时类 top 终端显示，每秒刷新（含 CPU/内存/磁盘进度条及进程列表） |
| `--file <path>` | 将每次采集结果以单行 JSON 对象追加到指定文件（JSON Lines 格式） |
| `--web` | 以 JSON 格式在 `GET /metrics` 暴露最新指标 |
| `--port <port>` | Web 服务器监听端口（默认：`8080`） |

多种模式可同时启用，互不干扰。

## Project Structure

```
.
├── main.go       # 程序入口：数据结构、主循环、告警检测
├── metrics.go    # 系统指标采集（CPU、内存、磁盘、网络、进程）
└── output.go     # 三种输出处理器（终端、文件、Web 服务器）
```

### main.go — 程序入口与数据模型

定义所有数据结构并驱动主循环：

- **数据结构**：`Metrics`、`MemoryMetrics`、`DiskMetrics`、`NetworkMetrics`、`ProcessMetrics`，均支持 JSON 序列化
- **命令行解析**：使用标准库 `flag` 解析 `--console`、`--file`、`--web`、`--port`
- **主循环**：每秒触发一次采集，将结果分发给各输出模块
- **告警检测**：`checkAlerts()` 在指标超阈值时向 stderr 输出告警日志
- **通道设计**：通过容量为 1 的带缓冲通道向 Web 服务器推送指标，采用非阻塞 `select/default` 保证主循环不被阻塞

### metrics.go — 系统指标采集

通过 [`gopsutil/v3`](https://github.com/shirou/gopsutil) 采集各项系统指标：

| 函数 | 说明 |
|------|------|
| `collectMetrics()` | 一次性采集所有指标并返回 `Metrics` 快照 |
| `getTopProcesses(n)` | 遍历所有进程，按 CPU 使用率降序返回前 n 个 |

采集项包括：
- **CPU**：全核平均使用率，采样间隔 1 秒
- **内存**：总量、已用量、使用率（物理 RAM）
- **磁盘**：根文件系统 `/` 的总量、剩余空间、使用率
- **网络**：所有网络接口的累计发送/接收字节数及数据包数
- **进程**：CPU 占用最高的前 10 个进程（PID、名称、CPU%、内存%）

进程采集失败时降级为空列表，不影响其他指标的正常输出。

### output.go — 输出处理器

包含三种独立的输出实现及辅助函数：

| 函数 | 说明 |
|------|------|
| `outputToConsole(metrics)` | 清屏后重绘终端界面，模拟 `top` 的原地刷新效果 |
| `outputToFile(metrics, path)` | 将指标序列化为 JSON 后追加到文件（每行一条记录） |
| `startWebServer(port, ch)` | 启动 Gin HTTP 服务器，`GET /metrics` 返回最新指标 JSON |
| `bar(percent, width)` | 生成文本进度条，用于终端显示 |
| `repeat(ch, n)` | 生成重复字符串的辅助函数 |

Web 服务器在独立 goroutine 中运行，通过通道接收主循环推送的最新指标，HTTP 请求始终读取最近一次采集结果。

## Console Example

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

终端界面每秒清屏重绘，效果类似 `top`。

## Alerts

当指标超过以下阈值时，告警信息会输出到 stderr：

| 指标 | 阈值 |
|------|------|
| CPU 使用率 | > 80% |
| 内存使用率 | > 90% |
| 磁盘使用率 | > 95% |

## Build

```bash
go build        # 编译生成可执行文件
go test ./...   # 运行测试
```

## Requirements

- Go 1.21+
- 依赖库（通过 `go mod tidy` 安装）：
  - [`github.com/shirou/gopsutil/v3`](https://github.com/shirou/gopsutil) — 跨平台系统指标采集
  - [`github.com/gin-gonic/gin`](https://github.com/gin-gonic/gin) — HTTP 服务器
