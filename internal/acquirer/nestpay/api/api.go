package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"testStand/internal/acquirer/helper"
)

const (
	paymentEndpoint    = "fim/api"
	payment3dsEndpoint = "fim/est3dgate"
)

type Client struct {
	baseAddress string
	client      *http.Client
}

func NewClient(baseAddress string, timeout *int) *Client {
	return &Client{
		baseAddress: baseAddress,
		client:      http.DefaultClient,
	}
}

func (c *Client) Validate3Ds(ctx context.Context, req *Request3Ds, storeKey string) (PaymentData, error) {
	req.Hash = createHash(req, storeKey)

	form := url.Values{}
	form.Set("clientid", req.ClientId)
	form.Set("storetype", req.StoreType)
	form.Set("trantype", req.TranType)
	form.Set("currency", req.Currency)
	form.Set("oid", req.Oid)
	form.Set("rnd", req.Rnd)
	form.Set("pan", req.Pan)
	form.Set("Ecom_Payment_Card_ExpDate_Month", req.EcomPaymentCardExpDateMonth)
	form.Set("Ecom_Payment_Card_ExpDate_Year", req.EcomPaymentCardExpDateYear)
	form.Set("cv2", req.Cv2)
	form.Set("hashAlgorithm", req.HashAlgorithm)
	form.Set("hash", req.Hash)
	form.Set("encoding", req.Encoding)

	body, err := c.sendRequest(ctx, payment3dsEndpoint, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return PaymentData{}, err
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return PaymentData{}, err
	}

	if values.Get("mdStatus") != "1" {
		return PaymentData{}, fmt.Errorf("auth failed, mdStatus=%s, ErrMsg=%s", values.Get("mdStatus"), values.Get("ErrMsg"))
	}

	return PaymentData{
		MD:   values.Get("md"),
		CAVV: values.Get("cavv"),
		ECI:  values.Get("eci"),
		XID:  values.Get("xid"),
	}, nil
}

func (c *Client) MakePayment(ctx context.Context, request *PaymentRequest) (*NestpayResponse, error) {
	resp := &NestpayResponse{}
	err := c.makeRequest(ctx, request, resp, paymentEndpoint)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) makeRequest(ctx context.Context, payload, outResponse any, endpoint string) error {
	body, err := xml.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	resp, err := c.sendRequest(ctx, endpoint, "application/xml", bytes.NewReader(body))
	if err != nil {
		return err
	}
	if err := xml.NewDecoder(bytes.NewReader(resp)).Decode(outResponse); err != nil {
		return fmt.Errorf("failed decode: %w, body=%s", err, string(resp))
	}
	return nil
}

func (c *Client) sendRequest(ctx context.Context, endpoint, contentType string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, helper.JoinUrl(c.baseAddress, endpoint), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed, status=%d, body=%s", resp.StatusCode, string(b))
	}
	return b, nil
}
