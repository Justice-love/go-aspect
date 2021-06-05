package parse

import (
	"fmt"
	"testing"
)

func TestTestDataParse(t *testing.T) {
	source := SourceParse("../testdata/parse_file")
	if source == nil {
		t.Fatal()
	}
	fmt.Println(SourcePrettyText([]*SourceStruct{source}))
}

func TestSourceParse(t *testing.T) {
	source := SourceParse("../testdata/source_file")
	if source == nil {
		t.Fatal()
	}
	fmt.Println(SourcePrettyText([]*SourceStruct{source}))
}
