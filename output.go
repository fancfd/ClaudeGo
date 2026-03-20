package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func outputToConsole(metrics Metrics) {
	fmt.Printf("Timestamp: %s\n", metrics.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("CPU: %.2f%%\n", metrics.CPU)
	fmt.Printf("Memory: %.2f%% used (%d MB / %d MB)\n",
		metrics.Memory.UsedPercent,
		metrics.Memory.Used/1024/1024,
		metrics.Memory.Total/1024/1024)
	fmt.Printf("Disk: %.2f%% used (%d GB free / %d GB total)\n",
		metrics.Disk.UsedPercent,
		metrics.Disk.Free/1024/1024/1024,
		metrics.Disk.Total/1024/1024/1024)
	fmt.Printf("Network: %d MB sent, %d MB recv\n",
		metrics.Network.BytesSent/1024/1024,
		metrics.Network.BytesRecv/1024/1024)
	fmt.Printf("Top Processes:\n")
	for i, p := range metrics.Processes {
		if i >= 5 { // Limit to top 5 for console
			break
		}
		fmt.Printf("  %s (PID %d): CPU %.2f%%, Mem %.2f%%\n",
			p.Name, p.PID, p.CPUPercent, p.MemoryPercent)
	}
	fmt.Println("---")
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
