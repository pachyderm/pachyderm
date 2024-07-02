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
	"github.com/pachyderm/pachyderm/v2/src/internal/meters"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/prometheus/procfs"
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

func getOneProcessStats(p procfs.Proc, stat procfs.ProcStat) (*ProcessStats, error) {
	var errs error
	var pstat ProcessStats

	if fds, err := p.FileDescriptors(); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "FileDescriptors"))
	} else {
		pstat.FDCount = len(fds)
	}

	if io, err := p.IO(); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "IO"))
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
		errors.JoinInto(&errs, err)
	} else {
		var err error
		pstat.OOMScore, err = strconv.Atoi(strings.TrimSpace(string(oomScoreRaw)))
		if err != nil {
			errors.JoinInto(&errs, errors.Wrap(err, "parse oom_score"))
		}
	}
	return &pstat, errs
}

// getProcessStats returns information about the provided process; 0 is "self", less than 0 is an
// entire process group.
func getProcessStats(fs procfs.FS, pid int) (*ProcessStats, error) {
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
		pstat, err := getOneProcessStats(p, stat)
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
				errors.JoinInto(&errs, errors.Wrapf(err, "Stat(%v)", p.PID))
				continue
			}
			if stat.PGRP != pgid {
				continue
			}
			pstat, err := getOneProcessStats(p, stat)
			if err != nil {
				errors.JoinInto(&errs, errors.Wrapf(err, "getProcessStats(%v)", p.PID))
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
		result.RBytes += s.RBytes
		result.WBytes += s.WBytes
		result.CanceledWBytes += s.CanceledWBytes
		result.RChars += s.RChars
		result.WChars += s.WChars
		result.ResidentMemory += s.ResidentMemory
		result.RSysc += s.RSysc
		result.WSysc += s.WSysc
		result.OOMScore = miscutil.Max(result.OOMScore, s.OOMScore)
	}
	return &result, nil
}

type SystemStats struct {
	CPUTime float64
	// TODO(jonathan): We should monitor disk usage of /tmp, /pfs, and /pach.
}

func getSystemStats(fs procfs.FS) (*SystemStats, error) {
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
		pctx.WithDelta("self_open_fd_count", 10, meters.Deferred()),
		pctx.WithDelta("self_rchar_bytes", int(10e6), meters.Deferred()),
		pctx.WithDelta("self_wchar_bytes", int(10e6), meters.Deferred()),
		pctx.WithDelta("self_bytes_written_bytes", int(10e6), meters.Deferred()),
		pctx.WithDelta("self_bytes_read_bytes", int(10e6), meters.Deferred()),
		pctx.WithDelta("self_canceled_write_bytes", int(10e6), meters.Deferred()),
		pctx.WithDelta("self_read_syscall_count", int(1e6), meters.Deferred()),
		pctx.WithDelta("self_write_syscall_count", int(1e6), meters.Deferred()),
		pctx.WithDelta("self_cpu_time_seconds", float64(runtime.GOMAXPROCS(-1)), meters.Deferred()),
		pctx.WithDelta("self_resident_memory_bytes", int(100e6), meters.Deferred()),
		pctx.WithGauge("self_oom_score", 0, meters.WithFlushInterval(30*time.Minute)),
		pctx.WithDelta("system_cpu_time_seconds", float64(60), meters.Deferred()),
	)
	for {
		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			log.Info(ctx, "MonitorSelf thread exiting", zap.Error(context.Cause(ctx)))
			return
		}
		fs, err := procfs.NewDefaultFS()
		if err != nil {
			log.Info(ctx, "problem opening procfs")
			continue
		}

		stats, err := getProcessStats(fs, 0)
		if err != nil {
			log.Info(ctx, "problem getting self stats", zap.Error(err))
			continue
		}
		meters.Set(ctx, "self_open_fd_count", stats.FDCount)
		meters.Set(ctx, "self_rchar_bytes", int(stats.RChars))
		meters.Set(ctx, "self_wchar_bytes", int(stats.WChars))
		meters.Set(ctx, "self_bytes_read_bytes", int(stats.RBytes))
		meters.Set(ctx, "self_bytes_written_bytes", int(stats.WBytes))
		meters.Set(ctx, "self_canceled_write_bytes", int(stats.CanceledWBytes))
		meters.Set(ctx, "self_read_syscall_count", int(stats.RSysc))
		meters.Set(ctx, "self_write_syscall_count", int(stats.WSysc))
		meters.Set(ctx, "self_cpu_time_seconds", stats.CPUTime)
		meters.Set(ctx, "self_resident_memory_bytes", stats.ResidentMemory)
		meters.Set(ctx, "self_oom_score", stats.OOMScore)

		sys, err := getSystemStats(fs)
		if err != nil {
			log.Info(ctx, "problem getting system stats", zap.Error(err))
			continue
		}
		meters.Set(ctx, "system_cpu_time_seconds", sys.CPUTime)
	}
}

// MonitorProcessGroup produces metrics about a process group until the context is canceled.
func MonitorProcessGroup(ctx context.Context, pid int) {
	if pid > 0 {
		pid = -pid
	}
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		log.Error(ctx, "unable to monitor process group", zap.Int("pid", pid), zap.Error(err))
		return
	}
	ctx = pctx.Child(ctx, "",
		pctx.WithGauge("open_fd_count", 0, meters.WithFlushInterval(30*time.Second)),
		pctx.WithDelta("rchar_bytes", int(1e6), meters.Deferred()),
		pctx.WithDelta("wchar_bytes", int(1e6), meters.Deferred()),
		pctx.WithDelta("bytes_written_bytes", int(1e6), meters.Deferred()),
		pctx.WithDelta("bytes_read_bytes", int(1e6), meters.Deferred()),
		pctx.WithDelta("canceled_write_bytes", int(1e6), meters.Deferred()),
		pctx.WithDelta("read_syscall_count", int(1e4), meters.Deferred()),
		pctx.WithDelta("write_syscall_count", int(1e4), meters.Deferred()),
		pctx.WithDelta("cpu_time_seconds", float64(30), meters.WithFlushInterval(30*time.Second)),
		pctx.WithDelta("resident_memory_bytes", int(100e6), meters.WithFlushInterval(30*time.Second)),
		pctx.WithGauge("oom_score", 0, meters.WithFlushInterval(30*time.Minute)),
	)
	for {
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			log.Debug(ctx, "MonitorUserCode exiting", zap.Int("pid", pid))
			return
		}
		stats, err := getProcessStats(fs, pid)
		if err != nil {
			log.Info(ctx, "problem getting child stats", zap.Error(err))
			continue
		}
		meters.Set(ctx, "open_fd_count", stats.FDCount)
		meters.Set(ctx, "rchar_bytes", int(stats.RChars))
		meters.Set(ctx, "wchar_bytes", int(stats.WChars))
		meters.Set(ctx, "bytes_read_bytes", int(stats.RBytes))
		meters.Set(ctx, "bytes_written_bytes", int(stats.WBytes))
		meters.Set(ctx, "canceled_write_bytes", int(stats.CanceledWBytes))
		meters.Set(ctx, "read_syscall_count", int(stats.RSysc))
		meters.Set(ctx, "write_syscall_count", int(stats.WSysc))
		meters.Set(ctx, "cpu_time_seconds", stats.CPUTime)
		meters.Set(ctx, "resident_memory_bytes", stats.ResidentMemory)
		meters.Set(ctx, "oom_score", stats.OOMScore)
	}
}
