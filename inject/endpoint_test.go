package inject

import (
	"fmt"
	"testing"
)

func TestEndpoints(t *testing.T) {
	arr := endpoints("../testdata/aspect_around.point")
	fmt.Println(EndpointPrettyText(arr))
}
