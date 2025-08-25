package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// MessageSend is the payload for sending a message.
type MessageSend struct {
	Content string      `json:"content,omitempty"`
	TTS     bool        `json:"tts,omitempty"`
	Embeds  interface{} `json:"embeds,omitempty"`
}

// SendMessage sends a message to a channel.
func (d Discord) SendMessage(ctx context.Context, channelID string, msg *MessageSend) error {
	url := fmt.Sprintf("https://discord.com/api/v9/channels/%s/messages", channelID)
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	return makeAPIRequest(ctx, d.token, "POST", url, bytes.NewReader(body))
}

// ReactMessage adds a reaction to a message.
func (d Discord) ReactMessage(ctx context.Context, channelID, messageID, emoji string) error {
	encodedEmoji := url.PathEscape(emoji)
	url := fmt.Sprintf("https://discord.com/api/v9/channels/%s/messages/%s/reactions/%s/@me", channelID, messageID, encodedEmoji)
	return makeAPIRequest(ctx, d.token, "PUT", url, nil)
}

// PinMessage pins a message in a channel.
func (d Discord) PinMessage(ctx context.Context, channelID, messageID string) error {
	url := fmt.Sprintf("https://discord.com/api/v9/channels/%s/pins/%s", channelID, messageID)
	return makeAPIRequest(ctx, d.token, "PUT", url, nil)
}

// DeleteMessage deletes a message.
func (d Discord) DeleteMessage(ctx context.Context, channelID, messageID string) error {
	url := fmt.Sprintf("https://discord.com/api/v9/channels/%s/messages/%s", channelID, messageID)
	return makeAPIRequest(ctx, d.token, "DELETE", url, nil)
}

// API呼び出しを行う関数
func makeAPIRequest(ctx context.Context, token, method, url string, body io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return err
	}
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("API request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return err
	}

	log.Printf("API Response: %s", string(respBody))
	return nil
}