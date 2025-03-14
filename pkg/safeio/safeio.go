package safeio

import (
	"fmt"
	"io"
)

var ErrLimitReached = fmt.Errorf("reached limit")

type ByteSize float64

const (
	_           = iota
	KB ByteSize = 1 << (10 * iota) // 1KB
	MB
	GB
	TB
	PB
	EB
	YB
)

func (b ByteSize) String() string {
	switch {
	case b >= YB:
		return fmt.Sprintf("%.1fYB", b/YB)
	case b >= EB:
		return fmt.Sprintf("%.1fEB", b/EB)
	case b >= PB:
		return fmt.Sprintf("%.1fPB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%.1fTB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.1fGB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.1fMB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.1fKB", b/KB)
	default:
	}
	return fmt.Sprintf("%.1fB", b)
}

func ReadAllLimit(r io.Reader, n ByteSize) ([]byte, error) {
	limit := int(n + 1)
	buf, err := io.ReadAll(io.LimitReader(r, int64(limit)))
	if err != nil {
		return buf, err
	}
	if len(buf) >= limit {
		return buf[:limit-1], ErrLimitReached
	}
	return buf, nil
}
