package transaction

import (
	"restapi/internal/dao/mysql"

	"restapi/db"
)

type Service struct {
	transactionDao *mysql.TransactionDao
}

func NewTransactionService(dB *db.DB,
	masterDB *db.DB,
) *Service {
	if dB == nil {
		panic("db cannot be null")
	}

	return &Service{
		transactionDao: mysql.NewTransactionDao(dB),
	}
}
