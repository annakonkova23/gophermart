package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/pkg/errors"
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
