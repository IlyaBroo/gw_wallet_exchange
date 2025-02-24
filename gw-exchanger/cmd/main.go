package main

import (
	"context"
	"gw-exchanger/internal/initenv"
	"gw-exchanger/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		cancel()
	}()

	lg, cfgAdr, err := initenv.InitEnv(ctx)
	if err != nil {
		lg.FatalCtx(ctx, "Error init env:", err)
	}

	err = server.Start(ctx, cfgAdr, lg)
	if err != nil {
		lg.FatalCtx(ctx, "Error start server:", err)
	}
}
