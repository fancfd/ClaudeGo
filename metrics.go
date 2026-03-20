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

func collectMetrics() (Metrics, error) {
	now := time.Now()

	// CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return Metrics{}, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Memory
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return Metrics{}, fmt.Errorf("failed to get memory info: %w", err)
	}

	// Disk (root filesystem)
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return Metrics{}, fmt.Errorf("failed to get disk info: %w", err)
	}

	// Network (all interfaces)
	netInfo, err := net.IOCounters(false)
	if err != nil || len(netInfo) == 0 {
		return Metrics{}, fmt.Errorf("failed to get network info: %w", err)
	}
	totalNet := netInfo[0]

	// Processes (top 10 by CPU)
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

func getTopProcesses(limit int) ([]ProcessMetrics, error) {
	allProcesses, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var procMetrics []ProcessMetrics
	for _, p := range allProcesses {
		name, err := p.Name()
		if err != nil {
			continue
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

	// Sort by CPU percent descending
	sort.Slice(procMetrics, func(i, j int) bool {
		return procMetrics[i].CPUPercent > procMetrics[j].CPUPercent
	})

	if len(procMetrics) > limit {
		procMetrics = procMetrics[:limit]
	}

	return procMetrics, nil
}