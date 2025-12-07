package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type AccrualClient struct {
	client        *resty.Client
	jobs          chan string
	checkInterval time.Duration
	results       chan *model.StatusOrder
}

func NewAccrualClient(ctx context.Context, baseURL string, workers int, interval time.Duration, jobs int) *AccrualClient {
	ac := &AccrualClient{
		client: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(5 * time.Second),
		jobs:          make(chan string, jobs),
		results:       make(chan *model.StatusOrder, jobs),
		checkInterval: interval,
	}

	for i := 0; i < workers; i++ {
		go ac.workerOrders(ctx, i+1)
	}

	return ac
}

func (ac *AccrualClient) AddOrder(ctx context.Context, order string) error {
	select {
	case ac.jobs <- order:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (ac *AccrualClient) workerOrders(ctx context.Context, workerID int) {
	logrus.Infof("worker %d started", workerID)
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("worker %d stopped (ctx done)", workerID)
			return

		case order := <-ac.jobs:
			ac.pollOrder(ctx, workerID, order)
		}
	}
}

// pollOrder — цикл опроса одного заказа до конечного статуса или отмены контекста.
func (ac *AccrualClient) pollOrder(ctx context.Context, workerID int, order string) {
	firstStatus := false
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("worker %d: остановился на заказе %s (ctx done)", workerID, order)
			return
		default:
		}

		status, err := ac.getStatus(ctx, order)
		if err != nil {
			logrus.Infof("worker %d: заказ %s: ошибка получения статуса заказа: %v", workerID, order, err)
		} else {
			js, _ := json.Marshal(status)
			logrus.Infof("worker %d: : заказ %s: статус = %s", workerID, order, string(js))

			if !firstStatus && !model.StatusesCalcFinish[status.Status] {
				logrus.Infof("worker %d: заказ %s: получен первый статус: %s", workerID, order, status.Status)
				ac.results <- status
				firstStatus = true
			} else if model.StatusesCalcFinish[status.Status] {
				logrus.Infof("worker %d: заказ %s получен финальный статус: %s", workerID, order, status.Status)
				ac.results <- status
				return
			}
		}

		select {
		case <-ctx.Done():
			logrus.Infof("worker %d: stop polling order %s while waiting", workerID, order)
			return
		case <-time.After(ac.checkInterval):
		}
	}
}

func (ac *AccrualClient) getStatus(ctx context.Context, order string) (*model.StatusOrder, error) {
	var resp *model.StatusOrder

	r, err := ac.client.R().
		SetContext(ctx).
		SetResult(&resp).
		SetPathParams(map[string]string{
			"number": order,
		}).
		Get("GET /api/orders/{number}")

	if err != nil {
		return nil, fmt.Errorf("Сетевая ошибка: %w", err)
	}

	switch r.StatusCode() {
	case 200:
		logrus.Infof("Статус заказа %s получен", order)
		return resp, nil

	case 204:
		logrus.Infof("Заказа %s нет в системе", order)
		return nil, fmt.Errorf("Заказа %s нет в системе", order)

	case 429:
		logrus.Infoln("Превышено количество запросов")
		return nil, fmt.Errorf("rate limited: status 429")

	default:
		return nil, fmt.Errorf("Ошибочный статус %d", r.StatusCode())
	}
}

func (ac *AccrualClient) Results() <-chan *model.StatusOrder {
	return ac.results
}
