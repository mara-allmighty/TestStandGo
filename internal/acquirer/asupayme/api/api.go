package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"testStand/internal/acquirer/helper"
)

const (
	payoutPath = "/api/v1/withdraw" // #
)

type Client struct {
	baseAddress string
	secretKey   string
	apiKey      string
	httpClient  *http.Client
}

func NewClient(ctx context.Context, baseAddress, apiKey, secretKey string) *Client {
	return &Client{
		baseAddress: baseAddress,
		secretKey:   secretKey,
		apiKey:      apiKey,
		httpClient:  http.DefaultClient,
	}
}

// Payout
func (c *Client) MakePayout(ctx context.Context, requestBody *WithdrawRequestBody) (*AsupaymeResponse, error) {
	sign := createSign(requestBody, c.secretKey)
	requestBody.Signature = sign

	resp := &AsupaymeResponse{}

	err := c.makeRequest(ctx, requestBody, resp, payoutPath)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) makeRequest(_ context.Context, payload, outResponse any, endpoint string) error {

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req) // #
	if err != nil {
		return err
	}
	data, _ := httputil.DumpResponse(resp, true)
	fmt.Println(string(data))

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&outResponse)
	if err != nil {
		return nil
	}

	//

	return nil
}
