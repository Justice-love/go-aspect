package parse

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
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

func TestImportParse(t *testing.T) {
	t.Run("import multiple line", func(t *testing.T) {
		a := assert.New(t)
		i := "\t\"bufio\"\n\t\"bytes\"\n\t\"fmt\"\n\t\"github.com/Justice-love/go-aspect/util\"\n\tlog \"github.com/sirupsen/logrus\"\n\t\"io\"\n\t\"os\"\n\t\"strings\"\n)\n"
		reader := bufio.NewReader(strings.NewReader(i))
		is, _ := importParse(reader, "import (")
		a.Equal(8, len(is))
		a.Equal("log", is[4].ImportTag)
	})
	t.Run("import inline", func(t *testing.T) {
		a := assert.New(t)
		i := "import \"bytes\""
		reader := bufio.NewReader(strings.NewReader(i))
		is, _ := importParse(reader, "import \"bytes\"")
		a.Equal(1, len(is))
	})
	t.Run("function inline", func(t *testing.T) {
		a := assert.New(t)
		i := "func TestImportParse(t *testing.T) {"
		line := 10
		is := funcInline(i, &line)
		a.NotNil(is)
		a.Equal(1, len(is.Params))
		a.True(is.Params[0].Pointer)
		a.Equal("T", is.Params[0].ParamType)
		a.Equal("t", is.Params[0].Name)
		a.Equal("TestImportParse", is.FuncName)
	})
}
