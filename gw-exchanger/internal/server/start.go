package server

import (
	"context"
	"fmt"
	"gw-exchanger/internal/config"
	"gw-exchanger/internal/handlers"
	"gw-exchanger/internal/logger"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	exchange "github.com/IlyaBroo/exchange_grpc/exchange"
	"google.golang.org/grpc"
)

func Start(ctx context.Context, cfg *config.ConfigAdr, lg logger.Logger) error {
	lg.InfoCtx(ctx, "Starting server ...")
	lis, err := net.Listen("tcp", cfg.APP_ADR)
	if err != nil {
		lg.ErrorCtx(ctx, fmt.Sprintf("Failed to listen: %v", err))
		return err
	}
	s := grpc.NewServer()

	server := handlers.NewServer(lg, ctx, cfg)
	exchange.RegisterExchangeServiceServer(s, server)
	lg.InfoCtx(ctx, fmt.Sprintf("gRPC server listening on port %s", cfg.APP_ADR))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.Serve(lis); err != nil {
			lg.FatalCtx(ctx, "error serving", err)
		}
	}()

	<-stop
	lg.InfoCtx(ctx, "Shutting down server...")
	ctxout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lg.InfoCtx(ctx, "Calling GracefulStop...")
	s.GracefulStop()
	lg.InfoCtx(ctx, "GracefulStop called, waiting for server to finish...")

	select {
	case <-ctxout.Done():
		lg.WarnCtx(ctx, "Server timeout reached, forcing exit")
	case <-time.After(5 * time.Second):
		lg.InfoCtx(ctx, "Server shutdown completed")
	}

	return nil
}
