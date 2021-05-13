package testdata

import (
	"context"
	"fmt"
	"time"
)

type X struct {
}

func (*X) Some(ctx context.Context) {
	ctx = context.WithValue(ctx, "date", time.Now())

	fmt.Println("abc")
	fmt.Println(ctx.Value("date"))
	fmt.Println("456")
	fmt.Println("789")

}
