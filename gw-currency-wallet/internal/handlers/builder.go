package handlers

import (
	"context"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/logger"
	"gw-currency-wallet/internal/storages"
	"net/http"

	exchange "github.com/IlyaBroo/exchange_grpc/exchange"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

type ServerWallet struct {
	HttpClient *http.Client
	db         storages.RepositoryInterface
	lg         logger.Logger
	grpcclient exchange.ExchangeServiceClient
}

type ErrorResponse struct {
	message string `json:"error"`
}

func NewServerWallet(httpClient *http.Client, lg logger.Logger, cfg *config.ConfigAdr, ctx context.Context) (*ServerWallet, error) {

	conn, err := grpc.Dial(cfg.Grpc_Adr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	grpcClient := exchange.NewExchangeServiceClient(conn)

	db := storages.NewRepository(lg, ctx, cfg)
	s := new(ServerWallet)
	s.HttpClient = httpClient
	s.lg = lg
	s.db = db
	s.grpcclient = grpcClient
	return s, nil
}
