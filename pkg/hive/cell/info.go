package cell

import (
	"fmt"
	"io"
	"strings"
)

const (
	identBy = 4
)

type InfoPrinter struct {
	io.Writer
	width int
}

type Info interface {
	Print(indent int, w *InfoPrinter)
}

type InfoNode struct {
	header    string
	condensed bool

	children []Info
}

func NewInfoNode(header string) *InfoNode {
	return &InfoNode{header: header}
}

func (n *InfoNode) Add(child Info) {
	n.children = append(n.children, child)
}

func (n *InfoNode) Print(indent int, w *InfoPrinter) {
	if n.header != "" {
		fmt.Fprintf(w, "%s%s:\n", strings.Repeat(" ", indent), n.header)
		indent += identBy
	}

	for i, child := range n.children {
		child.Print(indent, w)
		if !n.condensed && i != len(n.children)-1 {
			w.Write([]byte{'\n'})
		}
	}
}
