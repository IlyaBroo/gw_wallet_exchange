package handlers

import (
	"encoding/json"
	"fmt"
	"gw-currency-wallet/internal/middleware"
	"net/http"
)

// @Summary Получение баланса пользователя
// @Description Позволяет пользователю получить информацию о своем балансе по всем валютам.
// @Tags wallet
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT_TOKEN"
// @Success 200 {object} storages.Balance
// @Failure 500 {object} ErrorResponse "Could not get balance"
// @Router /balance [get]
func (s *ServerWallet) GetBalance(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value(middleware.User_id).(int)
	errRes := new(ErrorResponse)
	balance, err := s.db.GetBalance(user_id, r.Context())
	if err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("error getting balance: %v", err))
		errRes.message = "Could not get balance"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Could not get balance", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
	s.lg.InfoCtx(r.Context(), fmt.Sprintf("User %d requested their balance", user_id))
}
