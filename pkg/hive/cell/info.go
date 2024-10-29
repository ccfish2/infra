package cell

import "io"

type InfoPrinter struct {
	io.Writer
	width int
}
type Info interface {
	Print(indent int, w *InfoPrinter)
}
