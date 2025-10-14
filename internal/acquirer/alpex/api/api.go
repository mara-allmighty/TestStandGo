package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"testStand/internal/acquirer/helper"

	"github.com/labstack/gommon/log"
)

const (
	getTokenEndpoint = "/v1/auth/login"
	payInOutEndpoint = "/v1/offer/external"
	signatureEnd     = "v1/user/generate-signature-key"
)

type Client struct {
	login       string
	password    string
	baseAddress string
	httpClient  *http.Client
}

func NewClient(ctx context.Context, baseAddress, login, password string) *Client {
	return &Client{
		login:       login,
		password:    password,
		baseAddress: baseAddress,
		httpClient:  http.DefaultClient,
	}
}

// Pay In/Out
func (c *Client) MakePay(ctx context.Context, requestBody *AlpexRequest) (*AlpexResponse, error) {
	logger := log.New("api/makePay")

	response := &AlpexResponse{}
	err := c.makeRequest(ctx, requestBody, response, payInOutEndpoint)
	if err != nil {
		logger.Info("Error occured while making request")
		return nil, err
	}

	return response, nil
}

// Send request
func (c *Client) makeRequest(_ context.Context, payload, outResponse any, endpoint string) error {
	logger := log.New("api.makeRequest")

	token, err := c.getToken() // +
	if err != nil {
		logger.Info("Error occured while get token")
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logger.Info("Error occured while marshal")
		return err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		logger.Info("Error occured while creating req")
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Info("Error occured while sending http")
		return err
	}
	data, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("\n--------makeRequest-------- resp with callb\n%s\n", string(data))

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		logger.Info("Error occured while decode")
		return nil
	}

	return nil
}

// Validte callback
func (c *Client) IsCallbackSignValid(callback *Callback) (bool, error) {
	apiKey, err := c.getToken()
	if err != nil {
		return false, err
	}
	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, signatureEnd), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var callbackSignature string
	if err := json.NewDecoder(resp.Body).Decode(&callbackSignature); err != nil {
		return false, err
	}

	mySignature := createSign(callback, callbackSignature)
	if mySignature != callback.Signature {
		return false, nil
	}

	return true, nil
}

// Get token
func (c *Client) getToken() (string, error) {
	logger := log.New("getToken")

	data := LoginRequest{
		Email:    c.login,
		Password: c.password,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Info("Err occured while marshal data")
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, getTokenEndpoint), bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Info("Err occured while creating req")
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(req)
	if err != nil {
		logger.Info("Err occured while sending http")
		return "", err
	}
	defer response.Body.Close()

	var token Token
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		logger.Info("Err occured while finally getting token")
		return "", err
	}

	return token.AccessToken, nil
}
