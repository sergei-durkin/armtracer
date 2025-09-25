package main

import (
	"math/rand"
	"tracer"
)

func main() {
	tracer.Begin()
	defer tracer.End()

	SuperFunction()
	a()
	recursive(10)
	a()
}

func recursive(n int) {
	defer tracer.EndTrace(tracer.BeginTrace(""))

	if n <= 0 {
		return
	}

	recursive(n - 1)
}

func a() {
	defer tracer.EndTrace(tracer.BeginTrace("a"))

	for i := 0; i < 10000; i++ {
		_ = i * i
	}

	b()
}

func b() {
	defer tracer.EndTrace(tracer.BeginTrace("b"))

	for i := 0; i < 100000000; i++ {
		_ = rand.Int63n(1 + int64(i))
	}

	c()
	d()
}

func c() {
	defer tracer.EndTrace(tracer.BeginTrace("c"))

	for i := 0; i < 100000000; i++ {
		_ = i * i
	}

	d()
}

func d() {
	defer tracer.EndTrace(tracer.BeginTrace("d"))

	for i := 0; i < 100000; i++ {
		_ = i * i
	}

	SuperFunction()
}

func SuperFunction() {
	defer tracer.EndTrace(tracer.BeginTrace("SuperFunction"))

	for i := 0; i < 100000000; i++ {
		_ = rand.Int63n(1 + int64(i))
	}

	GigaFunction()

	GigaFunction()
}

func GigaFunction() {
	defer tracer.EndTrace(tracer.BeginTrace("GigaFunction"))

	for i := 0; i < 1000000000; i++ {
		_ = i * i
	}
}
