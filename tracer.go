package tracer

import (
	"fmt"
	"os"
	"runtime"
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

	traces map[int32][]trace
}

var (
	Profiler profiler
)

//go:nosplit
func BeginTrace(name string) trace {
	ptr := getCallerPC(unsafe.Pointer(&name))

	return begin(ptr, name)
}

//go:nosplit
func EndTrace(t trace) {
	t.end = getCnt()
	Profiler.traces[t.id] = append(Profiler.traces[t.id], t)
	Profiler.current = t.parrent
}

func Begin() {
	Profiler.traces = make(map[int32][]trace, size)
	Profiler.start = getCnt()
	Profiler.current = -1
}

func End() {
	Profiler.end = getCnt()
	total := Profiler.end - Profiler.start

	gh := make(map[int32][]trace)
	for _, ts := range Profiler.traces {
		for _, t := range ts {
			if t.start == 0 {
				continue
			}

			gh[t.parrent] = append(gh[t.parrent], t)
		}
	}

	var printTraces func(t trace, prefix string, isLast bool)
	printTraces = func(t trace, prefix string, isLast bool) {
		conn := "└── "
		if !isLast {
			conn = "├── "
		}

		cum := t.end - t.start
		flat := cum
		if children, ok := gh[t.id]; ok {
			for _, ct := range children {
				flat -= ct.end - ct.start
			}
		}

		cumPercent := float64(cum) * 100 / float64(total)
		flatPercent := float64(flat) * 100 / float64(total)

		fmt.Fprintf(os.Stderr, "%s%sTrace [%s]: flat: %d %.2f%% cum: %d %.2f%%\n", prefix, conn, t.name, flat, flatPercent, cum, cumPercent)

		for i, c := range gh[t.id] {
			next := prefix
			if isLast {
				next += "    "
			} else {
				next += "│   "
			}

			printTraces(c, next, i == len(gh[t.id])-1)
		}
	}

	for i, ts := range gh[-1] {
		printTraces(ts, "", i == len(gh[-1])-1)
	}

	fmt.Fprintf(os.Stderr, "Total: %d\n", Profiler.end-Profiler.start)
}

func begin(pc uintptr, name string) trace {
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
	t.id = int32(pc)
	t.name = name
	t.parrent = Profiler.current
	Profiler.current = t.id

	t.start = getCnt()

	return t
}

func add(ptr unsafe.Pointer, x int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) + uintptr(x))
}

//go:nosplit
func getCallerPC(ptr unsafe.Pointer) uintptr {
	return *(*uintptr)(add(ptr, -int(unsafe.Sizeof(ptr))))
}
