package handlers

import (
	"encoding/json"
	"fmt"
	"gw-currency-wallet/internal/middleware"
	"gw-currency-wallet/internal/storages"
	"net/http"

	exchange "github.com/IlyaBroo/exchange_grpc/exchange"
	"google.golang.org/grpc/metadata"
)

type ExchangeResponse struct {
	Rates map[string]float32 `json:"rates"`
}

type ExchangeForCurrencyReq struct {
	From   string  `json:"from_currency"`
	To     string  `json:"to_currency"`
	Amount float64 `json:"amount"`
}

type ExchangeResponseForCurrency struct {
	Message     string             `json:"message"`
	Amount      float64            `json:"amount"`
	New_balance map[string]float64 `json:"new_balance"`
}

// @Summary Получение курсов валют
// @Description Позволяет получить актуальные курсы валют из внешнего gRPC-сервиса.
// @Tags exchange
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT_TOKEN"
// @Success 200 {object} ExchangeResponse "rates:"
// @Failure 500 {object} ErrorResponse "Failed to retrieve exchange rates"
// @Router /rates [get]
func (s *ServerWallet) ExchangeRates(w http.ResponseWriter, r *http.Request) {
	var exchangeRes ExchangeResponse
	errRes := new(ErrorResponse)
	reqId := r.Context().Value("requestID").(string)
	ctx := metadata.AppendToOutgoingContext(r.Context(), "requestID", reqId)
	in := new(exchange.Empty)

	res, err := s.grpcclient.GetExchangeRates(ctx, in)
	if err != nil {
		s.lg.ErrorCtx(ctx, err.Error())
		errRes.message = "Failed to retrieve exchange rates"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Failed to retrieve exchange rates", http.StatusInternalServerError)
		return
	}
	exchangeRes.Rates = res.Rates
	s.lg.InfoCtx(ctx, fmt.Sprintf("rates: %v", res.Rates))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exchangeRes)
}

// @Summary Обмен валют
// @Description Позволяет обменять одну валюту на другую. Проверяет наличие средств для обмена и обновляет баланс пользователя.
// @Tags exchange
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT_TOKEN"
// @Param exchange body ExchangeForCurrencyReq true "Данные для обмена валют"
// @Success 200 {object} ExchangeResponseForCurrency "Successfully exchanged currency"
// @Failure 400 {object} ErrorResponse "Error decoding currency request"
// @Failure 400 {object} ErrorResponse "Insufficient funds or invalid amount"
// @Failure 500 {object} ErrorResponse "Error fetching exchange rate"
// @Failure 500 {object} ErrorResponse "Error exchanging currency"
// @Router /exchange [post]
func (s *ServerWallet) ExchangeRatesForCurrency(w http.ResponseWriter, r *http.Request) {
	errRes := new(ErrorResponse)
	s.lg.InfoCtx(r.Context(), "Exchange rates for currency")
	reqId := r.Context().Value("requestID").(string)
	user_id := r.Context().Value(middleware.User_id).(int)
	ctx := metadata.AppendToOutgoingContext(r.Context(), "requestID", reqId)
	s.lg.InfoCtx(ctx, "Exchange rates for currency")
	req := new(ExchangeForCurrencyReq)
	in := new(exchange.CurrencyRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.lg.ErrorCtx(ctx, fmt.Sprintf("error decoding json: %v", err))
		errRes.message = "Error decoding currency request"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Error decoding currency request", http.StatusBadRequest)
		return
	}
	in.FromCurrency = req.From
	in.ToCurrency = req.To
	resp, err := s.grpcclient.GetExchangeRateForCurrency(ctx, in)
	if err != nil {
		s.lg.ErrorCtx(ctx, fmt.Sprintf("error getting exchange rate: %v", err))
		errRes.message = "Error fetching exchange rate"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Error fetching exchange rate", http.StatusInternalServerError)
		return
	}
	exchangeRes := new(ExchangeResponseForCurrency)
	mapres, err := s.db.ExchangeForCurrency(ctx, req.From, req.To, req.Amount, resp.Rate, user_id)
	if err != nil {
		if err == storages.ErrExch {
			s.lg.ErrorCtx(ctx, fmt.Sprintf("error : %v", err))
			errRes.message = "Insufficient funds or invalid amount"
			json.NewEncoder(w).Encode(errRes)
			http.Error(w, "Insufficient funds or invalid amount", http.StatusBadRequest)
			return
		} else {
			s.lg.ErrorCtx(ctx, fmt.Sprintf("error exchanging currency: %v", err))
			errRes.message = "Error exchanging currency"
			json.NewEncoder(w).Encode(errRes)
			http.Error(w, "Error exchanging currency", http.StatusInternalServerError)
			return
		}
	}
	exchangeRes.New_balance = mapres
	exchangeRes.Amount = req.Amount
	exchangeRes.Message = "Successfully exchanged currency"
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exchangeRes)
	s.lg.InfoCtx(ctx, fmt.Sprintf("User newbalance %v ", exchangeRes.New_balance))
}
