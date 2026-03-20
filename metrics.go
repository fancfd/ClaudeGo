package main

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// collectMetrics 采集一次完整的系统指标快照并返回。
// CPU 采样会阻塞约 1 秒（由 gopsutil 内部实现）。
func collectMetrics() (Metrics, error) {
	now := time.Now()

	// CPU 使用率：false 表示返回所有核的平均值（单个元素切片）
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return Metrics{}, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// 虚拟内存信息（物理 RAM）
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return Metrics{}, fmt.Errorf("failed to get memory info: %w", err)
	}

	// 根文件系统磁盘使用情况
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return Metrics{}, fmt.Errorf("failed to get disk info: %w", err)
	}

	// 网络 I/O 统计：false 表示聚合所有接口，返回单条汇总记录
	netInfo, err := net.IOCounters(false)
	if err != nil || len(netInfo) == 0 {
		return Metrics{}, fmt.Errorf("failed to get network info: %w", err)
	}
	totalNet := netInfo[0]

	// 进程列表：采集失败时降级为空列表，不中断整体采集
	processes, err := getTopProcesses(10)
	if err != nil {
		log.Printf("Warning: failed to get process info: %v", err)
		processes = []ProcessMetrics{}
	}

	return Metrics{
		Timestamp: now,
		CPU:       cpuPercent[0],
		Memory: MemoryMetrics{
			Total:       memInfo.Total,
			Used:        memInfo.Used,
			UsedPercent: memInfo.UsedPercent,
		},
		Disk: DiskMetrics{
			Total:       diskInfo.Total,
			Free:        diskInfo.Free,
			UsedPercent: diskInfo.UsedPercent,
		},
		Network: NetworkMetrics{
			BytesSent:   totalNet.BytesSent,
			BytesRecv:   totalNet.BytesRecv,
			PacketsSent: totalNet.PacketsSent,
			PacketsRecv: totalNet.PacketsRecv,
		},
		Processes: processes,
	}, nil
}

// getTopProcesses 返回 CPU 占用率最高的前 limit 个进程。
// 遍历所有进程时，读取任意字段失败的进程会被静默跳过。
func getTopProcesses(limit int) ([]ProcessMetrics, error) {
	allProcesses, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var procMetrics []ProcessMetrics
	for _, p := range allProcesses {
		name, err := p.Name()
		if err != nil {
			continue // 进程可能已退出，跳过
		}
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			continue
		}
		memPercent, err := p.MemoryPercent()
		if err != nil {
			continue
		}
		procMetrics = append(procMetrics, ProcessMetrics{
			PID:           p.Pid,
			Name:          name,
			CPUPercent:    cpuPercent,
			MemoryPercent: memPercent,
		})
	}

	// 按 CPU 使用率降序排列，取前 limit 条
	sort.Slice(procMetrics, func(i, j int) bool {
		return procMetrics[i].CPUPercent > procMetrics[j].CPUPercent
	})

	if len(procMetrics) > limit {
		procMetrics = procMetrics[:limit]
	}

	return procMetrics, nil
}