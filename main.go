package main

import (
	"flag"
	"log"
	"time"
)

type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskMetrics struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkMetrics struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

type ProcessMetrics struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
}

type Metrics struct {
	Timestamp time.Time       `json:"timestamp"`
	CPU       float64         `json:"cpu_percent"`
	Memory    MemoryMetrics   `json:"memory"`
	Disk      DiskMetrics     `json:"disk"`
	Network   NetworkMetrics  `json:"network"`
	Processes []ProcessMetrics `json:"processes"`
}

func main() {
	consoleFlag := flag.Bool("console", false, "Enable console output")
	fileFlag := flag.String("file", "", "File path for JSON output (appends each metric as a line)")
	webFlag := flag.Bool("web", false, "Enable web server for metrics")
	portFlag := flag.String("port", "8080", "Port for web server")
	flag.Parse()

	if !*consoleFlag && *fileFlag == "" && !*webFlag {
		log.Fatal("At least one output mode must be enabled: --console, --file, or --web")
	}

	metricsChan := make(chan Metrics, 1)

	if *webFlag {
		go startWebServer(*portFlag, metricsChan)
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics, err := collectMetrics()
		if err != nil {
			log.Printf("Error collecting metrics: %v", err)
			continue
		}

		// Check for alerts
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
			select {
			case metricsChan <- metrics:
			default:
				// Channel full, skip this update
			}
		}
	}
}

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