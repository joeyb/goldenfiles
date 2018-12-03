package goldenfiles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrGoldenFileMismatch(t *testing.T) {
	err := errGoldenFileMismatch{
		goldenFileName: "test-goldenfile.txt",

		golden:   []byte("golden-data"),
		received: []byte("received-data"),
	}

	msg := err.Error()

	assert.Contains(t, msg, err.goldenFileName)
}

func TestErrGoldenFileMissing(t *testing.T) {
	err := errGoldenFileMissing{
		goldenFileName: "test-goldenfile.txt",

		received: []byte("received-data"),
	}

	msg := err.Error()

	assert.Contains(t, msg, err.goldenFileName)
}
