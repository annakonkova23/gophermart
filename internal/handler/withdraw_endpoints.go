package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/annakonkova23/gophermart/internal/handler/middleware"
	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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
