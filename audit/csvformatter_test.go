package audit

import (
	. "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCsvOutpu(t *testing.T) {
	data := map[string]interface{}{
		"stuff":   1,
		"another": "me",
	}
	entry := Entry{Data: data}

	formatter := CsvFormatter{}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, e)
	assert.Equal(t, "another,stuff\nme,1\n", string(logMessage))
}
