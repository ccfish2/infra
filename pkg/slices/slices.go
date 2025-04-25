package slices

import (
	"slices"

	"golang.org/x/exp/constraints"
)

// deduplicates the elements in the input slice, preserving their ordering and
// modifying the slice in place.
func Unique[S ~[]T, T comparable](s S) S {
	if len(s) < 2 {
		return s
	}

	last := 0

	if len(s) < 192 {
	Loop:
		for i := 0; i < len(s); i++ {
			for j := 0; j < last; j++ {
				if s[i] == s[j] {
					continue Loop
				}
			}
			s[last] = s[i]
			last++
		}
	} else {
		set := make(map[T]struct{}, len(s))
		for i := 0; i < len(s); i++ {
			if _, ok := set[s[i]]; ok {
				continue
			}
			set[s[i]] = struct{}{}
			s[last] = s[i]
			last++
		}
	}

	return s[:last]
}

func SortedUniqs[S ~[]T, T constraints.Ordered](s S) S {
	return slices.Compact(s)
}
