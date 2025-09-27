package main

import (
	"math/rand"
	"time"

	"github.com/sergei-durkin/armtracer"
)

func main() {
	armtracer.Begin()
	defer armtracer.End()

	SuperFunction()
	a()
	recursive(10)
	a()

	compareWithTimer(int64(1e7))
}

func compareWithTimer(cycles int64) {
	defer armtracer.EndTrace(armtracer.BeginTrace("e"))

	t1 := armtracer.BeginTrace("tracerCalls")
	for i := int64(0); i < cycles; i++ {
		millionTracerCalls()
	}
	armtracer.EndTrace(t1)

	t := armtracer.BeginTrace("timerCalls")
	for i := int64(0); i < cycles; i++ {
		millionTimerCalls()
	}
	armtracer.EndTrace(t)
}

func millionTracerCalls() {
	defer armtracer.EndTrace(armtracer.BeginTrace(""))

	_ = 1 + 1
}

var elapsed int64

func millionTimerCalls() {
	t := time.Now().UnixNano()
	defer func() {
		elapsed += time.Now().UnixNano() - t
	}()

	_ = 1 + 1
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

	slice := []*int{}
	for i := 0; i < int(rand.Int31n(500)); i++ {
		slice = append(slice, getIntPtr())
	}

	for i := 0; i < 1000; i++ {
		_ = i * i
		slice[i%len(slice)] = getIntPtr()
	}
}

func getIntPtr() *int {
	defer armtracer.EndTrace(armtracer.BeginTrace(""))
	var x int

	return &x
}
