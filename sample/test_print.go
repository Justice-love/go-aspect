package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

func BeforePrint(ctx context.Context) context.Context {
	log.Debug("test print")
	return context.WithValue(ctx, "date", time.Now())
}

func AfterPrint(ctx context.Context) {
	log.Debug(ctx.Value("date"))
}

func DoPrint() {
	log.Debug("Do")
}
