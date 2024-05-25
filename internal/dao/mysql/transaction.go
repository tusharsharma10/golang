package mysql

import (
	"restapi/db"
)

type TransactionDao struct {
	*database
}

func NewTransactionDao(dB *db.DB) *TransactionDao {
	if dB == nil {
		panic("DB cannot be null")
	}

	return &TransactionDao{
		database: &database{db: dB},
	}
}
