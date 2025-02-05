package main

import (
	"cmp"
	"fmt"
	"iter"
	"slices"
)

type List[T cmp.Ordered] struct {
	items   []T
	sorting Sorting[T]
}

func NewList[T cmp.Ordered](items []T) *List[T] {
	return &List[T]{
		items:   items,
		sorting: &QuickSort[T]{},
	}
}

func (r *List[T]) Sorting(sort Sorting[T]) {
	r.sorting = sort
}

func (r *List[T]) Sort() {
	r.sorting.Sort(r.items)
}

func (r *List[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := range r.items {
			if !yield(r.items[i]) {
				return
			}
		}
	}
}

type Sorting[T cmp.Ordered] interface {
	Sort(items []T)
}

type QuickSort[T cmp.Ordered] struct{}

func (r QuickSort[T]) Sort(items []T) {
	slices.Sort(items)
}

type MergeSort[T cmp.Ordered] struct{}

func (r MergeSort[T]) Sort(items []T) {
	slices.Sort(items)
}

func main() {
	numbers := NewList[int]([]int{5, 2, 8, 1, 9, 3, 6, 4, 7})
	numbers.Sorting(&MergeSort[int]{})
	numbers.Sort()

	for item := range numbers.All() {
		fmt.Println(item) // Output: 1 2 3 4 5 6 7 8 9
	}
}
