package service

import (
	"context"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/annakonkova23/gophermart/internal/config"
	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/annakonkova23/gophermart/internal/service/accrual"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

var loginRegexp = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,32}$`)

type AccumulationSystem struct {
	users           sync.Map
	database        *sqlx.DB
	client          *accrual.AccrualClient
	newOrders       chan string
	currentOrder    sync.Map
	balance         sync.Map
	timeoutInterval time.Duration
}

func NewAccumulationSystem(ctx context.Context, database *sqlx.DB, cfg *config.Config) *AccumulationSystem {
	as := &AccumulationSystem{
		users:           sync.Map{},
		database:        database,
		client:          accrual.NewAccrualClient(ctx, cfg.AccrualAddress, 10, 10*time.Second, 1000),
		newOrders:       make(chan string, 1000),
		timeoutInterval: time.Second * 10,
	}
	/*функции инициализации*/
	go as.monitorNewOrder(ctx)
	go as.monitorResults(ctx)
	return as
}

func (as *AccumulationSystem) NewUser(user *model.User) error {
	if user.Empty() {
		return model.ErrorIncorrectRequest
	}

	err := as.LoginValid(user.Login)
	if err != nil {
		return model.ErrorIncorrectRequest
	}

	err = user.HashPassword()
	if err != nil {
		return err
	}

	err = as.сreateUserDB(user)
	if err != nil {
		return err
	}
	return nil
}

func (as *AccumulationSystem) AuthUser(user *model.User) error {
	if user.Empty() {
		return model.ErrorIncorrectRequest
	}

	userCurrent, err := as.getUserByLoginDB(user.Login)
	if err != nil {
		return err
	}

	if userCurrent.Equals(user) {
		return nil
	} else {
		return model.ErrorNotEqual
	}

}

func (as *AccumulationSystem) NewOrder(user, number string) error {

	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return model.ErrorNotAuthorization
	}

	if !as.LuhnValid(number) {
		return model.ErrorNotValidNumber
	}
	err = as.createOrderDB(user, number)
	if err != nil {
		return err
	}
	as.AddCurrentOrder(&model.Order{Number: number, Status: model.StatusNew, UploadedAt: time.Now(), User: user})
	return nil
}

func (as *AccumulationSystem) GetOrders(user string) ([]*model.Order, error) {

	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return nil, model.ErrorNotAuthorization
	}

	orders, err := as.getOrderDB(user)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (as *AccumulationSystem) GetBalance(user string) (*model.Balance, error) {

	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return nil, model.ErrorNotAuthorization
	}

	balance, err := as.getBalanceDB(user)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (as *AccumulationSystem) WithdrawBonus(user string, req *model.Withdraw) error {
	logrus.Infof("Запрос на списание пользователь %s, заказ %s, сумма %s", user, req.OrderNumber, req.Sum.String())
	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return model.ErrorNotAuthorization
	}

	if !as.LuhnValid(req.OrderNumber) {
		return model.ErrorNotValidNumber
	}
	err = as.withdrawBalance(context.Background(), user, req)
	if err != nil {
		return err
	}
	return nil
}

//func

func (as *AccumulationSystem) GetWithdrawals(user string) ([]*model.Withdraw, error) {

	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return nil, model.ErrorNotAuthorization
	}

	withdrawals, err := as.getWithdrawalsDB(user)
	if err != nil {
		return nil, err
	}
	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].ProcessedAt.After(withdrawals[j].ProcessedAt)
	})

	return withdrawals, nil
}

func (as *AccumulationSystem) monitorNewOrder(ctx context.Context) {
	logrus.Infoln("Старт мониторинга новых заказов")
	for {
		select {
		case <-ctx.Done():
			logrus.Infoln("Прерывание мониторинга новых заказов")
		case order := <-as.newOrders:
			as.client.AddOrder(ctx, order)
		}
	}
}

func (as *AccumulationSystem) monitorResults(ctx context.Context) {
	logrus.Infoln("Старт мониторинга результатов")
	for res := range as.client.Results() {
		logrus.Infof("Получен результат для сохранения статуса, заказ %s, статус: %s", res.Number, res.Status)
		err := as.compareAndSaveStatusOrder(ctx, res)
		if err != nil {
			logrus.Errorln("Ошибка сохранения статуса заказа", err)
		}
	}
}

func (as *AccumulationSystem) compareAndSaveStatusOrder(ctx context.Context, orderStatus *model.StatusOrder) error {
	logrus.Infoln("Обновление заказа:", orderStatus.Number)
	orderCurrent, ok := as.GetCurrentOrder(orderStatus.Number)
	if !ok {
		logrus.Errorf("Не найден заказ %s в системе", orderStatus.Number)
	}
	dZero := decimal.NewFromInt(0)
	orderCurrent.Mx.Lock()
	defer orderCurrent.Mx.Unlock()
	if model.StatusesXmapCalc[orderStatus.Status] != orderCurrent.Status {
		orderCurrent.Status = model.StatusesXmapCalc[orderStatus.Status]
		orderCurrent.Accrual = &orderStatus.Accrual
		if model.StatusesCalcFinish[orderStatus.Status] {
			as.saveEndStatusOrderDB(ctx, orderCurrent)
			if orderStatus.Accrual.GreaterThan(dZero) {
				logrus.Infof("Получен баланс %s для заказа %s:", orderStatus.Accrual.String(), orderStatus.Number)
				err := as.addBalance(orderCurrent.User, *orderCurrent.Accrual)
				if err != nil {
					return err
				}
			}
			as.DeleteProcessedOrder(orderCurrent.Number)
		} else {
			err := as.saveStatusOrderDB(orderCurrent)
			if err != nil {
				logrus.Errorf("Ошибка сохранения статуса заказа %s в базу:%s", orderCurrent.Number, err.Error())
				return err
			}
			as.UpdateCurrentOrder(orderCurrent.Number, orderCurrent)
		}

	} else {
		logrus.Infof("Статус заказа %s не обновился", orderStatus.Number)
	}
	return nil
}

func (as *AccumulationSystem) addBalance(user string, accrual decimal.Decimal) error {
	balanceUser := &model.Balance{Balance: accrual}
	balance, ok := as.balance.LoadOrStore(user, balanceUser)
	if ok {
		balanceUser = balance.(*model.Balance)
	}
	balanceUser.Mx.Lock()
	defer balanceUser.Mx.Unlock()
	balanceUser.Balance.Add(accrual)
	err := as.saveBalanceDB(user, balanceUser)
	if err != nil {
		return err
	}
	as.balance.Store(user, balanceUser)
	return nil
}

func (as *AccumulationSystem) withdrawBalance(ctx context.Context, user string, req *model.Withdraw) error {
	balance, ok := as.balance.Load(user)
	if !ok {
		logrus.Errorf("У пользователя %s нет баланса", user)
		return model.ErrorInsufficientFunds
	}

	balanceUser := balance.(*model.Balance)

	balanceUser.Mx.Lock()
	defer balanceUser.Mx.Unlock()

	if balanceUser.Balance.GreaterThanOrEqual(req.Sum) {
		balanceUser.Balance.Sub(req.Sum)
		balanceUser.Withdrawn.Add(req.Sum)
		logrus.Infof("У пользователя %s списываем баллы %s", user, req.Sum.String())
		req.ProcessedAt = time.Now()
		err := as.saveBalanceAndWithdrawDB(ctx, user, balanceUser, req)
		if err != nil {
			return err
		}
		as.balance.Store(user, balanceUser)
	} else {
		logrus.Errorf("У пользователя %s недостаточно средств", user)
		return model.ErrorInsufficientFunds
	}
	return nil
}
