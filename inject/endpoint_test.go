package inject

import (
	"fmt"
	"testing"
)

func TestEndpoints(t *testing.T) {
	arr := endpoints("../testdata/aspect.point")
	fmt.Println(EndpointPrettyText(arr))
}
