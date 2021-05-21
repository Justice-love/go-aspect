package file

import (
	"github.com/Justice-love/go-aspect/inject"
	"testing"
)

func TestX_IteratorSource(t *testing.T) {
	points := inject.Endpoints("/Users/xuyi/go/src/github.com/Justice-love/go-aspect/testdata")
	x := X{
		RootPath: "/Users/xuyi/go/src/github.com/Justice-love/go-aspect/testdata",
		Points:   points,
	}
	advices := x.IteratorSource("/Users/xuyi/go/src/github.com/Justice-love/go-aspect/testdata")
	inject.DoInjectCode(advices)
}
