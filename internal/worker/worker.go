package worker

import (
	"context"

	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/annakonkova23/gophermart/internal/service"
	"github.com/sirupsen/logrus"
)

type WorkerAccumulationSystem struct {
	accumSystem *service.AccumulationSystem
}

func NewWorker(accumSystem *service.AccumulationSystem) *WorkerAccumulationSystem {
	return &WorkerAccumulationSystem{accumSystem: accumSystem}
}

func (was *WorkerAccumulationSystem) StartWorkers(ctx context.Context, workers int) {

	for i := 0; i < workers; i++ {
		go was.resultWorker(ctx, i+1, was.accumSystem.GetInputChan())
	}
}

func (was *WorkerAccumulationSystem) resultWorker(ctx context.Context, id int, in <-chan *model.StatusOrder) {
	logrus.Infof("worker обработки полученного результата статуса заказа %d запущен", id)
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("worker обработки полученного результата статуса заказа %d остановлен (ctx done)", id)
			return

		case res, ok := <-in:
			if !ok {
				logrus.Infof("worker обработки полученного результата статуса заказа %d: канал закрыт", id)
				return
			}

			logrus.Infof("Получен результат для сохранения статуса, заказ %s, статус: %s", res.Number, res.Status)
			err := was.accumSystem.CompareAndSaveStatusOrder(ctx, res)
			if err != nil {
				logrus.Errorln("Ошибка сохранения статуса заказа", err)
			}
		}
	}
}
