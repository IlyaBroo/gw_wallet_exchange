package handlers

import (
	"encoding/json"
	"fmt"
	"gw-currency-wallet/internal/middleware"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your_secret_key")

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// @Summary Авторизация пользователя
// @Description Позволяет пользователю войти в систему и получить JWT-токен для дальнейшей аутентификации.
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Данные для авторизации"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse "Error decoding LoginResponse"
// @Failure 401 {object} ErrorResponse "User not found"
// @Failure 401 {object} ErrorResponse "Invalid password"
// @Failure 500 {object} ErrorResponse "Could not generate token"
// @Router /login [post]

func (s *ServerWallet) Autherisation(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	errRes := new(ErrorResponse)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Error decoding: %v", err))
		errRes.message = "Error decoding LoginRespons"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUser(req.Username, r.Context())
	if err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("error getting user: %v", err))
		errRes.message = "User not found"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Invalid credentials"))
		errRes.message = "Invalid password"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.Id, user.Username)
	if err != nil {
		s.lg.ErrorCtx(r.Context(), fmt.Sprintf("Error generating token: %v", err))
		errRes.message = "Could not generate token"
		json.NewEncoder(w).Encode(errRes)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
	s.lg.InfoCtx(r.Context(), fmt.Sprintf("User %s logged in successfully", user.Username))
}

func generateToken(user_id int, username string) (string, error) {
	expirationTime := time.Now().Add(50 * time.Minute)
	claims := new(middleware.Claims)
	claims.Id = user_id
	claims.Username = username
	claims.StandardClaims = jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
