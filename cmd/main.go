package main

import (
	"math/rand"
	"tracer"
)

func main() {
	tracer.Begin()
	defer tracer.End()

	a()
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
}
