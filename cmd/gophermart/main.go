package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/annakonkova23/gophermart/internal/config"
	"github.com/annakonkova23/gophermart/internal/config/db"
	"github.com/annakonkova23/gophermart/internal/handler"
	"github.com/annakonkova23/gophermart/internal/service"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.GetConfig()
	database, err := db.NewConnect(cfg.DBUri)
	if err != nil {
		log.Fatal("Ошибка при подключении к базе данных:", err)
	}
	logrus.Println("Подключение к базе данных успешно")
	if err := db.RunMigrations(cfg.DBUri); err != nil {
		log.Fatal("Ошибка при установке миграций:", err)
	}
	accumSystem, err := service.NewAccumulationSystem(ctx, database, cfg)
	if err != nil {
		log.Fatal(err)
	}
	server := handler.NewServer(cfg.Host, accumSystem)

	go func() {
		logrus.Printf("Сервер запущен на: %s", cfg.Host)
		if err := server.Start(ctx); err != nil {
			logrus.Error(err)
		}
	}()

	<-ctx.Done()
	logrus.Println("Сервер остановлен")

}
