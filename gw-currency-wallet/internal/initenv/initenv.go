package initenv

import (
	"context"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/logger"
	"log"
)

func InitEnv(ctx context.Context) (logger.Logger, *config.ConfigAdr, error) {

	cfg, cfgAdr, err := config.LoadConfig("internal/config/config.yaml")
	if err != nil {
		log.Printf("error load config: %v", err)
		return nil, nil, err
	}
	lg, err := logger.NewLogger(logger.WithCfg(cfg))
	if err != nil {
		log.Printf("error create new logger: %v", err)
		return nil, nil, err
	}

	lg.InfoCtx(ctx, "Starting service...")

	return lg, cfgAdr, err
}
