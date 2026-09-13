package line

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	channelSecret      string
	channelAccessToken string
	httpClient         *http.Client
}

func NewClient(channelSecret, channelAccessToken string) Client {
	return Client{
		channelSecret:      channelSecret,
		channelAccessToken: channelAccessToken,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
	}
}

func (c Client) ValidateSignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(c.channelSecret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (c Client) ReplyText(ctx context.Context, replyToken, text string) error {
	payload := replyRequest{
		ReplyToken: replyToken,
		Messages: []replyMessage{
			{Type: "text", Text: text},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.line.me/v2/bot/message/reply", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("line reply failed with status %d", resp.StatusCode)
	}
	return nil
}

type replyRequest struct {
	ReplyToken string         `json:"replyToken"`
	Messages   []replyMessage `json:"messages"`
}

type replyMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
