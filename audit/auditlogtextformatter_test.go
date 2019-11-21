package audit;

import (
	"errors"
	. "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestFormatterErrorsIfEntryNotValid(t *testing.T) {
	data := map[string]interface{}{
		"stuff": 1,
		"another": "me",
	}
	entry := Entry{Data: data}

	formatter := AuditLogTextFormatter{}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, logMessage)
	assert.NotNil(t, e)
	assert.Equal(t, errors.New("fields passed did not match the expected values for an audit log. You should probably look at setting the formatter to something else"), e)
}
