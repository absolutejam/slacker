package report

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshallJson(t *testing.T) {
	assert := assert.New(t)

	testFile := "examples/minimal.json"
	bytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("failed to read '%s'", testFile)
	}

	report, err := FromJson(bytes)
	assert.Equal(report.Environments[0].Name, "dev1")
}
