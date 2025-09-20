package asupayme

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/asupayme/api"
	"testStand/internal/acquirer/helper"
	"testStand/internal/models"
	"testStand/internal/repos"

	"github.com/labstack/gommon/log"

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
	ApiKey    string `json:"api_key"`
	MerchId   string `json:"merchant_id"`
	SecretKey string `json:"secret_key"`
}

type Acquirer struct {
	api                  *api.Client
	dbClient             *repos.Repo
	channelParams        *ChannelParams
	callbackUrl          string
	percentageDifference *decimal.Decimal
}

// NewAcquirer
func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams, callbackUrl string) *Acquirer {
	return &Acquirer{
		channelParams:        channelParams,
		api:                  api.NewClient(ctx, gatewayParams.Transport.BaseAddress, channelParams.SecretKey, gatewayParams.Transport.Timeout),
		dbClient:             db,
		callbackUrl:          callbackUrl,
		percentageDifference: gatewayParams.PercentageDifference,
	}
}

// Payment
func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	request := &api.Request{
		UserRef:         fmt.Sprintf("%d", txn.TxnId),
		MerchId:         a.channelParams.MerchId,
		Extra:           txn.PayMethodId,
		Amount:          strconv.FormatInt(txn.TxnAmountSrc, 10),
		Currency:        txn.TxnCurrencySrc,
		UserId:          fmt.Sprintf("%s", txn.Customer.AccountId),
		UserIp:          txn.Customer.Ip,
		NotificationUrl: a.callbackUrl,
	}

	response, err := a.api.MakePayment(ctx, request)
	if err != nil {
		return nil, err
	}

	if !response.OK {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		Outputs: map[string]string{
			"credentials": response.P2PDestination,
			"bank":        response.P2PBank,
			"description": response.P2PName,
		},
		GtwTxnId: &response.Id,
	}

	return handleStatus(tr, response.Status)
}

// Payout
func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {

	requestBody := &api.Request{
		MerchId:         fmt.Sprintf("%d", txn.TxnId),
		Extra:           txn.PayMethodId,
		Amount:          strconv.FormatInt(txn.TxnAmountSrc, 10),
		Currency:        txn.TxnCurrencySrc,
		NotificationUrl: a.callbackUrl,
		UserRef:         fmt.Sprintf("%d", txn.TxnId),
	}

	response, err := a.api.MakePayout(ctx, requestBody, a.channelParams.SecretKey)
	if err != nil {
		return nil, err
	}

	if !response.OK {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code": response.Error,
			},
		}, nil
	}

	tr := &acquirer.TransactionStatus{
		GtwTxnId: &response.Id,
	}

	return handleStatus(tr, response.Status)
}

// HandleCallback
func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	logger := log.New("dev")

	callbackBody, ok := txn.TxnInfo["callback"]

	if !ok {
		return nil, errors.New("callback body is missing")
	}

	callback := api.Callback{}
	err := json.Unmarshal([]byte(callbackBody), &callback)
	if err != nil {
		logger.Error("Error unmarshalling callback body - ", callbackBody)
		return nil, err
	}

	if callback.Sign != base64.StdEncoding.EncodeToString(helper.GenerateHMAC(sha1.New, []byte(callback.Id), a.channelParams.SecretKey)) {
		logger.Error("Invalid callback - ", callbackBody)
		return nil, err
	}

	tr := &acquirer.TransactionStatus{}
	if len(callback.Description) != 0 {
		tr.Info = map[string]string{
			"ps_error_code": callback.Description,
		}
	}

	return handleStatus(tr, callback.Status)
}

// FinalizePending
func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	statusReq := &api.StatusRequest{
		Id:      *txn.GtwTxnId,
		MerchId: a.channelParams.MerchId,
	}

	resp, err := a.api.GetStatus(ctx, statusReq, a.channelParams.SecretKey)
	if err != nil {
		return nil, err
	}

	tr := &acquirer.TransactionStatus{}

	return handleStatus(tr, resp.Status)
}

func handleStatus(tr *acquirer.TransactionStatus, status string) (*acquirer.TransactionStatus, error) {
	switch status {
	case api.Reconciled:
		tr.Status = acquirer.APPROVED
		return tr, nil
	case api.Decline:
		tr.Status = acquirer.REJECTED
		return tr, nil
	case api.Pending:
		fallthrough
	default:
		tr.Status = acquirer.PENDING
		return tr, nil
	}
}
