package main

import (
	"context"
	"github.com/sirupsen/logrus"
)

func main() {
	Do(context.Background())
}

func Do(ctx context.Context) {
	logrus.Debug("main.do")
	return
}
