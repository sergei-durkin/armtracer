package armtracer

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"unsafe"
)

const (
	size = 1 << 14
)

func isb()
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
	isb()
	pc := caller(unsafe.Pointer(&name))

	return begin(pc, name)
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

	traces := Profiler.traces
	sort.Slice(traces, func(i, j int) bool {
		if traces[i].hit == 0 {
			return false
		}
		if traces[j].hit == 0 {
			return true
		}

		flatAvgI := traces[i].flat / uint64(traces[i].hit)
		flatAvgJ := traces[j].flat / uint64(traces[j].hit)
		if flatAvgI != flatAvgJ {
			return flatAvgI > flatAvgJ
		}

		flatPercentI := 100 * float64(traces[i].flat) / float64(totalCycles)
		flatPercentJ := 100 * float64(traces[j].flat) / float64(totalCycles)
		if flatPercentI != flatPercentJ {
			return flatPercentI > flatPercentJ
		}

		if traces[i].flat != traces[j].flat {
			return traces[i].flat > traces[j].flat
		}

		return traces[i].hit > traces[j].hit
	})

	sumElapsed := uint64(0)
	for _, t := range traces {
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
func caller(ptr unsafe.Pointer) uintptr {
	return *(*uintptr)(add(ptr, -int(unsafe.Sizeof(ptr))))
}

func add(p unsafe.Pointer, x int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + uintptr(x))
}

//go:nosplit
func begin(pc uintptr, name string) trace {
	pc = caller(unsafe.Pointer(&pc))

	var t trace

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
