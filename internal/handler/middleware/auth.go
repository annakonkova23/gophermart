package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

type SessionsByToken struct {
	session sync.Map
}

type UserCtxKey struct{}

func (sbt *SessionsByToken) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil || c.Value == "" {
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}

		user, ok := sbt.session.Load(c.Value)
		if !ok {
			fmt.Println("not author")
			http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		} else {
			fmt.Println(user)
		}

		// кладём юзера в контекст
		ctx := context.WithValue(r.Context(), UserCtxKey{}, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (sbt *SessionsByToken) AddSession(token, user string) {
	sbt.session.Store(token, user)
}

// helper, чтобы доставать пользователя в хендлерах
func CurrentUser(r *http.Request) string {
	if v := r.Context().Value(UserCtxKey{}); v != nil {
		if u, ok := v.(string); ok {
			return u
		}
	}
	return ""
}
