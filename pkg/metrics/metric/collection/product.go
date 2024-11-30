package collections

func CartesianProduct[T any](vs ...[]T) [][]T {
	if len(vs) == 0 {
		return [][]T{}
	}

	dimension := len(vs) // Each output will be a vector of this length.
	// Iterate to find out the number of output vectors.
	size := len(vs[0])
	for i := 1; i < len(vs); i++ {
		size *= len(vs[i])
	}

	// Allocate the output vectors.
	dst := make([][]T, size)
	for i := range dst {
		dst[i] = make([]T, dimension)
	}

	lastm := 1
	for i := 0; i < dimension; i++ {
		permuteColumn[T](dst, i, lastm, vs[i])
		lastm = lastm * len(vs[i])
	}
	return dst
}

func permuteColumn[T any](dst [][]T, col int, leftPermSize int, vec []T) {
	for i := 0; i < len(dst); i += leftPermSize { // So we're skipping n rows at a time,
		vi := (i / leftPermSize) % len(vec)
		for off := 0; off < leftPermSize; off++ { // this is a repeat
			dst[i+off][col] = vec[vi]
		}
	}
}
