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
$ go run -tags=armtracer ./cmd/main.go

CPU Frequency: 24.00 MHz

CPU stats:
| hits || flat ms  || flat %   || flat ns/op   || cum[ms]  || cum[%]   || cum[ns/op]    || function    |
========================================================================================================
|    1 || 0.0039ms ||  6.3179% || 3875.00ns/op || 0.0185ms || 30.2310% || 18541.67ns/op || main.run    |
|  100 || 0.0147ms || 23.9130% ||  146.67ns/op || 0.0147ms || 23.9130% ||   146.67ns/op || someWork    |

Memory stats:
| heap obj || alloc bytes || mallocs  || frees    || num gc   |
===============================================================
|        5 ||         176 ||        6 ||        1 ||        0 |

Total: time 0.06ms, cycles 1472, accounted 0.02ms (30.23%)
```

## Comparison with builtin timers
```bash
$ go run -tags=armtracer ./cmd/main.go

CPU Frequency: 24.00 MHz

CPU stats:
| hits     || flat ms    || flat %   || flat ns/op        || cum[ms]    || cum[%]   || cum[ns/op]        || function                |
=====================================================================================================================================
|        1 || 558.8115ms || 64.5928% || 558811458.33ns/op || 558.8115ms || 64.5928% || 558811458.33ns/op || timerCalls              |
|        1 || 224.8632ms || 25.9919% || 224863250.00ns/op || 306.2867ms || 35.4036% || 306286666.67ns/op || tracerCalls             |
|        1 ||   0.0008ms ||  0.0001% ||       750.00ns/op || 865.0989ms || 99.9965% || 865098875.00ns/op || e                       |
| 10000000 ||  81.4234ms ||  9.4117% ||         8.14ns/op ||  81.4234ms ||  9.4117% ||         8.14ns/op || main.millionTracerCalls |

Memory stats:
| heap obj || alloc bytes || mallocs  || frees    || num gc   |
===============================================================
|        0 ||           0 ||        0 ||        0 ||        0 |

Total: time 865.13ms, cycles 20763101, accounted 865.10ms (100.00%)
```

## Notes
- Flat time — time spent *only* in the traced function excluding child function calls.
- Cum (Cumulative) time — total time spent in the function including all nested function calls.
