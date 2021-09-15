package writer

import (
	"testing"
)

func TestReplaceFunctionName(t *testing.T) {
	ReplaceFunctionName("../testdata/source_file_write", "SourcePrettyText", 402)
}
