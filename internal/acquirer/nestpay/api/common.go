package api

import (
	"encoding/base64"
	"strings"
	"testStand/internal/acquirer/helper"
)

type PaymentRequest struct {
	XMLName       struct{} `xml:"CC5Request"`
	Name          string   `xml:"Name"`
	Password      string   `xml:"Password"`
	ClientId      string   `xml:"ClientId"`
	Type          string   `xml:"Type"`
	Number        string   `xml:"Number"`                  //md
	PayerAuthCode string   `xml:"PayerAuthenticationCode"` //cavv
	PayerSecLevel string   `xml:"PayerSecurityLevel"`      //eci
	PayerTxnId    string   `xml:"PayerTxnId"`              //xid
	Currency      string   `xml:"Currency"`
	OrderId       string   `xml:"OrderId"`
}

type NestpayResponse struct {
	ErrMsg         string `xml:"ErrMsg"`
	ProcReturnCode string `xml:"ProcReturnCode"`
	Response       string `xml:"Response"`
	OrderId        string `xml:"OrderId"`
}

type Request3Ds struct {
	ClientId                    string `url:"clientid"`
	StoreType                   string `url:"storetype"`
	TranType                    string `url:"trantype"`
	Currency                    string `url:"currency"`
	Oid                         string `url:"oid"`
	Rnd                         string `url:"rnd"`
	Pan                         string `url:"pan"`
	EcomPaymentCardExpDateMonth string `url:"Ecom_Payment_Card_ExpDate_Month"`
	EcomPaymentCardExpDateYear  string `url:"Ecom_Payment_Card_ExpDate_Year"`
	Cv2                         string `url:"cv2"`
	HashAlgorithm               string `url:"hashAlgorithm"`
	Hash                        string `url:"hash"`
	Encoding                    string `url:"encoding"`
}

type PaymentData struct {
	MD   string
	CAVV string
	ECI  string
	XID  string
}

func createHash(req *Request3Ds, storeKey string) string {
	params := []string{
		req.ClientId,
		req.Currency,
		req.Cv2,
		req.EcomPaymentCardExpDateMonth,
		req.EcomPaymentCardExpDateYear,
		req.HashAlgorithm,
		req.Oid,
		req.Pan,
		req.Rnd,
		req.StoreType,
		req.TranType,
	}
	data := strings.Join(params, "|") + "|" + storeKey
	hashed := helper.GenerateSHA512Hash(data)
	hash := base64.StdEncoding.EncodeToString(hashed[:])
	return hash
}
