package mysql

import (
	"context"
	"os"

	"restapi/db"
	"restapi/logger"
	helpers "restapi/util"

	"github.com/jmoiron/sqlx"
)

type database struct {
	db       *db.DB
	masterDB *db.DB
}

// nitin: let's move this to goofy ?
func (dB *database) Transaction(caller func(tx *sqlx.Tx) (interface{}, error)) (interface{}, error) {
	transaction, err := dB.db.Dbx.Beginx()

	defer func() {
		if err := recover(); err != nil {
			logger.Debug(context.Background(), "panic in transaction", logger.Z{
				"error": err,
			})

			if transaction != nil {
				// nolint:errcheck
				transaction.Rollback()
			}
		}
	}()

	if err != nil {
		return nil, err
	}

	result, err := caller(transaction)
	if err != nil {
		errCheck := transaction.Rollback()
		if errCheck != nil {
			return nil, errCheck
		}

		return nil, err
	}

	return result, transaction.Commit()
}

func decryptRedeemCode(input string) (string, error) {
	output, err := helpers.DecryptWithRandomIV([]byte(os.Getenv("ENCRYPTION_SECRET_KEY")), input)

	return string(output), err
}

func encryptRedeemCode(input string) string {
	return helpers.EncryptWithRandomIV(
		[]byte(os.Getenv("ENCRYPTION_SECRET_KEY")),
		[]byte(input))
}
