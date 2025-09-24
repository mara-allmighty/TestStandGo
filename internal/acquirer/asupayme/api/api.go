package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"testStand/internal/acquirer/helper"
)

type Client struct {
	baseAddress string
	secretKey   string
	apiKey      string
	merchantID  string
	httpClient  *http.Client
}

func NewClient(ctx context.Context, merchantID, baseAddress, apiKey, secretKey string, timeout *int) *Client {
	return &Client{
		baseAddress: baseAddress,
		secretKey:   secretKey,
		apiKey:      apiKey,
		merchantID:  merchantID,
		httpClient:  http.DefaultClient,
	}
}

// Payout
func (c *Client) MakePayout(ctx context.Context, request *Request, apiKey string) (*Response, error) {
	log.Println("MakePayout is called")

	sign, err := createSign(request, apiKey)
	if err != nil {
		log.Println("An error occured while creating a sign")
		return nil, err
	}

	resp := &Response{}
	err = c.makeRequest(ctx, request, resp, sign)
	if err != nil {
		log.Println("An error occured while sending request to asupayme")
		return nil, err
	}

	return resp, nil
}

func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, endpoint string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Println("An error occured while marshal payload")
		return err
	}

	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), bytes.NewReader(body))
	if err != nil {
		log.Println("An error occured while http.NewRequest()")
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Println("An error occured while httpCliend.Do working")
		return err
	}

	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(bodyBytes, &outResponse)
	if err != nil {
		log.Println("An error occured while Unmarshal bodyBytes to &outResponse")
		return nil
	}

	return nil
}
