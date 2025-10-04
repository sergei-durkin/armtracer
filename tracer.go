//go:build armtracer
// +build armtracer

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

	traces        []trace
	startMemStats runtime.MemStats
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
	runtime.ReadMemStats(&Profiler.startMemStats)

	Profiler.start = getCntvct()
}

type resultRow struct {
	name string

	hits string

	flatMs      string
	flatPercent string
	flatNsPerOp string

	cumMs      string
	cumPercent string
	cumNsPerOp string
}

func End() {
	Profiler.end = getCntvct()

	var endMemStats runtime.MemStats
	runtime.ReadMemStats(&endMemStats)

	allocBytes := endMemStats.Alloc - Profiler.startMemStats.Alloc
	mallocs := endMemStats.Mallocs - Profiler.startMemStats.Mallocs
	frees := endMemStats.Frees - Profiler.startMemStats.Frees
	numGC := endMemStats.NumGC - Profiler.startMemStats.NumGC
	heapObj := endMemStats.HeapObjects - Profiler.startMemStats.HeapObjects

	totalCycles := Profiler.end - Profiler.start
	totalMsec := cyclesToMsec(totalCycles)

	traces := Profiler.traces
	sort.Slice(traces, func(i, j int) bool {
		if traces[i].hit == 0 {
			return false
		}
		if traces[j].hit == 0 {
			return true
		}

		nsPerOpI := cyclesToNsec(traces[i].flat) / float64(traces[i].hit)
		nsPerOpJ := cyclesToNsec(traces[j].flat) / float64(traces[j].hit)
		if nsPerOpI != nsPerOpJ {
			return nsPerOpI > nsPerOpJ
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

	hitsLn := 4

	flatMsLn := 8
	flatPercentLn := 8
	flatNsPerOpLn := 11

	cumMsLn := 8
	cumPercentLn := 8
	cumNsPerOpLn := 11

	nameLn := 11

	rows := make([]resultRow, 0, len(traces))
	sumElapsed := uint64(0)
	for _, t := range traces {
		if t.hit == 0 {
			continue
		}

		var r resultRow

		name := t.name
		if name == "" {
			f := runtime.FuncForPC(t.pc)
			if f != nil {
				name = f.Name()
			} else {
				name = "unknown"
			}
		}

		sumElapsed += t.flat
		flatPercent := 100 * float64(t.flat) / float64(totalCycles)
		flatNsPerOp := cyclesToNsec(t.flat) / float64(t.hit)

		cumPercent := 100 * float64(t.cum) / float64(totalCycles)
		cumNsPerOp := cyclesToNsec(t.cum) / float64(t.hit)

		r.name = fmt.Sprintf("%s", name)
		nameLn = max(nameLn, len(r.name))

		r.hits = fmt.Sprintf("%d", t.hit)
		hitsLn = max(hitsLn, len(r.hits))

		r.flatMs = fmt.Sprintf("%.4fms", cyclesToMsec(t.flat))
		flatMsLn = max(flatMsLn, len(r.flatMs))

		r.flatPercent = fmt.Sprintf("%.4f%%", flatPercent)
		flatPercentLn = max(flatPercentLn, len(r.flatPercent))

		r.flatNsPerOp = fmt.Sprintf("%.2fns/op", flatNsPerOp)
		flatNsPerOpLn = max(flatNsPerOpLn, len(r.flatNsPerOp))

		r.cumMs = fmt.Sprintf("%.4fms", cyclesToMsec(t.cum))
		cumMsLn = max(cumMsLn, len(r.cumMs))

		r.cumPercent = fmt.Sprintf("%.4f%%", cumPercent)
		cumPercentLn = max(cumPercentLn, len(r.cumPercent))

		r.cumNsPerOp = fmt.Sprintf("%.2fns/op", cumNsPerOp)
		cumNsPerOpLn = max(cumNsPerOpLn, len(r.cumNsPerOp))

		rows = append(rows, r)
	}

	fmt.Fprintf(os.Stderr, "\nCPU stats:\n")
	{

		fmt.Fprintf(os.Stderr, "| %-[1]*s |", hitsLn, "hits")

		fmt.Fprintf(os.Stderr, "| %-[1]*s |", flatMsLn, "flat ms")
		fmt.Fprintf(os.Stderr, "| %-[1]*s |", flatPercentLn, "flat %")
		fmt.Fprintf(os.Stderr, "| %-[1]*s |", flatNsPerOpLn, "flat ns/op")

		fmt.Fprintf(os.Stderr, "| %-[1]*s |", cumMsLn, "cum[ms]")
		fmt.Fprintf(os.Stderr, "| %-[1]*s |", cumPercentLn, "cum[%]")
		fmt.Fprintf(os.Stderr, "| %-[1]*s |", cumNsPerOpLn, "cum[ns/op]")

		fmt.Fprintf(os.Stderr, "| %-[1]*s |\n", nameLn, "function")

		{ // print divider
			sum := hitsLn + flatMsLn + flatPercentLn + flatNsPerOpLn + cumMsLn + cumPercentLn + cumNsPerOpLn + nameLn + 32
			hr := make([]byte, sum+1)
			for i := 0; i < sum; i++ {
				hr[i] = '='
			}
			hr[sum] = '\n'

			fmt.Fprint(os.Stderr, string(hr))
		}

		for i := 0; i < len(rows); i++ {
			r := rows[i]

			fmt.Fprintf(os.Stderr, "| %[1]*s |", hitsLn, r.hits)

			fmt.Fprintf(os.Stderr, "| %[1]*s |", flatMsLn, r.flatMs)
			fmt.Fprintf(os.Stderr, "| %[1]*s |", flatPercentLn, r.flatPercent)
			fmt.Fprintf(os.Stderr, "| %[1]*s |", flatNsPerOpLn, r.flatNsPerOp)

			fmt.Fprintf(os.Stderr, "| %[1]*s |", cumMsLn, r.cumMs)
			fmt.Fprintf(os.Stderr, "| %[1]*s |", cumPercentLn, r.cumPercent)
			fmt.Fprintf(os.Stderr, "| %[1]*s |", cumNsPerOpLn, r.cumNsPerOp)

			fmt.Fprintf(os.Stderr, "| %-[1]*s |", nameLn, r.name)

			fmt.Fprint(os.Stderr, "\n")
		}
	}

	fmt.Fprintf(os.Stderr, "\nMemory stats:\n")
	{
		heapObjStr := fmt.Sprintf("%d", heapObj)
		heapObjLn := max(8, len(heapObjStr))

		allocBytesStr := fmt.Sprintf("%d", allocBytes)
		allocBytesLn := max(11, len(allocBytesStr))

		mallocsStr := fmt.Sprintf("%d", mallocs)
		mallocsLn := max(8, len(mallocsStr))

		freesStr := fmt.Sprintf("%d", frees)
		freesLn := max(8, len(freesStr))

		numGCStr := fmt.Sprintf("%d", numGC)
		numGCLn := max(8, len(numGCStr))

		{ // memory stats header
			fmt.Fprintf(os.Stderr, "| %-[1]*s |", heapObjLn, "heap obj")

			fmt.Fprintf(os.Stderr, "| %-[1]*s |", allocBytesLn, "alloc bytes")
			fmt.Fprintf(os.Stderr, "| %-[1]*s |", mallocsLn, "mallocs")
			fmt.Fprintf(os.Stderr, "| %-[1]*s |", freesLn, "frees")

			fmt.Fprintf(os.Stderr, "| %-[1]*s |\n", numGCLn, "num gc")
		}

		{ // print divider
			sum := heapObjLn + allocBytesLn + mallocsLn + freesLn + numGCLn + 20
			hr := make([]byte, sum+1)
			for i := 0; i < sum; i++ {
				hr[i] = '='
			}
			hr[sum] = '\n'
			fmt.Fprint(os.Stderr, string(hr))
		}

		{ // memory stats row
			fmt.Fprintf(os.Stderr, "| %[1]*d |", heapObjLn, heapObj)

			fmt.Fprintf(os.Stderr, "| %[1]*d |", allocBytesLn, allocBytes)
			fmt.Fprintf(os.Stderr, "| %[1]*d |", mallocsLn, mallocs)
			fmt.Fprintf(os.Stderr, "| %[1]*d |", freesLn, frees)

			fmt.Fprintf(os.Stderr, "| %[1]*d |\n", numGCLn, numGC)
		}
	}

	fmt.Fprintf(os.Stderr, "\nTotal: time %.2fms, cycles %d, accounted %.2fms (%.2f%%)\n", totalMsec, totalCycles, cyclesToMsec(sumElapsed), 100*float64(sumElapsed)/float64(totalCycles))
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

func cyclesToNsec(c uint64) float64 {
	if CPUFreqHz == 0 {
		return 0
	}

	return 1000000000 * float64(c) / float64(CPUFreqHz)
}
