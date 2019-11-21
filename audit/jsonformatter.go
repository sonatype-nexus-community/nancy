package audit;

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/sirupsen/logrus"
)

type JsonFormatter struct {
	PrettyPrint bool
}

func (f *JsonFormatter) Format(entry *Entry) ([]byte, error) {
	// Note this doesn't include Time, Level and Message which are available on
	// the Entry. Consult `godoc` on information about those fields or read the
	// source of the official loggers.
	var b *bytes.Buffer = &bytes.Buffer{}
	encoder := json.NewEncoder(b)
	if f.PrettyPrint {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(entry.Data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}
	return b.Bytes(), nil
}
