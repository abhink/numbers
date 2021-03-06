// This file simply contains various way to collect numbers over a channel and sort them.
// For sufficiently large count of numbers, there is not much performance difference
// in mapThenAppend and mapAndAppend, so NumbersGetter uses mapThenAppend.
// mapNoAppend has best performance but requires total count of numbers to be
// known in advance.

// Bechmark can be run using: `go test numbers -bench=. -run=Bench`
package numbers

import (
	"math/rand"
	"sort"
	"testing"
)

var benchResult []int

func getNumbers() []int {
	return rand.Perm(10000)
}

func mapThenAppend(l []int) []int {
	numbersMap := make(map[int]bool)
	for _, n := range l {
		numbersMap[n] = true
	}

	response := []int{}
	for k, _ := range numbersMap {
		response = append(response, k)
	}

	sort.Ints(response)
	return response
}

func mapAndAppend(l []int) []int {
	numbersMap := make(map[int]bool)
	response := []int{}

	for _, n := range l {
		if !numbersMap[n] {
			response = append(response, n)
		}
		numbersMap[n] = true
	}

	sort.Ints(response)
	return response
}

func mapNoAppend(l []int) []int {
	numbersMap := make(map[int]bool)
	response := make([]int, len(l))

	for i, n := range l {
		if !numbersMap[n] {
			response[i] = n
		}
		numbersMap[n] = true
	}

	sort.Ints(response)
	return response
}

func BenchmarkMapThenAppend(b *testing.B) {
	var r []int
	l := getNumbers()
	for n := 0; n < b.N; n++ {
		r = mapThenAppend(l)
	}
	benchResult = r
}

func BenchmarkMapAndAppend(b *testing.B) {
	var r []int
	l := getNumbers()
	for n := 0; n < b.N; n++ {
		r = mapAndAppend(l)
	}
	benchResult = r
}

func BenchmarkMapNoAppend(b *testing.B) {
	var r []int
	l := getNumbers()
	for n := 0; n < b.N; n++ {
		r = mapNoAppend(l)
	}
	benchResult = r
}
