package slices

import (
	"slices"

	"golang.org/x/exp/constraints"
)

func SortedUniqs[S ~[]T, T constraints.Ordered](s S) S {
	return slices.Compact(s)
}
