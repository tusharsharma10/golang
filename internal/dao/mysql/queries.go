package mysql

import (
	"database/sql"
	"restapi/helpers"

	"github.com/jmoiron/sqlx"

	model "restapi/internal/model"
)

func (ad *TransactionDao) Create(tx *sqlx.Tx, action *model.Transaction) (int64, error) {
	query := helpers.CreateInsertQuery("transactions", []string{
		"code",
		"companyId",
		"jobProfileId",
	})

	var (
		res sql.Result
		err error
	)

	if tx != nil {
		res, err = tx.NamedExec(query, action)
	} else {
		res, err = ad.db.Dbx.NamedExec(query, action)
	}

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// func (ad *TransactionDao) Update(tx *sqlx.Tx, action *model.Transaction) (int64, error) {
// 	query := `UPDATE
// 		actions
// 		SET
// 		Info = ?,
// 		AdditionalInfo = ?
// 		WHERE Code = ?
// 		`

// 	var (
// 		res sql.Result
// 		err error
// 	)

// 	if tx != nil {
// 		res, err = tx.Exec(query,
// 			action.Info, action.AdditionalInfo, action.Code)
// 	} else {
// 		res, err = ad.db.Dbx.Exec(query,
// 			action.Info, action.AdditionalInfo, action.Code)
// 	}

// 	if err != nil {
// 		return 0, err
// 	}

// 	return res.RowsAffected()
// }

// func (ad *TransactionDao) Delete(code string) error {
// 	query := `DELETE FROM actions WHERE Code = ?`

// 	_, err := ad.db.Dbx.Exec(query, code)

// 	return err
// }

// func (ad *TransactionDao) GetActionByCode(code string) (*model.Transaction, error) {
// 	var action model.Transaction

// 	query := `SELECT
// 			Id,
// 			Code,
// 			Info,
// 			Active,
// 			CurrencyExpression,
// 			AdditionalInfo,
// 			Type
// 		FROM
// 			actions
// 		WHERE
// 			Active = 1 AND
// 			Code = ?
// 	`

// 	err := ad.db.Dbx.Get(&action, query, code)
// 	if err != nil {
// 		return &action, err
// 	}

// 	action.UnmarshalInfo()
// 	action.UnmarshalAdditionalInfo()

// 	return &action, nil
// }

// func (ad *TransactionDao) IsValidUser(userID string) bool {
// 	var blockedUser null.Int

// 	query := `
// 		SELECT
// 			UserId
// 		FROM blacklisted_users
// 		WHERE UserId = ?
// 	`

// 	err := ad.db.Dbx.Get(&blockedUser, query, userID)

// 	return err != nil
// }

// func (ad *TransactionDao) FetchActionCurrencyRules(actionID int) ([]model.TransactionCurrency, error) {
// 	var actionCurrency []model.TransactionCurrency

// 	query := `
// 		Select
// 			Id,
// 			ActionId,
// 			Type,
// 			Info
// 		FROM action_currency_rules
// 		WHERE
// 			ActionId = ?
// 	`

// 	err := ad.db.Dbx.Select(&actionCurrency, query, actionID)

// 	return actionCurrency, err
// }

// func (ad *TransactionDao) CountUserActionsInInterval(actionID int, userID string, days int) (int, error) {
// 	var count null.Int

// 	query := `
// 		SELECT COUNT(*)
// 		FROM transactions
// 		JOIN
// 			actions ON
// 			actions.Id = transactions.ActionId
// 		WHERE actions.Id = ? AND
// 		transactions.UserId = ?
// 		AND transactions.Created > NOW() - INTERVAL ? DAY
// 	`

// 	err := ad.db.Dbx.Get(&count, query, actionID, userID, days)

// 	return int(count.ValueOrZero()), err
// }

// func (ad *TransactionDao) AddActionCurrencyRules(currencyRules model.TransactionCurrency) (int64, error) {
// 	query := helpers.CreateInsertQuery("action_currency_rules", []string{
// 		"ActionId",
// 		"Type",
// 		"Info",
// 		"ModifiedBy",
// 	})

// 	res, err := ad.db.Dbx.NamedExec(query, currencyRules)
// 	if err != nil {
// 		return 0, err
// 	}

// 	return res.LastInsertId()
// }

func (ad *TransactionDao) FetchAllActiveActions() ([]model.Transaction, error) {
	var actions []model.Transaction

	query := `
		SELECT
			code,
			companyId,
			jobProfileId
		FROM transactions
	`

	err := ad.db.Dbx.Select(&actions, query)

	return actions, err
}
