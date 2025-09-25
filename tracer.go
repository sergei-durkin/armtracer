package tracer

import (
	"fmt"
	"os"
	"sort"
	"unsafe"
)

const (
	size = 1 << 14
)

func getCnt() uint64

type trace struct {
	pc   uintptr
	name string

	id      int32
	parrent int32

	start uint64
	end   uint64
}

type profiler struct {
	current int32

	start uint64
	end   uint64

	traces []trace
}

var (
	Profiler profiler
)

//go:nosplit
func BeginTrace(name string) trace {
	return begin(getCallerPC(unsafe.Pointer(&name)), name)
}

//go:nosplit
func EndTrace(t trace) {
	Profiler.traces[t.id].end = getCnt()
}

func Begin() {
	Profiler.traces = make([]trace, size)
	Profiler.start = getCnt()
	Profiler.current = -1
}

func End() {
	Profiler.end = getCnt()
	total := Profiler.end - Profiler.start

	sort.Slice(Profiler.traces, func(i, j int) bool {
		return Profiler.traces[i].start < Profiler.traces[j].start
	})

	gh := make(map[int32]int64)
	for _, t := range Profiler.traces {
		if t.start == 0 {
			continue
		}

		gh[t.parrent] += int64(t.end - t.start)
	}

	for _, t := range Profiler.traces {
		if t.start == 0 {
			continue
		}

		cum := t.end - t.start
		flat := cum - uint64(gh[t.id])

		fmt.Fprintf(os.Stderr, "Trace %s:\n\tflat: %d %.2f%%\n", t.name, flat, float64(flat)*100/float64(total))

		if cum != flat {
			fmt.Fprintf(os.Stderr, "\tcum: %d %.2f%%\n", cum, float64(cum)*100/float64(total))
		}
	}

	fmt.Fprintf(os.Stderr, "Total: %d\n", Profiler.end-Profiler.start)
}

func begin(pc uintptr, name string) trace {
	var t trace

	t.pc = pc
	t.id = idx(pc)
	t.name = name
	t.parrent = Profiler.current
	Profiler.current = t.id

	t.start = getCnt()
	Profiler.traces[t.id] = t

	return t
}

func add(ptr unsafe.Pointer, x int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) + uintptr(x))
}

//go:nosplit
func getCallerPC(ptr unsafe.Pointer) uintptr {
	return *(*uintptr)(add(ptr, -int(unsafe.Sizeof(ptr))))
}

func idx(pc uintptr) int32 {
	res := int32(pc % size)
	if res == 0 {
		return 1
	}

	return res
}
