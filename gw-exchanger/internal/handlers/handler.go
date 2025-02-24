package handlers

import (
	"context"
	"fmt"
	"gw-exchanger/internal/cache"
	"gw-exchanger/internal/config"
	"gw-exchanger/internal/logger"
	"gw-exchanger/internal/storages"
	"math"
	"time"

	exchange "github.com/IlyaBroo/exchange_grpc/exchange"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	exchange.UnimplementedExchangeServiceServer
	lg    logger.Logger
	db    storages.RepositoryInterface
	cache *cache.Cache
}

func NewServer(lg logger.Logger, ctx context.Context, cfg *config.ConfigAdr) *Server {
	db := storages.NewRepository(lg, ctx, cfg)
	cache := cache.NewCache(5 * time.Minute)
	s := new(Server)
	s.lg = lg
	s.db = db
	s.cache = cache
	return s

}

func (s *Server) GetExchangeRates(ctx context.Context, in *exchange.Empty) (*exchange.ExchangeRatesResponse, error) {
	reqID := s.getIDFromContext(ctx)
	ctx = context.WithValue(ctx, "requestID", reqID)
	s.lg.InfoCtx(ctx, fmt.Sprintf("Received request with ID: %s", reqID))
	excRateResponse := new(exchange.ExchangeRatesResponse)

	cachedRates := s.cache.GetAll()
	if len(cachedRates) > 0 {
		s.lg.InfoCtx(ctx, "Returning cached exchange rates")
		excRateResponse.Rates = cachedRates
		return excRateResponse, nil
	}

	res, err := s.db.GetRates(ctx)
	if err != nil {
		s.lg.ErrorCtx(ctx, "GetExchangeRates failed")
		return nil, err
	}
	s.cache.Set(res)
	s.lg.InfoCtx(ctx, "Set cache")

	excRateResponse.Rates = res

	s.lg.InfoCtx(ctx, fmt.Sprintf("ExchangeRateResponse : %v", excRateResponse.Rates))
	return excRateResponse, nil
}

func (s *Server) GetExchangeRateForCurrency(ctx context.Context, in *exchange.CurrencyRequest) (*exchange.ExchangeRateResponse, error) {
	reqID := s.getIDFromContext(ctx)
	ctx = context.WithValue(ctx, "requestID", reqID)
	s.lg.InfoCtx(ctx, fmt.Sprintf("Received request with ID: %s", reqID))
	excRateResponse := new(exchange.ExchangeRateResponse)
	if in.FromCurrency == in.ToCurrency {
		s.lg.InfoCtx(ctx, "From and to currency are the same")
		return nil, fmt.Errorf("from and to currency are the same")
	}

	keystring := fmt.Sprintf("%s%s", in.FromCurrency, in.ToCurrency)

	cachedRate, ok := s.cache.GetSpecificRate(keystring)
	if ok {
		s.lg.InfoCtx(ctx, "Returning cached from special exchange rate")
		excRateResponse.Rate = cachedRate
		return excRateResponse, nil
	}
	cachedRates := s.cache.GetAll()
	if len(cachedRates) > 0 {
		s.lg.InfoCtx(ctx, "Returning cached from all exchange rate")
		excRateResponse.Rate = calculateRate(cachedRates[in.FromCurrency], cachedRates[in.ToCurrency])
		return excRateResponse, nil
	}
	res, err := s.db.GetRatesForCurrency(ctx, in.FromCurrency, in.ToCurrency)
	if err != nil {
		s.lg.ErrorCtx(ctx, "GetExchangeRateForCurrency failed")
		return nil, err
	}
	s.cache.SetSpecificRate(keystring, res)
	excRateResponse.Rate = res
	s.lg.InfoCtx(ctx, fmt.Sprintf("ExchangeRateResponse : %v", excRateResponse.Rate))
	return excRateResponse, nil
}

func (s *Server) getIDFromContext(ctx context.Context) string {

	arraystring := metadata.ValueFromIncomingContext(ctx, "requestID")

	if len(arraystring) == 0 {
		s.lg.ErrorCtx(ctx, "requestID not found")
		return "unknown"
	}
	id := arraystring[0]
	s.lg.DebugCtx(ctx, fmt.Sprintf("id is %s", id))
	return id
}

func calculateRate(from, to float32) float32 {
	res := 1 / from * to
	res = float32(math.Round(float64(res)*100) / 100)
	return res
}
