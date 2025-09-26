# armtracer

Lightweight function-level tracer for ARM architecture, written in Go.

> [!WARNING]
> Not thread-safe.

## Installation

Add `armtracer` to your project:
```bash
go get github.com/sergei-durkin/armtracer
```

## Usage

Basic example:
```go
func main() {
	armtracer.Begin()
	defer armtracer.End()

	run()
}

func run() {
	defer armtracer.EndTrace(armtracer.BeginTrace("")) // trace name is optional

	for i := 0; i < 100; i++ {
		someWork()
	}
}

func someWork() {
	defer armtracer.EndTrace(armtracer.BeginTrace("someWork"))

	_ = rand.Intn(10)
}
```

Example output:
```bash
$ go run ./cmd/main.go

CPU Frequency: 24.00 MHz
    Hits    Avg (ms)   Flat (ms)    Cum (ms)    Flat% Function
       1        0.02        0.02        0.03   13.25% main.run
     100        0.00        0.02        0.02   12.47% someWork
Total: time 0.13ms, cycles 3103, accounted 0.03ms (25.72%)
```

## Comparison with builtin timers
```bash
$ go run ./cmd/main.go

CPU Frequency: 24.00 MHz
    Hits    Avg (ms)   Flat (ms)    Cum (ms)    Flat% Function
       1      607.40      607.40      607.40   57.88% timerCalls
       1      360.26      360.26      441.92   34.33% tracerCalls
       1        0.01        0.01     1049.33    0.00% e
10000000        0.00       81.66       81.66    7.78% main.millionTracerCalls
Total: time 1049.33ms, cycles 25183961, accounted 1049.33ms (100.00%)
```

## Notes
- Flat time — time spent *only* in the traced function excluding child function calls.
- Cum (Cumulative) time — total time spent in the function including all nested function calls.
