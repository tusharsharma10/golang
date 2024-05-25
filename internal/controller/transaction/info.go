package transaction

import (
	"net/http"
	"restapi/helpers"
	models "restapi/internal/model"

	"github.com/gin-gonic/gin"
)

type actionInfo struct {
	Code               string      `json:"Code"`
	InfoJSON           interface{} `json:"InfoJson"`
	AdditionalInfoJSON interface{} `json:"AdditionalInfoJson"`
	CurrencyExpression string      `json:"CurrencyExpression"`
	Type               string      `json:"Type"`
}

func presentActionsInfo(input []models.Transaction) []actionInfo {

	actions := make([]actionInfo, 0)

	return actions
}

func (ac *Controller) Info(c *gin.Context) {
	defer helpers.Recover(c, "all-actions-info")

	result, err := ac.actionService.Info()
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, helpers.NewResponse(presentActionsInfo(result), nil))
}
