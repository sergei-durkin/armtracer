package main

import (
	"armtracer"
	"math/rand"
)

func main() {
	armtracer.Begin()
	defer armtracer.End()

	SuperFunction()
	a()
	recursive(10)
	a()
}

func recursive(n int) {
	defer armtracer.EndTrace(armtracer.BeginTrace(""))

	if n <= 0 {
		return
	}

	recursive(n - 1)
}

func a() {
	defer armtracer.EndTrace(armtracer.BeginTrace("a"))

	for i := 0; i < 10000; i++ {
		_ = i * i
	}

	b()
}

func b() {
	defer armtracer.EndTrace(armtracer.BeginTrace("b"))

	for i := 0; i < 100000000; i++ {
		_ = rand.Int63n(1 + int64(i))
	}

	c()
	d()
}

func c() {
	defer armtracer.EndTrace(armtracer.BeginTrace("c"))

	for i := 0; i < 100000000; i++ {
		_ = i * i
	}

	d()
}

func d() {
	defer armtracer.EndTrace(armtracer.BeginTrace("d"))

	for i := 0; i < 100000; i++ {
		_ = i * i
	}

	SuperFunction()
}

func SuperFunction() {
	defer armtracer.EndTrace(armtracer.BeginTrace("SuperFunction"))

	for i := 0; i < 100000000; i++ {
		_ = rand.Int63n(1 + int64(i))
	}

	GigaFunction()

	GigaFunction()
}

func GigaFunction() {
	defer armtracer.EndTrace(armtracer.BeginTrace("GigaFunction"))

	for i := 0; i < 1000000000; i++ {
		_ = i * i
	}
}
