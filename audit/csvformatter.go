package audit;

import (
	"bytes"
	"encoding/csv"
	"fmt"
	. "github.com/sirupsen/logrus"
	"sort"
)

type CsvFormatter struct {
}

func (f *CsvFormatter) Format(entry *Entry) ([]byte, error) {
	// Note this doesn't include Time, Level and Message which are available on
	// the Entry. Consult `godoc` on information about those fields or read the
	// source of the official loggers.
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var header []string
	var body []string
	for _, k := range keys {
		header = append(header, k)
		body = append(body, fmt.Sprintf("%v", entry.Data[k]))
	}
	w.Write(header)
	w.Write(body)
	w.Flush()

	return buf.Bytes(), nil
}
