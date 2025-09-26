package armtracer

import (
	"fmt"
	"os"
	"runtime"
)

const (
	size = 1 << 14
)

func getCntvct() uint64
func getCntfrq() uint64

type trace struct {
	pc   uintptr
	name string

	id      int32
	parrent int32

	start uint64
	end   uint64

	cum  uint64
	flat uint64

	hit int32
}

type profiler struct {
	current int32

	start uint64
	end   uint64

	traces []trace
}

var (
	CPUFreqHz uint64 = 0
	Profiler  profiler
)

func init() {
	CPUFreqHz = getCntfrq()

	fmt.Fprintf(os.Stderr, "CPU Frequency: %.2f MHz\n", float64(CPUFreqHz)/1e6)
}

//go:nosplit
func BeginTrace(name string) trace {
	return begin(name)
}

//go:nosplit
func EndTrace(t trace) {
	t.end = getCntvct()

	Profiler.traces[t.parrent].flat -= t.end - t.start

	Profiler.traces[t.id].id = t.id
	Profiler.traces[t.id].pc = t.pc
	Profiler.traces[t.id].name = t.name
	Profiler.traces[t.id].hit++

	Profiler.traces[t.id].flat += t.end - t.start
	Profiler.traces[t.id].cum += t.end - t.start

	Profiler.traces[t.id].start = t.start
	Profiler.traces[t.id].end = t.end

	Profiler.current = t.parrent
}

func Begin() {
	Profiler.traces = make([]trace, size)
	Profiler.start = getCntvct()
}

func End() {
	Profiler.end = getCntvct()
	totalCycles := Profiler.end - Profiler.start
	totalMsec := cyclesToMsec(totalCycles)

	fmt.Fprintf(os.Stderr, "%8s %11s %11s %11s %8s %s\n", "Hits", "Avg (ms)", "Flat (ms)", "Cum (ms)", "Flat%", "Function")

	sumElapsed := uint64(0)
	for _, t := range Profiler.traces {
		if t.hit == 0 {
			continue
		}

		sumElapsed += t.flat
		flatPercent := 100 * float64(t.flat) / float64(totalCycles)
		avgCycles := t.flat / uint64(t.hit)
		avgMsec := cyclesToMsec(avgCycles)

		fmt.Fprintf(os.Stderr, "%8d %11.2f %11.2f %11.2f %7.2f%% %s\n", t.hit, avgMsec, cyclesToMsec(t.flat), cyclesToMsec(t.cum), flatPercent, t.name)
	}

	fmt.Fprintf(os.Stderr, "Total: time %.2fms, cycles %d, accounted %.2fms (%.2f%%)\n", totalMsec, totalCycles, cyclesToMsec(sumElapsed), 100*float64(sumElapsed)/float64(totalCycles))
}

//go:nosplit
func begin(name string) trace {
	var t trace

	pc, _, _, _ := runtime.Caller(2)
	if name == "" {
		f := runtime.FuncForPC(pc)
		if f != nil {
			name = f.Name()
		} else {
			name = "unknown"
		}
	}

	t.pc = pc
	t.id = idx(pc)
	t.name = name
	t.parrent = Profiler.current
	Profiler.current = t.id

	t.start = getCntvct()

	return t
}

func idx(pc uintptr) int32 {
	res := int32(pc % size)
	if res == 0 {
		return 1
	}

	return res
}

func cyclesToMsec(c uint64) float64 {
	if CPUFreqHz == 0 {
		return 0
	}

	return 1000 * float64(c) / float64(CPUFreqHz)
}
