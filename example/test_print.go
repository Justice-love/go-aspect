package example

import (
	"context"
	"fmt"
	"time"
)

func BeforePrint(ctx context.Context) context.Context {
	fmt.Println("before print")
	return context.WithValue(ctx, "date", time.Now())
}

func AfterPrint(ctx context.Context) {
	fmt.Println(ctx.Value("date"))
}

func DoPrint() {
	fmt.Println("Do")
}
