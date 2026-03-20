package main

import (
	"flag"
	"log"
	"time"
)

// MemoryMetrics 保存内存使用情况的快照
type MemoryMetrics struct {
	Total       uint64  `json:"total"`        // 物理内存总量（字节）
	Used        uint64  `json:"used"`         // 已使用内存（字节）
	UsedPercent float64 `json:"used_percent"` // 内存使用率（百分比）
}

// DiskMetrics 保存根文件系统的磁盘使用情况
type DiskMetrics struct {
	Total       uint64  `json:"total"`        // 磁盘总容量（字节）
	Free        uint64  `json:"free"`         // 剩余可用空间（字节）
	UsedPercent float64 `json:"used_percent"` // 磁盘使用率（百分比）
}

// NetworkMetrics 保存所有网络接口的累计 I/O 统计
type NetworkMetrics struct {
	BytesSent   uint64 `json:"bytes_sent"`   // 累计发送字节数
	BytesRecv   uint64 `json:"bytes_recv"`   // 累计接收字节数
	PacketsSent uint64 `json:"packets_sent"` // 累计发送数据包数
	PacketsRecv uint64 `json:"packets_recv"` // 累计接收数据包数
}

// ProcessMetrics 保存单个进程的资源使用信息
type ProcessMetrics struct {
	PID           int32   `json:"pid"`            // 进程 ID
	Name          string  `json:"name"`           // 进程名称
	CPUPercent    float64 `json:"cpu_percent"`    // CPU 使用率（百分比）
	MemoryPercent float32 `json:"memory_percent"` // 内存使用率（百分比）
}

// Metrics 是每次采集周期的完整系统指标快照
type Metrics struct {
	Timestamp time.Time        `json:"timestamp"` // 采集时间
	CPU       float64          `json:"cpu_percent"`
	Memory    MemoryMetrics    `json:"memory"`
	Disk      DiskMetrics      `json:"disk"`
	Network   NetworkMetrics   `json:"network"`
	Processes []ProcessMetrics `json:"processes"` // CPU 占用最高的前 N 个进程
}

func main() {
	// 解析命令行参数，至少需要启用一种输出模式
	consoleFlag := flag.Bool("console", false, "Enable console output")
	fileFlag := flag.String("file", "", "File path for JSON output (appends each metric as a line)")
	webFlag := flag.Bool("web", false, "Enable web server for metrics")
	portFlag := flag.String("port", "8080", "Port for web server")
	flag.Parse()

	if !*consoleFlag && *fileFlag == "" && !*webFlag {
		log.Fatal("At least one output mode must be enabled: --console, --file, or --web")
	}

	// 带缓冲的通道，用于将最新指标推送给 Web 服务器
	// 容量为 1：主循环非阻塞发送，Web 服务器始终读取最新值
	metricsChan := make(chan Metrics, 1)

	if *webFlag {
		// Web 服务器在独立 goroutine 中运行，不阻塞采集循环
		go startWebServer(*portFlag, metricsChan)
	}

	// 每秒触发一次采集
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics, err := collectMetrics()
		if err != nil {
			log.Printf("Error collecting metrics: %v", err)
			continue
		}

		// 检查是否触发阈值告警（CPU >80%、内存 >90%、磁盘 >95%）
		checkAlerts(metrics)

		if *consoleFlag {
			outputToConsole(metrics)
		}

		if *fileFlag != "" {
			if err := outputToFile(metrics, *fileFlag); err != nil {
				log.Printf("Error writing to file: %v", err)
			}
		}

		if *webFlag {
			// 非阻塞发送：若通道已满则丢弃本次更新，避免主循环阻塞
			select {
			case metricsChan <- metrics:
			default:
				// 通道已满，跳过本次推送
			}
		}
	}
}

// checkAlerts 检查关键指标是否超过预设阈值，超过则输出告警日志
func checkAlerts(metrics Metrics) {
	if metrics.CPU > 80 {
		log.Printf("ALERT: High CPU usage: %.2f%%", metrics.CPU)
	}
	if metrics.Memory.UsedPercent > 90 {
		log.Printf("ALERT: High memory usage: %.2f%%", metrics.Memory.UsedPercent)
	}
	if metrics.Disk.UsedPercent > 95 {
		log.Printf("ALERT: Low disk space: %.2f%% used", metrics.Disk.UsedPercent)
	}
}