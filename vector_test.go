package main

import (
	"testing"

	"github.com/kyroy/kdtree"
)

func TestFindWithin(t *testing.T) {
	tree := kdtree.New([]kdtree.Point{
		Vector{x: 0, y: 0},
		Vector{x: 1, y: 0},
		Vector{x: 1, y: 1},
	})

	found := KDSearchRadius(tree, Vector{0, 0}, 1.0)
	if len(found) != 2 {
		t.FailNow()
	}
}
