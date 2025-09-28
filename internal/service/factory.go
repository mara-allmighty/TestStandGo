package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"testStand/internal/acquirer"
	"testStand/internal/acquirer/alpex"
	"testStand/internal/acquirer/asupayme"
	"testStand/internal/acquirer/auris"
	"testStand/internal/acquirer/paylink"
	"testStand/internal/acquirer/sequoia"
	"testStand/internal/models"
	"testStand/internal/repos"

	json "github.com/json-iterator/go"
	"github.com/labstack/gommon/log"
)

var ErrUnsupportedAcquirer = errors.New("unsupported acquirer")

const (
	AURIS    = "auris"
	SEQUOIA  = "sequoia"
	PAYLINK  = "paylink"
	ASUPAYME = "asupayme"
	ALPEX    = "alpex"
)

type Factory struct {
	dbClient *repos.Repo
}

// NewFactory
func NewFactory(dbClient *repos.Repo) *Factory {
	return &Factory{
		dbClient: dbClient,
	}
}

// Get Merchant and Acquirer params from database
func (f *Factory) Create(ctx context.Context, txn *models.Transaction) (any, error) {
	logger := log.New("dev")

	// Получаем конфиг
	gateway, err := f.dbClient.GetGateway(*txn.GtwName)
	if err != nil {
		logger.Error(fmt.Sprint(1, err))
		if err == sql.ErrNoRows {
			return nil, repos.ErrGtwNotFound
		}
		return nil, err
	}

	logger.Info(fmt.Sprintf("Loaded gateway: %v", gateway))

	// Получаем структуру Channel с кредами channel.Params клиента
	channel, err := f.dbClient.GetChannel(*txn.ChnName)
	if err != nil {
		logger.Error(fmt.Sprint(2, err))
		if err == sql.ErrNoRows {
			return nil, repos.ErrChnNotFound
		}
		return nil, err
	}

	callbackUrl := "" // TODO ЗАПОЛНИТЬ

	acq, err := f.create(ctx, txn, gateway, channel.Params, callbackUrl)

	return acq, err
}

// create & choose acquirer
func (f *Factory) create(ctx context.Context, txn *models.Transaction, gateway *repos.Gateway, channelParams repos.Params, callbackUrl string) (acquirer.Acquirer, error) {
	logger := log.New("dev")

	var err error
	var acq acquirer.Acquirer

	callbackUrl, err = url.JoinPath(callbackUrl, gateway.Adapter)
	if err != nil {
		return nil, err
	}

	switch gateway.Adapter {
	//
	case AURIS:
		var chParams auris.ChannelParams
		var gtwParams auris.GatewayParams
		if err = f.unmarshalParams(gateway.ParamsJson, channelParams.Credentials, &gtwParams, &chParams); err != nil {
			return nil, err
		}
		acq = auris.NewAcquirer(ctx, f.dbClient, chParams, gtwParams, callbackUrl)

	case SEQUOIA:
		var chParams sequoia.ChannelParams
		var gtwParams sequoia.GatewayParams
		if err = f.unmarshalParams(gateway.ParamsJson, channelParams.Credentials, &gtwParams, &chParams); err != nil {
			return nil, err
		}
		acq = sequoia.NewAcquirer(ctx, f.dbClient, &chParams, &gtwParams, callbackUrl)

	case PAYLINK:
		var chParams paylink.ChannelParams
		var gtwParams paylink.GatewayParams
		if err = f.unmarshalParams(gateway.ParamsJson, channelParams.Credentials, &gtwParams, &chParams); err != nil {
			return nil, err
		}
		acq = paylink.NewAcquirer(ctx, f.dbClient, &chParams, &gtwParams, callbackUrl)

	case ASUPAYME:
		var chParams asupayme.ChannelParams
		var gtwParams asupayme.GatewayParams
		if err = f.unmarshalParams(gateway.ParamsJson, channelParams.Credentials, &gtwParams, &chParams); err != nil {
			return nil, err
		}
		acq = asupayme.NewAcquirer(ctx, f.dbClient, chParams, gtwParams, callbackUrl)

	case ALPEX:
		var chnParams alpex.ChannelParams
		var gtwParams alpex.GatewayParams
		fmt.Printf("\nCreds: %s\n", channelParams.Credentials)

		if err = f.unmarshalParams(gateway.ParamsJson, channelParams.Credentials, &gtwParams, &chnParams); err != nil {
			return nil, err
		}
		fmt.Printf("chnParams after unmarshal: %v", chnParams)

		acq = alpex.NewAcquirer(ctx, f.dbClient, chnParams, gtwParams, callbackUrl)

	default:
		return nil, ErrUnsupportedAcquirer
	}

	logger.Info(fmt.Sprintf("Loaded acquirer: %s", gateway.Adapter))

	return acq, nil
}

// распаковываем реальные данные в структуры
func (f *Factory) unmarshalParams(gatewayParamsJson string, channelParamsJson []byte, gatewayParams any, channelParams any) error {
	if err := json.Unmarshal(channelParamsJson, channelParams); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(gatewayParamsJson), gatewayParams); err != nil {
		return err
	}
	return nil
}
