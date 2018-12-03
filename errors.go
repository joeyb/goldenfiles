package goldenfiles

import (
	"fmt"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type errGoldenFileMismatch struct {
	goldenFileName string

	golden   []byte
	received []byte
}

func (e errGoldenFileMismatch) Error() string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(e.golden), string(e.received), false)

	return fmt.Sprintf(
		"failed to verify %q, received value did not match golden file contents.\n\n%s\n",
		e.goldenFileName,
		diffmatchpatch.New().DiffPrettyText(diffs))
}

type errGoldenFileMissing struct {
	goldenFileName string

	received []byte
}

func (e errGoldenFileMissing) Error() string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(""), string(e.received), false)

	return fmt.Sprintf(
		"failed to verify %q, golden file missing.\n\n%s\n",
		e.goldenFileName,
		diffmatchpatch.New().DiffPrettyText(diffs))
}
