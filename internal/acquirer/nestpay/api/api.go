package api

import (
	"context"
	"net/http"
)

const (
	paymentPath = ""
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
func (c *Client) MakePayment() {
	return
}

func (c *Client) makeRequest(_ context.Context, payload, outResponse any, endpoint string) error {
	return nil
}
