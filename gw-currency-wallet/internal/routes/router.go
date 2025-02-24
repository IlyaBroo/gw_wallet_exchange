// @title Currency Wallet API
// @version 1.0
// @description API for managing currency wallet and exchange rates
// @host localhost:8080

package routes

import (
	"context"
	"fmt"
	_ "gw-currency-wallet/docs"
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/handlers"
	"gw-currency-wallet/internal/logger"
	"gw-currency-wallet/internal/middleware"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
)

func Start(ctx context.Context, cfg *config.ConfigAdr, lg logger.Logger) {

	r := chi.NewRouter()
	httpClient := &http.Client{Timeout: time.Second * 5}

	defer func() {
		if err := recover(); err != nil {
			lg.ErrorCtx(ctx, fmt.Sprintf("Паника в функции Start: ", err))
		}
	}()

	h, err := handlers.NewServerWallet(httpClient, lg, cfg, ctx)
	if err != nil {
		lg.FatalCtx(ctx, "Error create new wallet handler: ", err)
	}

	r.Use(middleware.ContextRequestMiddleware)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(cfg.Swagger_url),
	))

	r.Post("/register", h.RegisterUser)
	r.Post("/login", h.Autherisation)
	r.Group(func(r chi.Router) {
		r.Use(middleware.ValidateJWT)
		r.Get("/balance", h.GetBalance)
		r.Post("/deposit", h.Deposit)
		r.Post("/withdraw", h.Withdraw)
		r.Get("/rates", h.ExchangeRates)
		r.Post("/exchange", h.ExchangeRatesForCurrency)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.APP_ADR),
		Handler: r,
	}
	go func() {
		lg.InfoCtx(ctx, "Сервер  запускается на порту"+cfg.APP_ADR)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.ErrorCtx(ctx, "Ошибка запуска сервера: "+err.Error())
		}
	}()

	<-ctx.Done()
	shutdownCtxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtxt); err != nil {
		lg.ErrorCtx(ctx, "Ошибка завершения сервера: "+err.Error())
	} else {
		lg.InfoCtx(ctx, "Сервер коректно завершен")
	}
}
