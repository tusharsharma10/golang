package transaction

import models "restapi/internal/model"

func (as *Service) Info() ([]models.Transaction, error) {
	return as.transactionDao.FetchAllActiveActions()
}
