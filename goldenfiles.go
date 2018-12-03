package goldenfiles

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/kennygrant/sanitize"
)

var (
	// GoldenFilePath defines the where golden files are expected to be stored. If there is a mismatch, the received
	// output is also written to that directory.
	GoldenFilePath = "./testdata"

	// ReceivedFilePerms defines the file permissions that are set on any received output files that are created.
	ReceivedFilePerms os.FileMode = 0644
)

// TestingT defines the portion of the testing.T interface that we need to use. Extracting it into an interface allows
// us to create a mock implementation for our tests.
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
	Name() string
}

// Config defines additional options that can be overridden for each golden file assertion.
type Config struct {
	// Name provides an additional name segment which gets added to the golden file name. This is especially useful for
	// tests that need to make multiple golden file assertions.
	Name string

	// Suffix overrides the default golden file suffix for the current test type. For example, EqualJSON defaults to
	// the ".json" suffix.
	Suffix string
}

// EqualString tests for golden file equality against a string.
func EqualString(t TestingT, s string, config ...Config) bool {
	c := getConfig(config)

	goldenFileName := buildGoldenFileName(t, &c)
	err := compare(t, goldenFileName, strings.NewReader(s))
	if err != nil {
		switch err := err.(type) {
		case errGoldenFileMismatch:
			t.Errorf("%s", err.Error())
			writeReceivedFile(t, goldenFileName, err.received)
		case errGoldenFileMissing:
			t.Errorf("%s", err.Error())
			writeReceivedFile(t, goldenFileName, err.received)
		default:
			t.Errorf("%s", err.Error())
			t.FailNow()
		}
		return false
	}

	return true
}

// EqualJSON tests for golden file equality against an object that gets marshalled into a JSON string.
func EqualJSON(t TestingT, obj interface{}, config ...Config) bool {
	c := getConfig(config)

	json, err := prettyPrintJSON(obj)
	if err != nil {
		t.Errorf("failed to serialize received object as json: %v", err)
		return false
	}

	if c.Suffix == "" {
		c.Suffix = ".json"
	}

	return EqualString(t, json, c)
}

func buildGoldenFileName(t TestingT, c *Config) string {
	suffix := ".txt"
	if c.Suffix != "" {
		suffix = c.Suffix
	}

	name := c.Name
	if name != "" {
		name = "." + name
	}

	testName := t.Name()
	sanitizedTestName := sanitize.BaseName(testName)

	return fmt.Sprintf("%s%s.golden%s", sanitizedTestName, name, suffix)
}

func buildReceivedFileName(goldenFileName string) string {
	return fmt.Sprintf("%s.received", goldenFileName)
}

func buildReceivedFilePath(goldenFileName string) string {
	return path.Join(GoldenFilePath, buildReceivedFileName(goldenFileName))
}

func compare(t TestingT, goldenFileName string, r io.Reader) error {
	s, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	f, err := os.Open(path.Join(GoldenFilePath, goldenFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return errGoldenFileMissing{goldenFileName: goldenFileName, received: s}
		}
		return err
	}
	defer f.Close()

	golden, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	if bytes.Compare(golden, s) == 0 {
		// Received matches golden, success!
		return nil
	}

	return errGoldenFileMismatch{goldenFileName: goldenFileName, golden: golden, received: s}
}

func getConfig(c []Config) Config {
	if len(c) > 0 {
		return c[0]
	}

	return Config{}
}

func prettyPrintJSON(obj interface{}) (string, error) {
	b, err := json.MarshalIndent(obj, "", "  ")

	return string(b), err
}

func writeReceivedFile(t TestingT, goldenFileName string, received []byte) {
	receivedFilePath := buildReceivedFilePath(goldenFileName)

	if err := ioutil.WriteFile(receivedFilePath, received, ReceivedFilePerms); err != nil {
		t.Errorf("failed to write received file: %v", err)
		t.FailNow()
	}
}
