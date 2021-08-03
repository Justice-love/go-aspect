package inject

import (
	"fmt"
	"github.com/Justice-love/go-aspect/parse"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDuplicate(t *testing.T) {
	function := &parse.FuncStruct{}
	advice := []*Advice{
		{
			Source: &parse.SourceStruct{},
			Aspect: []*Aspect{
				{
					Function: function,
				},
				{
					Function: function,
				},
			},
		},
	}
	assert.Panics(t, func() {
		DoInjectCode(advice)
	})
}

func TestAround(t *testing.T) {
	s := &parse.SourceStruct{
		Path:     "../testdata/source_file_write",
		FileLine: 468,
	}
	a := &Aspect{
		Function: &parse.FuncStruct{
			FuncString: "func Contain(sourceStruct *SourceStruct, i *ImportStruct) bool {",
			FuncName:   "Contain",
			Returns:    []parse.ReturnType{1},
			NameLine:   390,
			Params: []*parse.ParamStruct{
				{
					Name: "sourceStruct",
				},
				{
					Name: "i",
				},
			},
		},
	}
	(&AroundInjectFile{}).InjectFunc(s, a)
}

func TestN(t *testing.T) {
	s := []string{"1"}
	a := s[1:]
	fmt.Println(a)
}
