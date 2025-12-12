package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annakonkova23/gophermart/internal/model"
)

// Фейковая система накопления пользователей
type fakeAccumulationSystem struct {
	newUserErr  error
	called      bool
	lastNewUser *model.User
}

func (f *fakeAccumulationSystem) NewUser(u *model.User) error {
	f.called = true
	f.lastNewUser = u
	return f.newUserErr
}

// Фейковый менеджер сессий
type fakeSessionStore struct {
	addCalls []struct {
		token string
		login string
	}
}

func (f *fakeSessionStore) AddSession(token, login string) {
	f.addCalls = append(f.addCalls, struct {
		token string
		login string
	}{token: token, login: login})
}

// helper для создания сервера
func newTestServer(accErr error) (*Server, *fakeAccumulationSystem, *fakeSessionStore) {
	acc := &fakeAccumulationSystem{newUserErr: accErr}
	sess := &fakeSessionStore{}

	s := &Server{
		accumulationSystem: acc,
		session:            sess,
	}

	return s, acc, sess
}
func TestRegisterUser_WrongContentType(t *testing.T) {
	s, _, _ := newTestServer(nil)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	s.registerUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Content-Type должен быть application/json") {
		t.Fatalf("expected error message about Content-Type, got %q", rr.Body.String())
	}
}

func TestRegisterUser_InvalidJSON(t *testing.T) {
	s, _, _ := newTestServer(nil)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"login":`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.registerUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Невалидный JSON") {
		t.Fatalf("expected error message about invalid JSON, got %q", rr.Body.String())
	}
}

func TestRegisterUser_IncorrectRequest(t *testing.T) {
	s, acc, _ := newTestServer(model.ErrorIncorrectRequest)

	body := `{"login":"user1","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.registerUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Некорректный формат логин/пароль") {
		t.Fatalf("expected error message about incorrect login/password, got %q", rr.Body.String())
	}
	if !acc.called {
		t.Fatalf("expected NewUser to be called")
	}
}

func TestRegisterUser_Conflict(t *testing.T) {
	s, _, _ := newTestServer(model.ErrorConflict)

	body := `{"login":"user1","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.registerUser(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Пользователь с таким логином уже существует") {
		t.Fatalf("expected conflict message, got %q", rr.Body.String())
	}
}

func TestRegisterUser_InternalErrorOnNewUser(t *testing.T) {
	s, _, _ := newTestServer(errors.New("some internal error"))

	body := `{"login":"user1","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.registerUser(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), http.StatusText(http.StatusInternalServerError)) {
		t.Fatalf("expected generic 500 message, got %q", rr.Body.String())
	}
}

func TestRegisterUser_Success(t *testing.T) {
	s, acc, sess := newTestServer(nil)

	login := "user1"
	body := `{"login":"` + login + `","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.registerUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Проверяем, что пользователь передался в NewUser
	if !acc.called {
		t.Fatalf("expected NewUser to be called")
	}
	if acc.lastNewUser == nil || acc.lastNewUser.Login != login {
		t.Fatalf("expected NewUser to be called with login %q, got %#v", login, acc.lastNewUser)
	}

	// Проверяем, что выставлена кука
	res := rr.Result()
	defer res.Body.Close()

	var sessionCookie *http.Cookie
	for _, c := range res.Cookies() {
		if c.Name == "session_token" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("expected session_token cookie to be set")
	}
	if sessionCookie.Value == "" {
		t.Fatalf("expected session_token cookie to have non-empty value")
	}
	if sessionCookie.Path != "/" {
		t.Fatalf("expected cookie Path '/', got %q", sessionCookie.Path)
	}
	if !sessionCookie.HttpOnly {
		t.Fatalf("expected cookie to be HttpOnly")
	}

	// Проверяем, что AddSession вызвался с тем же токеном и логином
	if len(sess.addCalls) != 1 {
		t.Fatalf("expected AddSession to be called once, got %d", len(sess.addCalls))
	}
	if sess.addCalls[0].login != login {
		t.Fatalf("expected AddSession login %q, got %q", login, sess.addCalls[0].login)
	}
	if sess.addCalls[0].token != sessionCookie.Value {
		t.Fatalf("expected AddSession token to match cookie value")
	}
}
