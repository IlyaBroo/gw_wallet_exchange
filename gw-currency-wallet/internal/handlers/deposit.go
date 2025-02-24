package handlers

import (
	"encoding/json"
	"fmt"
	"gw-currency-wallet/internal/middleware"
	"gw-currency-wallet/internal/storages"
	"net/http"
)

type DepositRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}
type DepositResponse struct {
	Message    string           `json:"message"`
	NewBalance storages.Balance `json:"new_balance"`
}

// @Summary Пополнение счета
// @Description Позволяет пользователю пополнить свой счет. Проверяется корректность суммы и валюты. Обновляется баланс пользователя в базе данных.
// @Tags wallet
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT_TOKEN"
// @Param deposit body DepositRequest true "Данные для пополнения счета"
// @Success 200 {object} DepositResponse
// @Failure 400 {object} ErrorResponse "Invalid amount or currency"
// @Failure 500 {object} ErrorResponse "Error depositing funds or getting balance"
// @Failure 500 {object} ErrorResponse "Error getting balance from db"
// @Router /deposit [post]
func (s *ServerWallet) Deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest
	res := new(DepositResponse)
	errRes := new(ErrorResponse)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Error decoding: %v", err))
		errRes.message = "Invalid amount or currency"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Invalid amount or currency", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user_id := ctx.Value(middleware.User_id).(int)

	err := s.db.Deposit(user_id, req.Amount, req.Currency, r.Context())
	if err != nil {
		if err == storages.ErrWalletid {
			s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Invalid amount or currency: %v", err))
			errRes.message = "Invalid amount or currency"
			json.NewEncoder(w).Encode(errRes)
			http.Error(w, "Invalid amount or currency", http.StatusBadRequest)
			return

		} else {

			s.lg.ErrorCtx(r.Context(), fmt.Sprintf("error depositing funds: %v", err))
			errRes.message = "Error depositing funds"
			json.NewEncoder(w).Encode(errRes)
			http.Error(w, "Error depositing funds", http.StatusInternalServerError)
			return
		}
	}
	res.NewBalance, err = s.db.GetBalance(user_id, r.Context())
	if err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("error getting balance: %v", err))
		errRes.message = "Error getting balance  from db"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Error getting balance", http.StatusInternalServerError)
		return
	}
	res.Message = "Account topped up successfully"
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	s.lg.InfoCtx(r.Context(), fmt.Sprintf("User %d deposited successfully", user_id))
}
