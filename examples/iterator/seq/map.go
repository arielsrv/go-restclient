package seq

import "iter"

func Map[T any, U any](seq iter.Seq[T], predicate func(item T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for item := range seq {
			yield(predicate(item))
		}
	}
}
