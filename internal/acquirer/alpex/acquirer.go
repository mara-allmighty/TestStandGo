package alpex

import (
	"context"
	"fmt"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/alpex/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/labstack/gommon/log"

	"github.com/shopspring/decimal"
)

type ChannelParams struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	WebhookUrl string `json:"webhook_url"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type GatewayParams struct {
	Transport            Transport        `json:"transport"`
	PercentageDifference *decimal.Decimal `json:"percentage_difference"`
}

type Acquirer struct {
	api                  *api.Client
	dbClient             *repos.Repo
	channelParams        ChannelParams
	percentageDifference *decimal.Decimal
}

func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams ChannelParams, gatewayParams GatewayParams, callbackURL string) *Acquirer {
	return &Acquirer{
		api:                  api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.Login, channelParams.Password),
		dbClient:             db,
		channelParams:        channelParams,
		percentageDifference: gatewayParams.PercentageDifference,
	}
}

// Pay In
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	logger := log.New("alpex/payment")

	requestBody := &api.AlpexRequest{
		FiatSymbol:      txn.TxnCurrencySrc,
		FiatAmount:      txn.TxnAmountSrc,
		CustomerName:    txn.Customer.FullName,
		CustomerAddress: txn.PaymentData.Object.Credentials,
		Direction:       "BUY",
		WebhookUrl:      a.channelParams.WebhookUrl,
	}

	response, err := a.api.MakePay(ctx, requestBody)
	if err != nil {
		return nil, err
	}
	logger.Info(fmt.Sprintf("Alpex response: %v", response))

	if response.Status != "PENDING" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	return &acquirer.TransactionStatus{
		Status:   acquirer.PENDING,
		GtwTxnId: &response.Id,
	}, err
}

// Pay Out
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	logger := log.New("alpex/payout")

	requestBody := &api.AlpexRequest{
		FiatSymbol:      txn.TxnCurrencySrc,
		FiatAmount:      txn.TxnAmountSrc,
		CustomerName:    txn.Customer.FullName,
		CustomerAddress: txn.PaymentData.Object.Credentials,
		Direction:       "SELL",
		GateId:          txn.Customer.AccountId, // gate_id = [account_id || payment_data_bank || or what?]
		WebhookUrl:      a.channelParams.WebhookUrl,
	}

	response, err := a.api.MakePay(ctx, requestBody)
	if err != nil {
		return nil, err
	}
	logger.Info(fmt.Sprintf("Alpex response: %v", response))

	if response.Status != "PENDING" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	return &acquirer.TransactionStatus{
		Status:   acquirer.PENDING,
		GtwTxnId: &response.Id,
	}, err
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}
