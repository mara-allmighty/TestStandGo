package asupayme

import (
	"context"
	"log"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/asupayme/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/shopspring/decimal"
)

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type GatewayParams struct {
	Transport            Transport        `json:"transport"`
	PercentageDifference *decimal.Decimal `json:"percentage_difference"`
}

type ChannelParams struct {
	ApiKey     string `json:"api_key"`
	SecretKey  string `json:"secret_key"`
	MerchantId string `json:"merchant_id"`
}

type Acquirer struct {
	api                  *api.Client
	dbClient             *repos.Repo
	channelParams        ChannelParams
	callbackUrl          string
	percentageDifference *decimal.Decimal
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams ChannelParams, gatewayParams GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		channelParams:        channelParams,
		api:                  api.NewClient(ctx, channelParams.MerchantId, gatewayParams.Transport.BaseAddress, channelParams.ApiKey, channelParams.SecretKey, gatewayParams.Transport.Timeout),
		dbClient:             db,
		callbackUrl:          callbackUrl,
		percentageDifference: gatewayParams.PercentageDifference,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	log.Println("Payout is called")

	requestBody := &api.Request{
		Merchant:   a.channelParams.MerchantId,
		WithdrawID: strconv.FormatInt(txn.TxnId, 10),
		Amount:     txn.TxnAmountSrc,
		CardData: &api.CardData{
			OwnerName:  txn.Customer.FullName,
			CardNumber: txn.PaymentData.Object.Credentials,
		},
	}

	response, err := a.api.MakePayout(ctx, requestBody, a.channelParams.SecretKey)
	if err != nil {
		log.Println("An error occured while MakePayout and get response")
		return nil, err
	}

	return &acquirer.TransactionStatus{ // ?
		Status:   acquirer.APPROVED,
		GtwTxnId: &response.Id,
	}, nil
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}
