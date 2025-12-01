package main

import "github.com/rafibayer/ants-again/kdtree"

type KD[T any] interface {
	Insert(p T)
	Points() []T
	Remove(p T) T
	RangeSearch(r [][2]float64) []T
	KNN(p T, k int) []T
}

type KyroyKD[T kdtree.Point] struct {
	inner *kdtree.KDTree
}

var _ KD[Vector] = &KyroyKD[Vector]{}

func NewKyroyKD[T kdtree.Point](inner *kdtree.KDTree) KD[T] {
	return &KyroyKD[T]{inner: inner}
}

func (t *KyroyKD[T]) Insert(p T) {
	t.inner.Insert(p)
}

func (t *KyroyKD[T]) KNN(p T, k int) []T {
	return Map(t.inner.KNN(p, k), func(p kdtree.Point) T {
		return p.(T)
	})
}

func (t *KyroyKD[T]) Points() []T {
	return Map(t.inner.Points(), func(p kdtree.Point) T {
		return p.(T)
	})
}

func (t *KyroyKD[T]) RangeSearch(r [][2]float64) []T {
	return Map(t.inner.RangeSearch(r), func(p kdtree.Point) T {
		return p.(T)
	})
}

func (t *KyroyKD[T]) Remove(p T) T {
	return t.inner.Remove(p).(T)
}
