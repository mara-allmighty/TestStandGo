package api

import (
	"encoding/hex"
	"testStand/internal/acquirer/helper"
)

const (
	Pending    = "new"
	Reconciled = "executed"
	Decline    = "cancelled"
)

type AsupaymeResponse struct { // #
	Status   string `json:"status"`
	Id       string `json:"uuid"`     // #
	Messages string `json:"messages"` // #
	Code     string `json:"code"`
}

type WithdrawRequestBody struct {
	Merchant   string    `json:"merchant"`
	WithdrawID string    `json:"withdraw_id"`
	CardData   *CardData `json:"card_data,omitempty"`
	Amount     string    `json:"amount"`
	Signature  string    `json:"signature"`
}

type CardData struct {
	OwnerName    string `json:"owner_name"`
	CardNumber   string `json:"card_number"`
	ExpiredMonth string `json:"expired_month"`
	ExpiredYear  string `json:"expired_year"`
}

// createSign
func createSign(req *WithdrawRequestBody, secretKey string) string {
	hashString := req.Merchant + req.CardData.CardNumber + req.Amount + secretKey // в доках такого нет?
	hashedString := helper.GenerateSHA256Hash(hashString)
	signature := hex.EncodeToString(hashedString)
	return signature
}
