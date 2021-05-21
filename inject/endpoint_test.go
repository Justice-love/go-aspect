package inject

import (
	"fmt"
	"testing"
)

func TestEndpoints(t *testing.T) {
	arr := Endpoints("/Users/xuyi/go/src/github.com/Justice-love/go-aspect")
	fmt.Println(arr)
}
