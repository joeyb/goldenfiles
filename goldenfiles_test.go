package goldenfiles

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestEqualString(t *testing.T) {
	testCases := []struct {
		name   string
		config *Config
		s      string
	}{
		{
			"DefaultConfig",
			nil,
			"default config test\n",
		},
		{
			"CustomConfig",
			&Config{Name: "custom-name", Suffix: ".test"},
			"custom config test\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.config == nil {
				assert.True(t, EqualString(t, tc.s))
			} else {
				assert.True(t, EqualString(t, tc.s, *tc.config))
			}
		})
	}
}

func TestEqualStringGoldenFileError(t *testing.T) {
	testCases := []struct {
		name string
		s    string
	}{
		{
			"GoldenFileMissing",
			"missingtest1\nmissingtest2\n",
		},
		{
			"GoldenFileMismatch",
			"mismatchtest1\nmismatchtest2\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockT := &mockT{realT: t}

			assert.False(t, EqualString(mockT, tc.s))

			assert.True(t, mockT.errorfCalled)

			goldenFileName := buildGoldenFileName(t, &Config{})
			receivedFilePath := buildReceivedFilePath(goldenFileName)
			defer removeReceivedFile(t, receivedFilePath)

			assert.FileExists(t, receivedFilePath)

			received, err := ioutil.ReadFile(receivedFilePath)
			require.NoError(t, err)

			assert.Equal(t, tc.s, string(received))
		})
	}
}

func TestEqualJSON(t *testing.T) {
	testCases := []struct {
		name   string
		config *Config
		obj    interface{}
	}{
		{
			"DefaultConfig",
			nil,
			map[string]string{"key1": "defaultConfigValue1", "key2": "defaultConfigValue2"},
		},
		{
			"CustomConfig",
			&Config{Name: "custom-name", Suffix: ".test"},
			map[string]string{"key1": "customConfigValue1", "key2": "customConfigValue2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.config == nil {
				assert.True(t, EqualJSON(t, tc.obj))
			} else {
				assert.True(t, EqualJSON(t, tc.obj, *tc.config))
			}
		})
	}
}

func TestEqualJSONGoldenFileError(t *testing.T) {
	testCases := []struct {
		name string
		obj  interface{}
	}{
		{
			"GoldenFileMissing",
			map[string]string{"key1": "missingValue1", "key2": "missingValue2"},
		},
		{
			"GoldenFileMismatch",
			map[string]string{"key1": "mismatchValue1", "key2": "mismatchValue2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockT := &mockT{realT: t}

			EqualJSON(mockT, tc.obj)

			assert.True(t, mockT.errorfCalled)

			goldenFileName := buildGoldenFileName(t, &Config{Suffix: ".json"})
			receivedFilePath := buildReceivedFilePath(goldenFileName)
			defer removeReceivedFile(t, receivedFilePath)

			assert.FileExists(t, receivedFilePath)

			received, err := ioutil.ReadFile(receivedFilePath)
			require.NoError(t, err)

			expected, err := prettyPrintJSON(tc.obj)
			require.NoError(t, err)

			assert.Equal(t, expected, string(received))
		})
	}
}

func TestEqualJSONFailureToMarshalToJSON(t *testing.T) {
	mockT := &mockT{realT: t}

	assert.False(t, EqualJSON(mockT, make(chan int)))

	assert.True(t, mockT.errorfCalled)
}

type mockT struct {
	realT *testing.T

	errorfCalled  bool
	failNowCalled bool
}

func (t *mockT) Errorf(format string, args ...interface{}) {
	t.realT.Logf(format, args...)

	t.errorfCalled = true
}

func (t *mockT) FailNow() {
	t.failNowCalled = true
}

func (t *mockT) Name() string {
	return t.realT.Name()
}

func removeReceivedFile(t *testing.T, receivedFilePath string) {
	if err := os.Remove(receivedFilePath); err != nil {
		t.Errorf("failed to removed received file: %v", err)
		t.FailNow()
	}
}
