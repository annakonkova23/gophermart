package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/annakonkova23/gophermart/internal/handler/middleware"
	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *Server) registerUser(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Content-Type должен быть application/json", http.StatusBadRequest)
		return
	}

	var user *model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Невалидный JSON", http.StatusBadRequest)
		return
	}

	err := s.accumulationSystem.NewUser(user)
	if err != nil {
		if errors.Is(err, model.ErrorIncorrectRequest) {
			http.Error(w, "Некорректный формат логин/пароль", http.StatusBadRequest)
			return
		}
		if errors.Is(err, model.ErrorConflict) {
			http.Error(w, "Пользователь с таким логином уже существует", http.StatusConflict)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Server) authUser(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Content-Type должен быть application/json", http.StatusBadRequest)
		return
	}

	var user *model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Некорретный JSON", http.StatusBadRequest)
		return
	}

	err := s.accumulationSystem.AuthUser(user)
	if err != nil {
		if errors.Is(err, model.ErrorIncorrectRequest) {
			http.Error(w, "Некорректный формат логин/пароль", http.StatusBadRequest)
			return
		}
		if errors.Is(err, model.ErrorNotEqual) {
			http.Error(w, "Не совпадает пара логин/пароль", http.StatusUnauthorized)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	sessionToken, err := generateSessionToken()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	s.session.AddSession(sessionToken, user.Login)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})

	w.WriteHeader(http.StatusOK)
}

func (s *Server) newOrder(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		http.Error(w, "Content-Type должен быть text/plain", http.StatusBadRequest)
		return
	}

	user := middleware.CurrentUser(r)

	if user == "" {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Некорретное считывание тело запроса", http.StatusBadRequest)
		return
	}

	text := string(body)

	err = s.accumulationSystem.NewOrder(user, text)
	if err != nil {
		logrus.Error(err)
		if errors.Is(err, model.ErrorDoubleOperation) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, model.ErrorNotAuthorization) {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		if errors.Is(err, model.ErrorNotValidNumber) {
			http.Error(w, "Некорректный формат номера", http.StatusUnprocessableEntity)
			return
		}

		if errors.Is(err, model.ErrorConflict) {
			http.Error(w, "Этот заказ зарегистрирован уже у другого пользователя", http.StatusConflict)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) getOrders(w http.ResponseWriter, r *http.Request) {

	user := middleware.CurrentUser(r)

	if user == "" {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	orders, err := s.accumulationSystem.GetOrders(user)
	if err != nil {
		logrus.Errorln(err)
		if errors.Is(err, model.ErrorNotContent) {
			http.Error(w, "У пользователя нет заказов", http.StatusNoContent)
			return
		}

		if errors.Is(err, model.ErrorNotAuthorization) {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	bodyResult, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bodyResult)
}

func (s *Server) getBalance(w http.ResponseWriter, r *http.Request) {

	user := middleware.CurrentUser(r)

	if user == "" {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	balance, err := s.accumulationSystem.GetBalance(user)
	if err != nil {
		logrus.Error(err)
		if errors.Is(err, model.ErrorNotAuthorization) {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	bodyResult, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bodyResult)
}

func (s *Server) withdrawBonus(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Content-Type должен быть application/json", http.StatusBadRequest)
		return
	}

	user := middleware.CurrentUser(r)

	if user == "" {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	var requestWithdraw *model.Withdraw
	if err := json.NewDecoder(r.Body).Decode(&requestWithdraw); err != nil {
		http.Error(w, "Невалидный JSON", http.StatusBadRequest)
		return
	}

	err := s.accumulationSystem.WithdrawBonus(user, requestWithdraw)
	if err != nil {

		if errors.Is(err, model.ErrorNotAuthorization) {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		if errors.Is(err, model.ErrorInsufficientFunds) {
			http.Error(w, "На счету недостаточно средств", http.StatusPaymentRequired)
			return
		}

		if errors.Is(err, model.ErrorNotValidNumber) {
			http.Error(w, "Некорректный формат номера", http.StatusUnprocessableEntity)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getWithdrawals(w http.ResponseWriter, r *http.Request) {

	user := middleware.CurrentUser(r)

	if user == "" {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	withdrawals, err := s.accumulationSystem.GetWithdrawals(user)
	if err != nil {
		logrus.Errorln(err)
		if errors.Is(err, model.ErrorNotAuthorization) {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	bodyResult, err := json.Marshal(withdrawals)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bodyResult)
}
