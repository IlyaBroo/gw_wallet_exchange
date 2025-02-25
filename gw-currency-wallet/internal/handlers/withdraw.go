package handlers

import (
	"encoding/json"
	"fmt"
	"gw-currency-wallet/internal/middleware"
	"gw-currency-wallet/internal/storages"
	"net/http"

	"github.com/shopspring/decimal"
)

type WithdrawRequest struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}
type WithdrawResponse struct {
	Message    string           `json:"message"`
	NewBalance storages.Balance `json:"new_balance"`
}

// @Summary Вывод средств
// @Description Позволяет пользователю вывести средства со своего счета. Проверяется наличие достаточного количества средств и корректность суммы.
// @Tags wallet
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT_TOKEN"
// @Param withdraw body WithdrawRequest true "Данные для вывода средств"
// @Success 200 {object} WithdrawResponse
// @Failure 400 {object} ErrorResponse "Error decoding WithdrawResponse"
// @Failure 400 {object} ErrorResponse "Insufficient funds or invalid amount"
// @Failure 400 {object} ErrorResponse "Amount cannot have more than two decimal places"
// @Failure 500 {object} ErrorResponse "Error depositing funds"
// @Failure 500 {object} ErrorResponse "Error getting balance from db"
// @Router /withdraw [post]
func (s *ServerWallet) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawRequest
	res := new(WithdrawResponse)
	errRes := new(ErrorResponse)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Error decoding: %v", err))
		errRes.message = "Error decoding WithdrawResponse"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Error decoding WithdrawResponse", http.StatusBadRequest)
		return
	}

	if req.Amount.Exponent() < -2 {
		s.lg.ErrorCtx(r.Context(), "Amount cannot have more than two decimal places")
		errRes.message = "Amount cannot have more than two decimal places"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Amount cannot have more than two decimal places", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user_id := ctx.Value(middleware.User_id).(int)

	err := s.db.Withdraw(user_id, req.Amount, req.Currency, r.Context())
	if err != nil {
		if err == storages.ErrWithdraw {
			s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Error insufficient funds or invalid amount: %v", err))
			errRes.message = "Insufficient funds or invalid amount"
			json.NewEncoder(w).Encode(errRes)
			http.Error(w, "Insufficient funds or invalid amount", http.StatusBadRequest)
			return
		} else {
			s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Error withdrawing funds: %v", err))
			errRes.message = "Error depositing funds"
			json.NewEncoder(w).Encode(errRes)
			http.Error(w, "Error depositing funds", http.StatusBadRequest)
			return
		}
	}
	res.NewBalance, err = s.db.GetBalance(user_id, r.Context())
	if err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("error getting balance: %v", err))
		errRes.message = "Error getting balance from db"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Error getting balance", http.StatusInternalServerError)
		return
	}
	res.Message = "Withdrawal successful"
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	s.lg.InfoCtx(r.Context(), fmt.Sprintf("User %d withdrew successfully", user_id))
}
