package inject

import (
	"fmt"
	"testing"
)

func TestEndpoints(t *testing.T) {
	arr := Endpoints("/Users/xuyi/go/src/eddy.org/go-aspect")
	fmt.Println(arr)
}
