package audit

import (
	. "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJsonOutpu(t *testing.T) {
	data := map[string]interface{}{
		"stuff":   1,
		"another": "me",
	}
	entry := Entry{Data: data}

	formatter := JsonFormatter{}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, e)
	assert.Equal(t, `{"another":"me","stuff":1}`, string(logMessage))
}

func TestJsonPrettyPrintOutpu(t *testing.T) {
	data := map[string]interface{}{
		"stuff":   1,
		"another": "me",
	}
	entry := Entry{Data: data}

	formatter := JsonFormatter{PrettyPrint: true}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, e)
	assert.Equal(t, `{
  "another": "me",
  "stuff": 1
}`, string(logMessage))
}
