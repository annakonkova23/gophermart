package handler

import (
	"encoding/json"
	"net/http"

	"github.com/annakonkova23/gophermart/internal/handler/middleware"
	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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
	_, err = w.Write(bodyResult)
	if err != nil {
		logrus.Errorln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
	_, err = w.Write(bodyResult)
	if err != nil {
		logrus.Errorln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
	_, err = w.Write(bodyResult)
	if err != nil {
		logrus.Errorln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
