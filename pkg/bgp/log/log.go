package log

import (
	"fmt"
	"strings"

	"github.com/ccfish2/infra/pkg/option"
	"github.com/sirupsen/logrus"
)

func (l *Logger) Log(args ...interface{}) error {
	if !option.Config.Debug {
		return nil
	}
	b := strings.Builder{}
	for _, a := range args {
		switch v := a.(type) {
		case string:
			b.WriteString(v)
			b.WriteString(" ")
		case error:
			b.WriteString(v.Error())
			b.WriteString(" ")
		case fmt.Stringer:
			b.WriteString(v.String())
			b.WriteString(" ")
		default:
			break
		}
	}
	l.Debug(strings.TrimSpace(b.String()))
	return nil
}

type Logger struct {
	*logrus.Entry
}
