package nestpay

import (
	"context"
	"fmt"
	"strconv"
	"testStand/internal/acquirer"
	"testStand/internal/acquirer/helper"
	"testStand/internal/acquirer/nestpay/api"
	"testStand/internal/models"
	"testStand/internal/repos"
)

type GatewayParams struct {
	Transport Transport `json:"transport"`
}

type Transport struct {
	BaseAddress string `json:"base_address"`
	Timeout     *int   `json:"timeout"`
}

type ChannelParams struct {
	ApiName     string `json:"Name"`
	ApiPassword string `json:"Password"`
	Currency    string `json:"Currency"`
	ClientId    string `json:"ClientId"`
	StoreKey    string `json:"StoreKey"`
}

type Acquirer struct {
	api           *api.Client
	dbClient      *repos.Repo
	channelParams *ChannelParams
}

func NewAcquirer(ctx context.Context, db *repos.Repo, channelParams *ChannelParams, gatewayParams *GatewayParams) *Acquirer {
	return &Acquirer{
		channelParams: channelParams,
		api:           api.NewClient(gatewayParams.Transport.BaseAddress, gatewayParams.Transport.Timeout),
		dbClient:      db,
	}
}

func (a *Acquirer) Payment(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	req3dsForm := &api.Request3Ds{
		ClientId:                    a.channelParams.ClientId,
		Currency:                    a.channelParams.Currency,
		Oid:                         strconv.FormatInt(txn.TxnId, 10),
		StoreType:                   "3d_pay",
		TranType:                    "Auth",
		Rnd:                         "asdf",
		Pan:                         txn.PaymentData.Object.Credentials,
		EcomPaymentCardExpDateMonth: txn.PaymentData.Object.ExpMonth,
		EcomPaymentCardExpDateYear:  txn.PaymentData.Object.ExpYear,
		Cv2:                         txn.PaymentData.Object.Cvv,
		Encoding:                    "utf-8",
		HashAlgorithm:               "ver3",
	}

	resp3ds, err := a.api.Validate3Ds(ctx, req3dsForm, a.channelParams.StoreKey)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp3ds)

	req := &api.PaymentRequest{
		Type:          "Auth",
		Currency:      a.channelParams.Currency,
		OrderId:       strconv.FormatInt(txn.TxnId, 10),
		Name:          a.channelParams.ApiName,
		Password:      a.channelParams.ApiPassword,
		ClientId:      a.channelParams.ClientId,
		Number:        resp3ds.MD,
		PayerAuthCode: resp3ds.CAVV,
		PayerSecLevel: resp3ds.ECI,
		PayerTxnId:    resp3ds.XID,
	}

	response, err := a.api.MakePayment(ctx, req)
	if err != nil {
		return nil, err
	}
	fmt.Println(response)

	if response.Response != "Approve" {
		return &acquirer.TransactionStatus{
			Status: acquirer.REJECTED,
			Info: map[string]string{
				"ps_error_code":    response.ProcReturnCode,
				"ps_error_message": response.ErrMsg,
			},
		}, nil
	}

	return &acquirer.TransactionStatus{
		Status:   acquirer.APPROVED,
		GtwTxnId: &response.OrderId,
	}, nil

}

func (a *Acquirer) Payout(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

func (a *Acquirer) HandleCallback(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}

func (a *Acquirer) FinalizePending(ctx context.Context, txn *models.Transaction) (*acquirer.TransactionStatus, error) {
	return helper.UnsupportedMethodError()
}
