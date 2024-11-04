package cell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/term"
)

const (
	identBy = 4
)

type InfoPrinter struct {
	io.Writer
	width int
}

func NewInfoPrinter() *InfoPrinter {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 120
	}
	return &InfoPrinter{
		Writer: os.Stdout,
		width:  width,
	}
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

type InfoLeaf string

func (l InfoLeaf) Print(indent int, w *InfoPrinter) {
	buf := bufio.NewWriter(w)
	indentString := strings.Repeat(" ", indent)
	buf.WriteString(indentString)
	currentLineLength := len(indentString)
	wrapped := false
	for _, f := range strings.Fields(string(l)) {
		newLineLength := currentLineLength + len(f) + 1
		if newLineLength >= w.width {
			buf.WriteByte('\n')
			if !wrapped {
				// Increase the indent for the wrapped lines so it's clear we
				// wrapped.
				wrapped = true
				indent += 2
				indentString = strings.Repeat(" ", indent)
			}
			buf.WriteString(indentString)
			currentLineLength = indent + len(f) + 1
		} else {
			currentLineLength = newLineLength
		}
		buf.WriteString(f)
		buf.WriteByte(' ')
	}
	buf.WriteByte('\n')
	buf.Flush()
}

func (n *InfoNode) AddLeaf(format string, args ...any) {
	n.Add(InfoLeaf(fmt.Sprintf(format, args...)))
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

type InfoStruct struct {
	value any
}

/**/
func (n *InfoStruct) Print(indent int, w *InfoPrinter) {
	scs := spew.ConfigState{Indent: strings.Repeat(" ", identBy), SortKeys: true}
	indentString := strings.Repeat(" ", indent)
	for i, line := range strings.Split(scs.Sdump(n.value), "\n") {
		if i == 0 {
			fmt.Fprintf(w, "%s⚙️ %s\n", indentString, line)
		} else {
			fmt.Fprintf(w, "%s%s\n", indentString, line)
		}
	}
}
