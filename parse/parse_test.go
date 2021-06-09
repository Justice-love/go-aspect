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
	if source.Funcs[20].NameLine != 402 {
		t.Fatal()
	}
	fmt.Println(SourcePrettyText([]*SourceStruct{source}))
}

func TestInlineFunction(t *testing.T) {
	f := "func funcMultiLine(reader *bufio.Reader, str string, line *int) *FuncStruct {"
	line := 10
	fu := funcInline(f, &line)
	fmt.Println(fu)
}
