package file

import (
	"eddy.org/go-aspect/inject"
	"testing"
)

func TestX_IteratorSource(t *testing.T) {
	points := inject.Endpoints("/Users/xuyi/go/src/eddy.org/go-aspect/testdata")
	x := X{
		RootPath: "/Users/xuyi/go/src/eddy.org/go-aspect/testdata",
		Points:   points,
	}
	advices := x.IteratorSource("/Users/xuyi/go/src/eddy.org/go-aspect/testdata")
	inject.DoInjectCode(advices)
}
