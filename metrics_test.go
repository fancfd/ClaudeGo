package main

import (
	"testing"
	"time"
)

func TestCollectMetrics(t *testing.T) {
	m, err := collectMetrics()
	if err != nil {
		t.Fatalf("collectMetrics() error: %v", err)
	}

	if m.Timestamp.IsZero() {
		t.Error("Timestamp is zero")
	}
	if time.Since(m.Timestamp) > 5*time.Second {
		t.Error("Timestamp is too old")
	}

	if m.CPU < 0 || m.CPU > 100 {
		t.Errorf("CPU percent out of range: %v", m.CPU)
	}

	if m.Memory.Total == 0 {
		t.Error("Memory.Total is 0")
	}
	if m.Memory.Used > m.Memory.Total {
		t.Errorf("Memory.Used (%d) > Memory.Total (%d)", m.Memory.Used, m.Memory.Total)
	}
	if m.Memory.UsedPercent < 0 || m.Memory.UsedPercent > 100 {
		t.Errorf("Memory.UsedPercent out of range: %v", m.Memory.UsedPercent)
	}

	if m.Disk.Total == 0 {
		t.Error("Disk.Total is 0")
	}
	if m.Disk.Free > m.Disk.Total {
		t.Errorf("Disk.Free (%d) > Disk.Total (%d)", m.Disk.Free, m.Disk.Total)
	}
	if m.Disk.UsedPercent < 0 || m.Disk.UsedPercent > 100 {
		t.Errorf("Disk.UsedPercent out of range: %v", m.Disk.UsedPercent)
	}
}

func TestGetTopProcesses(t *testing.T) {
	procs, err := getTopProcesses(5)
	if err != nil {
		t.Fatalf("getTopProcesses() error: %v", err)
	}

	if len(procs) > 5 {
		t.Errorf("expected at most 5 processes, got %d", len(procs))
	}

	for i, p := range procs {
		if p.PID <= 0 {
			t.Errorf("process[%d] has invalid PID: %d", i, p.PID)
		}
		if p.Name == "" {
			t.Errorf("process[%d] has empty name", i)
		}
		if p.CPUPercent < 0 {
			t.Errorf("process[%d] has negative CPUPercent: %v", i, p.CPUPercent)
		}
		if p.MemoryPercent < 0 {
			t.Errorf("process[%d] has negative MemoryPercent: %v", i, p.MemoryPercent)
		}
	}
}

func TestGetTopProcessesSortOrder(t *testing.T) {
	procs, err := getTopProcesses(10)
	if err != nil {
		t.Fatalf("getTopProcesses() error: %v", err)
	}

	for i := 1; i < len(procs); i++ {
		if procs[i].CPUPercent > procs[i-1].CPUPercent {
			t.Errorf("processes not sorted by CPU desc: [%d]=%.2f > [%d]=%.2f",
				i, procs[i].CPUPercent, i-1, procs[i-1].CPUPercent)
		}
	}
}

func TestGetTopProcessesLimit(t *testing.T) {
	for _, limit := range []int{1, 3, 10} {
		procs, err := getTopProcesses(limit)
		if err != nil {
			t.Fatalf("getTopProcesses(%d) error: %v", limit, err)
		}
		if len(procs) > limit {
			t.Errorf("getTopProcesses(%d) returned %d processes", limit, len(procs))
		}
	}
}

func TestCheckAlerts(t *testing.T) {
	// Verify checkAlerts does not panic for boundary and extreme values
	cases := []Metrics{
		{CPU: 0},
		{CPU: 80},
		{CPU: 81},
		{CPU: 100},
		{Memory: MemoryMetrics{UsedPercent: 90}},
		{Memory: MemoryMetrics{UsedPercent: 91}},
		{Disk: DiskMetrics{UsedPercent: 95}},
		{Disk: DiskMetrics{UsedPercent: 96}},
	}
	for _, m := range cases {
		checkAlerts(m) // should not panic
	}
}
