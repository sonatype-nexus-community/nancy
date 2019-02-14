package buildversion

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultVersion(t *testing.T) {
	assert.Equal(t, "development", BuildVersion)
}

func TestDefaultBuildTime(t *testing.T) {
	assert.Equal(t, "", BuildTime)
}

func TestDefaultBuildCommit(t *testing.T) {
	assert.Equal(t, "", BuildCommit)
}
