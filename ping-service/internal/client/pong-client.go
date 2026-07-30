package client

import (
    "context"
    "fmt"
    "net"
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
    transport := &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        DialContext: (&net.Dialer{
            Timeout:   5 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        MaxIdleConns:          100,              // Toplam boştaki bağlantı sınırı
        MaxIdleConnsPerHost:  100,              // Host başına boştaki bağlantı sınırı (Varsayılan 2 idi!)
        IdleConnTimeout:      90 * time.Second, // Boşta kalan bağlantının yaşam süresi
        TLSHandshakeTimeout: 5 * time.Second,
    }

    return &pongClient{
        httpClient: &http.Client{
            Transport: transport,
            Timeout:   3 * time.Second, // Toplam HTTP isteği zaman aşımı
        },
        pongURL: pongURL,
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