// Package proc contains utilities for monitoring the resource use of processes.
package proc

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/m"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/prometheus/procfs"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type ProcessStats struct {
	FDCount        int
	RChars, WChars uint64 // read/write to any fd
	RSysc, WSysc   uint64
	RBytes, WBytes uint64 // read/write to disk; see proc(5)
	CanceledWBytes int64
	CPUTime        float64
	ResidentMemory int
	OOMScore       int
}

func getProcessStats(p procfs.Proc, stat procfs.ProcStat) (*ProcessStats, error) {
	var errs error
	var pstat ProcessStats

	if fds, err := p.FileDescriptors(); err != nil {
		multierr.AppendInto(&errs, errors.Wrap(err, "FileDescriptors"))
	} else {
		pstat.FDCount = len(fds)
	}

	if io, err := p.IO(); err != nil {
		multierr.AppendInto(&errs, errors.Wrap(err, "IO"))
	} else {
		pstat.RChars = io.RChar
		pstat.WChars = io.WChar
		pstat.RSysc = io.SyscR
		pstat.WSysc = io.SyscW
		pstat.RBytes = io.ReadBytes
		pstat.WBytes = io.WriteBytes
		pstat.CanceledWBytes = io.CancelledWriteBytes
	}
	pstat.CPUTime = stat.CPUTime()
	pstat.ResidentMemory = stat.ResidentMemory()

	if oomScoreRaw, err := os.ReadFile(fmt.Sprintf("/proc/%d/oom_score", stat.PID)); err != nil {
		multierr.AppendInto(&errs, err)
	} else {
		var err error
		pstat.OOMScore, err = strconv.Atoi(strings.TrimSpace(string(oomScoreRaw)))
		if err != nil {
			multierr.AppendInto(&errs, errors.Wrap(err, "parse oom_score"))
		}
	}
	return &pstat, errs
}

// GetProcessStats returns information about the provided process; 0 is "self", less than 0 is a
// process group.
func GetProcessStats(pid int) (*ProcessStats, error) {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return nil, errors.Wrap(err, "init procfs")
	}

	var stats []*ProcessStats
	if pid >= 0 { // one process, maybe self
		var p procfs.Proc
		var err error
		if pid == 0 {
			p, err = fs.Self()
		} else {
			p, err = fs.Proc(pid)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "Proc(%v)", pid)
		}
		stat, err := p.Stat()
		if err != nil {
			return nil, errors.Wrapf(err, "Stat(%v)", p.PID)
		}
		pstat, err := getProcessStats(p, stat)
		if err != nil {
			return nil, errors.Wrapf(err, "getProcessStats(%v)", p.PID)
		}
		stats = append(stats, pstat)
	} else { // process group
		pgid := -pid
		all, err := fs.AllProcs()
		if err != nil {
			return nil, errors.Wrap(err, "AllProcs")
		}
		var errs error
		for _, p := range all {
			stat, err := p.Stat()
			if err != nil {
				multierr.AppendInto(&errs, errors.Wrapf(err, "Stat(%v)", p.PID))
				continue
			}
			if stat.PGRP != pgid {
				continue
			}
			pstat, err := getProcessStats(p, stat)
			if err != nil {
				multierr.AppendInto(&errs, errors.Wrapf(err, "getProcessStats(%v)", p.PID))
				continue
			}
			stats = append(stats, pstat)
		}
		if errs != nil {
			return nil, errs
		}
	}
	var result ProcessStats
	for _, s := range stats {
		result.CPUTime += s.CPUTime
		result.CanceledWBytes += s.CanceledWBytes
		result.FDCount += s.FDCount
		result.OOMScore += s.OOMScore
		result.RBytes += s.RBytes
		result.WBytes += s.WBytes
		result.CanceledWBytes += s.CanceledWBytes
		result.RChars += s.RChars
		result.WChars += s.WChars
		result.ResidentMemory += s.ResidentMemory
		result.RSysc += s.RSysc
		result.WSysc += s.WSysc
	}
	return &result, nil
}

type SystemStats struct {
	CPUTime float64
}

func GetSystemStats() (*SystemStats, error) {
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return nil, errors.Wrap(err, "init procfs")
	}
	sys, err := fs.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "Stat")
	}
	return &SystemStats{
		CPUTime: sys.CPUTotal.User + sys.CPUTotal.System + sys.CPUTotal.Steal,
	}, nil
}

func MonitorSelf(ctx context.Context) {
	ctx = pctx.Child(ctx, "",
		pctx.WithDelta("open_fd_count", 10, m.Deferred()),
		pctx.WithDelta("rchar_bytes", int(1e6), m.Deferred()),
		pctx.WithDelta("wchar_bytes", int(1e6), m.Deferred()),
		pctx.WithDelta("bytes_written_bytes", int(1e6), m.Deferred()),
		pctx.WithDelta("bytes_read_bytes", int(1e6), m.Deferred()),
		pctx.WithDelta("canceled_write_bytes", int(1e6), m.Deferred()),
		pctx.WithDelta("read_syscall_count", int(1e5), m.Deferred()),
		pctx.WithDelta("write_syscall_count", int(1e5), m.Deferred()),
		pctx.WithDelta("cpu_time_seconds", float64(runtime.GOMAXPROCS(-1)), m.Deferred()),
		pctx.WithDelta("resident_memory_bytes", int(100e6), m.Deferred()),
		pctx.WithGauge("oom_score", 0, m.WithFlushInterval(30*time.Minute)),
		pctx.WithDelta("system_cpu_time_seconds", float64(60), m.Deferred()),
	)
	for {
		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			log.Info(ctx, "MonitorSelf thread exiting", zap.Error(context.Cause(ctx)))
			return
		}
		stats, err := GetProcessStats(0)
		if err != nil {
			log.Info(ctx, "problem getting self stats", zap.Error(err))
			continue
		}
		m.Set(ctx, "open_fd_count", stats.FDCount)
		m.Set(ctx, "rchar_bytes", int(stats.RChars))
		m.Set(ctx, "wchar_bytes", int(stats.WChars))
		m.Set(ctx, "bytes_read_bytes", int(stats.RBytes))
		m.Set(ctx, "bytes_written_bytes", int(stats.WBytes))
		m.Set(ctx, "canceled_write_bytes", int(stats.CanceledWBytes))
		m.Set(ctx, "read_syscall_count", int(stats.RSysc))
		m.Set(ctx, "write_syscall_count", int(stats.WSysc))
		m.Set(ctx, "cpu_time_seconds", stats.CPUTime)
		m.Set(ctx, "resident_memory_bytes", stats.ResidentMemory)
		m.Set(ctx, "oom_score", stats.OOMScore)

		sys, err := GetSystemStats()
		if err != nil {
			log.Info(ctx, "problem getting system stats", zap.Error(err))
			continue
		}
		m.Set(ctx, "system_cpu_time_seconds", sys.CPUTime)
	}
}
