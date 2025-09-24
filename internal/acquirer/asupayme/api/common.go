package api

import (
	"encoding/hex"
	"strconv"
	"testStand/internal/acquirer/helper"
)

const (
	Pending    = "new"
	Reconciled = "executed"
	Decline    = "cancelled"
)

type Response struct {
	OK             bool   `json:"ok"`
	Status         string `json:"status"`
	Id             string `json:"id"`
	Url            string `json:"url"`
	P2PDestination string `json:"p2p_destination"`
	P2PBank        string `json:"p2p_bank"`
	P2PName        string `json:"p2p_name"`
	Error          string `json:"error"`
	Amount         int64  `json:"amount"`
}

// -----------
type Request struct { // ?
	Merchant   string    `json:"merchant"`
	WithdrawID string    `json:"withdraw_id"`
	CardData   *CardData `json:"card_data,omitempty"`
	Amount     int64     `json:"amount"`
	Signature  string    `json:"signature"`
}

type CardData struct {
	OwnerName    string `json:"owner_name"`
	CardNumber   string `json:"card_number"`
	ExpiredMonth string `json:"expired_month"`
	ExpiredYear  string `json:"expired_year"`
}

// -------------

type StatusRequest struct {
	Id      string `json:"id"`
	MerchId string `json:"merch_id"`
	UserRef string `json:"user_ref,omitempty"`
}

type Callback struct {
	Id               string `json:"id"`
	UserRef          string `json:"user_ref"`
	Status           string `json:"status"`
	Description      string `json:"description"`
	TimestampUpdated string `json:"timestamp_updated"`
	Amount           string `json:"amount"`
	Sign             string `json:"sign"`
}

// createSign
func createSign(req *Request, secretKey string) (string, error) { // ?
	amountStr := strconv.FormatInt(req.Amount, 10)
	stringForHash := req.Merchant + req.CardData.CardNumber + amountStr + secretKey
	hashed := helper.GenerateSHA256Hash(stringForHash)
	signature := hex.EncodeToString(hashed)
	return signature, nil
}
