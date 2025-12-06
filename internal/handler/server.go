package handler

import (
	"context"
	"net/http"

	"github.com/annakonkova23/gophermart/internal/handler/middleware"
	"github.com/annakonkova23/gophermart/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Server struct {
	url                string
	mux                *chi.Mux
	srv                *http.Server
	session            *middleware.SessionsByToken
	accumulationSystem *service.AccumulationSystem
}

func NewServer(url string, accSystem *service.AccumulationSystem) *Server {

	mux := chi.NewRouter()

	s := &Server{
		mux:                mux,
		url:                url,
		accumulationSystem: accSystem,
		session:            &middleware.SessionsByToken{},
	}
	s.mux.Use(middleware.CompressMiddleware)
	s.mux.Use(middleware.LoggingMiddleware)
	//s.mux.Use(middleware.AuthMiddleware(key))
	s.mux.Post("/api/user/register", s.registerUser)
	s.mux.Post("/api/user/login", s.authUser)

	s.mux.Group(func(pr chi.Router) {
		pr.Use(s.session.AuthMiddleware)
		pr.Post("/api/user/orders", s.newOrder)
		pr.Get("/api/user/orders", s.getOrders)
		pr.Get("/api/user/balance", s.getBalance)
		pr.Post("/api/user/balance/withdraw", s.withdrawBonus)
		pr.Get("/api/user/withdrawals", s.getWithdrawals)
	})

	s.srv = &http.Server{
		Addr:    url,
		Handler: mux,
	}
	return s
}

func (s *Server) Start(ctx context.Context) error {

	go func() {
		<-ctx.Done()
		if err := s.srv.Shutdown(ctx); err != nil {
			logrus.Error("Ошибка при закрытии сервера:", err)
		}
	}()

	err := s.srv.ListenAndServe()
	return err
}
