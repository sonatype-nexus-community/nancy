package audit

import (
	"errors"
	. "github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCsvOutputWhenQuiet(t *testing.T) {
	data := map[string]interface{}{
		"audited": []types.Coordinate{
			{Coordinates: "good1"},
			{Coordinates: "vuln1", Vulnerabilities: createVulnerabilities(1)},
		},
		"invalid": []types.Coordinate{
			{InvalidSemVer: true, Coordinates: "invalid1"},
		},
		"num_audited":    2,
		"num_vulnerable": 1,
		"version":        "development",
	}
	entry := Entry{Data: data}

	quiet := true
	formatter := CsvFormatter{Quiet: &quiet}
	logMessage, e := formatter.Format(&entry)
	assert.Nil(t, e)
	expectedCsv := `Summary
Audited Count,Vulnerable Count,Build Version
2,1,development

Audited Package(s)
Count,Package,Is Vulnerable,Num Vulnerabilities,Vulnerabilities
[2/2],vuln1,true,1,"[{""Id"":""123"",""Title"":""Vulnerability"",""Description"":""Description"",""CvssScore"":""7.88"",""CvssVector"":""What"",""Cve"":""CVE-123"",""Reference"":""Reference"",""Excluded"":false}]"
`
	assert.Equal(t, expectedCsv, string(logMessage))
}

func TestCsvOutput(t *testing.T) {
	data := map[string]interface{}{
		"audited": []types.Coordinate{
			{Coordinates: "good1"},
			{Coordinates: "vuln1", Vulnerabilities: createVulnerabilities(1)},
		},
		"invalid": []types.Coordinate{
			{InvalidSemVer: true, Coordinates: "invalid1"},
		},
		"num_audited":    2,
		"num_vulnerable": 1,
		"version":        "development",
	}
	entry := Entry{Data: data}

	quiet := false
	formatter := CsvFormatter{Quiet: &quiet}
	logMessage, e := formatter.Format(&entry)
	assert.Nil(t, e)
expectedCsv := `Summary
Audited Count,Vulnerable Count,Build Version
2,1,development

Invalid Package(s)
Count,Package,Reason
[1/1],invalid1,Does not use SemVer

Audited Package(s)
Count,Package,Is Vulnerable,Num Vulnerabilities,Vulnerabilities
[1/2],good1,false,0,null
[2/2],vuln1,true,1,"[{""Id"":""123"",""Title"":""Vulnerability"",""Description"":""Description"",""CvssScore"":""7.88"",""CvssVector"":""What"",""Cve"":""CVE-123"",""Reference"":""Reference"",""Excluded"":false}]"
`
	assert.Equal(t, expectedCsv, string(logMessage))
}

func TestCsvOutputWhenNotAuditLog(t *testing.T) {
	data := map[string]interface{}{
		"stuff":   1,
		"another": "me",
	}
	entry := Entry{Data: data}

	formatter := CsvFormatter{}
	logMessage, e := formatter.Format(&entry)

	assert.Nil(t, logMessage)
	assert.NotNil(t, e)
	assert.Equal(t, errors.New("fields passed did not match the expected values for an audit log. You should probably look at setting the formatter to something else"), e)
}
