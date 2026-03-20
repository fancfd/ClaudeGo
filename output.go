package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed web/dashboard.html
var dashboardHTML []byte

// bar 生成一个文本进度条，percent 为 0~100 的百分比，width 为总宽度（字符数）
// 示例输出：[################            ]
func bar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + fmt.Sprintf("%-*s", width, fmt.Sprintf("%s", repeat('#', filled))) + "]"
}

// truncate 将字符串截断到 n 个字符，超出部分替换为 "…"
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

// repeat 返回由 n 个字符 ch 组成的字符串
func repeat(ch rune, n int) string {
	s := make([]rune, n)
	for i := range s {
		s[i] = ch
	}
	return string(s)
}

// outputToConsole 以类 top 的方式将指标实时刷新到终端。
// 每次调用都会清屏并从头重新渲染，模拟原地更新效果。
func outputToConsole(metrics Metrics) {
	// ANSI 转义序列：清屏并将光标移到左上角
	fmt.Print("\033[2J\033[H")

	fmt.Printf("System Monitor — %s\n", metrics.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// 资源使用进度条（宽度 40 字符）+ 数值
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

	// 网络 I/O 累计值（自系统启动以来）
	fmt.Printf("Network: ↑ %d MB sent   ↓ %d MB recv\n",
		metrics.Network.BytesSent/1024/1024,
		metrics.Network.BytesRecv/1024/1024)
	fmt.Println()

	// 进程列表表头
	fmt.Printf("  %-6s  %-32s  %8s  %8s\n", "PID", "NAME", "CPU%", "MEM%")
	fmt.Println("  " + repeat('-', 60))
	for i, p := range metrics.Processes {
		if i >= 10 {
			break
		}
		fmt.Printf("  %-6d  %-32s  %7.2f%%  %7.2f%%\n",
			p.PID, truncate(p.Name, 32), p.CPUPercent, p.MemoryPercent)
	}
}

// outputToFile 将指标以 JSON Lines 格式追加写入文件（每行一个 JSON 对象）。
// 文件不存在时自动创建。
func outputToFile(metrics Metrics, filePath string) error {
	// O_APPEND 保证多次写入不覆盖历史记录
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

// startWebServer 启动 Gin HTTP 服务器，通过 GET /metrics 以 JSON 格式暴露最新指标。
// GET / 提供实时仪表盘页面（来自 web/dashboard.html，编译时通过 //go:embed 打包）。
// metricsChan 由主循环写入；此函数在独立 goroutine 中运行，阻塞直到服务器退出。
func startWebServer(port string, metricsChan <-chan Metrics) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	var latestMetrics Metrics

	// 在后台 goroutine 中持续接收最新指标，供 HTTP 处理器读取
	go func() {
		for metrics := range metricsChan {
			latestMetrics = metrics
		}
	}()

	// GET / 返回实时仪表盘 HTML 页面
	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", dashboardHTML)
	})

	// GET /metrics 返回最近一次采集的完整指标 JSON（供 API 调用或调试）
	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, latestMetrics)
	})

	log.Printf("Web server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
