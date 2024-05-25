package transaction

import (
	transaction "restapi/internal/service/transaction"

	"restapi/db"
)

type Controller struct {
	actionService *transaction.Service
}

func NewTransactionController(dB *db.DB,
	masterDB *db.DB) *Controller {

	if dB == nil {
		panic("db cannot be null")
	}

	return &Controller{
		actionService: transaction.NewTransactionService(dB, masterDB),
	}
}
