package main

import "testing"

func BenchmarkUpdate(b *testing.B) {
	// todo: new Params in case default changes, don't want it to affect benchmarks
	g := NewGame(nil)
	b.ResetTimer()

	b.ReportAllocs()

	for b.Loop() {
		g.Update()
	}
}
