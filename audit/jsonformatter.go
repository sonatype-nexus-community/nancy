package audit

import (
	"encoding/json"

	. "github.com/sirupsen/logrus"
)

type JsonFormatter struct {
	PrettyPrint bool
}

func (f *JsonFormatter) Format(entry *Entry) ([]byte, error) {
	// Note this doesn't include Time, Level and Message which are available on
	// the Entry. Consult `godoc` on information about those fields or read the
	// source of the official loggers.
	if f.PrettyPrint {
		return json.MarshalIndent(entry.Data, "", "  ")
	} else {
		return json.Marshal(entry.Data)
	}
}
