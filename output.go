package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func bar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + fmt.Sprintf("%-*s", width, fmt.Sprintf("%s", repeat('#', filled))) + "]"
}

func repeat(ch rune, n int) string {
	s := make([]rune, n)
	for i := range s {
		s[i] = ch
	}
	return string(s)
}

func outputToConsole(metrics Metrics) {
	// Clear screen and move cursor to top-left
	fmt.Print("\033[2J\033[H")

	fmt.Printf("System Monitor — %s\n", metrics.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Println()

	fmt.Printf("CPU:    %s %5.1f%%\n", bar(metrics.CPU, 40), metrics.CPU)
	fmt.Printf("Memory: %s %5.1f%%  (%d MB / %d MB)\n",
		bar(metrics.Memory.UsedPercent, 40),
		metrics.Memory.UsedPercent,
		metrics.Memory.Used/1024/1024,
		metrics.Memory.Total/1024/1024)
	fmt.Printf("Disk:   %s %5.1f%%  (%d GB free / %d GB total)\n",
		bar(metrics.Disk.UsedPercent, 40),
		metrics.Disk.UsedPercent,
		metrics.Disk.Free/1024/1024/1024,
		metrics.Disk.Total/1024/1024/1024)
	fmt.Println()

	fmt.Printf("Network: ↑ %d MB sent   ↓ %d MB recv\n",
		metrics.Network.BytesSent/1024/1024,
		metrics.Network.BytesRecv/1024/1024)
	fmt.Println()

	fmt.Printf("  %-6s  %-20s  %8s  %8s\n", "PID", "NAME", "CPU%", "MEM%")
	fmt.Println("  " + repeat('-', 48))
	for i, p := range metrics.Processes {
		if i >= 10 {
			break
		}
		fmt.Printf("  %-6d  %-20s  %7.2f%%  %7.2f%%\n",
			p.PID, p.Name, p.CPUPercent, p.MemoryPercent)
	}
}

func outputToFile(metrics Metrics, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(data) + "\n")
	return err
}

func startWebServer(port string, metricsChan <-chan Metrics) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	var latestMetrics Metrics

	// Goroutine to update latest metrics
	go func() {
		for metrics := range metricsChan {
			latestMetrics = metrics
		}
	}()

	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, latestMetrics)
	})

	log.Printf("Web server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
