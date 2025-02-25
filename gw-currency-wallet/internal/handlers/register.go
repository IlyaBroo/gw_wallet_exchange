package handlers

import (
	"encoding/json"
	"fmt"
	"gw-currency-wallet/internal/storages"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// @Summary Регистрация пользователя
// @Description Позволяет зарегистрировать нового пользователя. Проверяется уникальность имени пользователя и адреса электронной почты. Пароль должен быть зашифрован перед сохранением в базе данных.
// @Tags auth
// @Accept json
// @Produce json
// @Param register body storages.RegisterRequest true "Данные для регистрации пользователя"
// @Success 201 {object} map[string]string "User  registered successfully"
// @Failure 400 {object} ErrorResponse "Username or email already exists"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 500 {object} ErrorResponse "Could not hash password"
// @Failure 500 {object} ErrorResponse "Could not create user"
// @Router /register [post]
func (s *ServerWallet) RegisterUser(w http.ResponseWriter, r *http.Request) {
	errRes := new(ErrorResponse)
	var req storages.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Error decoding: %v", err))
		errRes.message = "Invalid input"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	exists, err := s.db.CheckUser(req.Username, req.Email, r.Context())
	if err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("error checking: %v", err))
		errRes.message = "Internal server error"
		json.NewEncoder(w).Encode(errRes)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Username or email already exists"))
		errRes.message = "Username or email already exists"
		json.NewEncoder(w).Encode(errRes)
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Username or email already exists", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Error hashing password: %v", err))
		errRes.message = "Could not hash password"
		json.NewEncoder(w).Encode(errRes)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	req.Password = string(hashedPassword)

	if err := s.db.AddUser(req, r.Context()); err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Error adding user: %v", err))
		errRes.message = "Could not create user"
		json.NewEncoder(w).Encode(errRes)
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": fmt.Sprintf("User %s registered successfully", req.Username)})
	s.lg.InfoCtx(r.Context(), fmt.Sprintf("User %s registered successfully", req.Username))
}
