package main

import (
	"context"
	"gw-currency-wallet/internal/initenv"
	"gw-currency-wallet/internal/routes"

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

	routes.Start(ctx, cfgAdr, lg)
}
