package api

// Payment/Payout
type AlpexRequest struct {
	FiatSymbol      string `json:"fiat_symbol"`
	FiatAmount      int64  `json:"fiat_amount"`
	CustomerName    string `json:"customer_name"`
	CustomerAddress string `json:"customer_address"`
	Direction       string `json:"direction"`
	GateId          string `json:"gate_id"`
	WebhookUrl      string `json:"webhook_url"`
}

type AlpexResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Id     string `json:"_id"`
}

type AlpexCallbackResponse struct {
	Id         string `json:"_id"`
	Status     string `json:"status"`
	Signature  string `json:"signature"`
	ExternalId string `json:"external_id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	AccessToken string `json:"access_token"`
}
