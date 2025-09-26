package asupayme

import (
	"context"
	"fmt"
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
	MerchantID string `json:"merchant_id"`
}

type Acquirer struct {
	api                  *api.Client
	dbClient             *repos.Repo
	channelParams        ChannelParams
	callbackURL          string
	percentageDifference *decimal.Decimal
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams ChannelParams, gatewayParams GatewayParams, callbackURL string) *Acquirer {
	return &Acquirer{
		api:                  api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.ApiKey, channelParams.SecretKey),
		dbClient:             db,
		channelParams:        channelParams,
		percentageDifference: gatewayParams.PercentageDifference,
		callbackURL:          callbackURL,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	requestBody := &api.WithdrawRequestBody{
		Merchant:   a.channelParams.MerchantID,
		WithdrawID: strconv.FormatInt(txn.TxnId, 10),
		Amount:     strconv.FormatInt(txn.TxnAmountSrc, 10),
		CardData: &api.CardData{
			OwnerName:  txn.Customer.FullName,
			CardNumber: txn.PaymentData.Object.Credentials,
		},
	}

	response, err := a.api.MakePayout(ctx, requestBody)
	if err != nil {
		return nil, err
	}
	fmt.Println(response)

	if response.Status != "success" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Code,
			},
		}, nil
	}

	return &acquirer.TransactionStatus{
		Status:   acquirer.APPROVED,
		GtwTxnId: &response.Id, // #
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
