package ports

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type PaymentClient struct {
	baseURL string
	client  *http.Client
}

func NewPaymentClient(url string, c *http.Client) *PaymentClient {
	return &PaymentClient{baseURL: url, client: c}
}

func (p *PaymentClient) Pay(ctx context.Context, orderID string, amount int64) (string, error) {
	body := map[string]interface{}{
		"order_id": orderID,
		"amount":   amount,
	}

	data, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/payments", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "Declined", nil
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if status, ok := result["Status"].(string); ok {
		return status, nil
	}
	if status, ok := result["status"].(string); ok {
		return status, nil
	}

	return "", nil
}
