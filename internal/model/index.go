package models

import (
	_ "context"
	_ "encoding/json"

	_ "restapi/logger"
)

type Transaction struct {
	TxnId        int32  `db:"txnId"`
	Code         string `db:"code"`
	CompanyId    int32  `db:"companyId"`
	JobprofileId int32  `db:"jobProfileId"`

	//InfoJSON           Info
	//AdditionalInfoJSON AdditionalInfo
}

// func (ac *Action) UnmarshalInfo() {
// 	err := json.Unmarshal([]byte(ac.Info), &ac.InfoJSON)
// 	if err != nil {
// 		logger.Debug(context.Background(), "error in unmarshalling action json info", logger.Z{
// 			"error": err,
// 			"data":  ac,
// 		})
// 	}
// }

// func (ac *Action) UnmarshalAdditionalInfo() {
// 	err := json.Unmarshal([]byte(ac.AdditionalInfo), &ac.AdditionalInfoJSON)
// 	if err != nil {
// 		logger.Debug(context.Background(), "error in unmarshalling action json additional info", logger.Z{
// 			"error": err,
// 			"data":  ac,
// 		})
// 	}
// }

// type Info struct {
// 	Narration         string `json:"Narration"`
// 	ModifiedNarration string `json:"ModifiedNarration"`
// 	EventNarration    string `json:"EventNarration"`
// }

// type AdditionalInfo struct {
// 	MaxCoins     int `json:"MaxCoins"`
// 	ValidityDays int `json:"ValidityDays"`
// }

// type ActionRewardMapping struct {
// 	ID       int `db:"Id"`
// 	ActionID int `db:"ActionId"`
// 	RewardID int `db:"RewardId"`
// }

// type ActionCurrency struct {
// 	ID         int    `db:"Id"`
// 	ActionID   int    `db:"ActionId"`
// 	Type       string `db:"Type"`
// 	Info       string `db:"Info"`
// 	ModifiedBy string `db:"ModifiedBy"`
// }
