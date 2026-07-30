package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type PongClient interface {
	SendPongRequest(ctx context.Context) error
}

type pongClient struct {
	httpClient *http.Client
	pongURL    string
}

func NewPongClient(pongURL string) PongClient {
	return &pongClient{
		httpClient: &http.Client{},
		pongURL:    pongURL,
	}
}

func (c *pongClient) SendPongRequest(ctx context.Context) error {
	// İsteğe katı 2 saniye timeout konuyor (Cascading Failure koruması)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodPost, c.pongURL+"/internal/process", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("pong-service baglanti hatasi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pong-service beklenmeyen statu kodu dondu: %d", resp.StatusCode)
	}

	return nil
}