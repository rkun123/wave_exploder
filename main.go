package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coder/websocket"
	"github.com/joho/godotenv"
	"github.com/rkun123/wave_exploder/discord"
	"github.com/rkun123/wave_exploder/songlink"
)

// Discord WebSocketイベントペイロード
type Payload struct {
	Op int         `json:"op"`
	D  interface{} `json:"d,omitempty"`
	S  *int        `json:"s,omitempty"`
	T  string      `json:"t,omitempty"`
}

// 認証ペイロード
type IdentifyPayload struct {
	Token      string      `json:"token"`
	Intents    int         `json:"intents"`
	Properties Properties  `json:"properties"`
	Presence   interface{} `json:"presence"`
}

type Properties struct {
	OS      string `json:"$os"`
	Browser string `json:"$browser"`
	Device  string `json:"$device"`
}

// ハートビートペイロード
type HeartbeatPayload struct {
	Op int  `json:"op"`
	D  *int `json:"d"`
}

// 投稿者情報
type Author struct {
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}

// メッセージペイロードのデータ部分
type MessageData struct {
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id"`
	ID        string `json:"id"`
	Content   string `json:"content"`
	Author    Author `json:"author"`
}

// Interactionペイロードのデータ部分
type InteractionData struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

// APIリクエストのボディ
type APIRequest struct {
	Content   string      `json:"content,omitempty"`
	TTS       bool        `json:"tts,omitempty"`
	Embeds    interface{} `json:"embeds,omitempty"`
	Name      string      `json:"name,omitempty"`
	Type      int         `json:"type,omitempty"`
	Topic     string      `json:"topic,omitempty"`
	Bitrate   int         `json:"bitrate,omitempty"`
	UserLimit int         `json:"user_limit,omitempty"`
	Options   interface{} `json:"options,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// Gateway URLのレスポンス
type GatewayResponse struct {
	URL string `json:"url"`
}

var GatewaySeqNum *int

// GetGateway APIを呼び出してWebSocket URLを取得する
func getGatewayURL(ctx context.Context) (string, error) {
	resp, err := http.Get("https://discord.com/api/v9/gateway")
	if err != nil {
		return "", fmt.Errorf("failed to get gateway URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gateway API returned non-200 status: %d", resp.StatusCode)
	}

	var gatewayResp GatewayResponse
	if err := json.NewDecoder(resp.Body).Decode(&gatewayResp); err != nil {
		return "", fmt.Errorf("failed to decode gateway URL response: %w", err)
	}

	return gatewayResp.URL + "?v=10&encoding=json", nil
}

func main() {
	// .envファイルから環境変数をロード
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN not found in .env file")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = discord.InitDiscord(ctx, token)
	ctx = songlink.InitSonglink(ctx)

	// 終了シグナルを待機
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("Shutting down...")
		cancel()
	}()

	// Gateway URLを取得
	gatewayURL, err := getGatewayURL(ctx)
	if err != nil {
		log.Fatalf("Failed to get Gateway URL: %v", err)
	}
	log.Printf("Successfully obtained gateway URL: %s", gatewayURL)

	// WebSocket接続の確立
	conn, _, err := websocket.Dial(ctx, gatewayURL, nil)
	if err != nil {
		log.Fatalf("WebSocket connection failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// 認証ペイロードの構築と送信
	identifyPayload := &Payload{
		Op: 2,
		D: IdentifyPayload{
			Token:   token,
			Intents: 35328,
			Properties: Properties{
				OS:      "linux",
				Browser: "chrome",
				Device:  "chrome",
			},
		},
	}

	sendJSON(ctx, conn, identifyPayload)
	log.Println("Identify payload sent.")

	heartbeatTicker := time.NewTicker(24 * time.Second) // 初期値
	defer heartbeatTicker.Stop()

	// WebSocketからのメッセージ受信とハートビート送信を同時に処理
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-heartbeatTicker.C:
				// ハートビートペイロードの送信
				heartbeatPayload := &Payload{
					Op: 1,
					D:  GatewaySeqNum,
				}
				sendJSON(ctx, conn, heartbeatPayload)
				log.Println("Heartbeat sent.")
			}
		}
	}()

	for {
		// WebSocketからメッセージを受信
		_, data, err := conn.Read(ctx)
		if err != nil {
			log.Printf("Failed to read message: %v", err)
			return
		}

		var payload Payload
		if err := json.Unmarshal(data, &payload); err != nil {
			log.Printf("Failed to unmarshal JSON: %v", err)
			continue
		}

		log.Printf("Received payload: %+v\n", payload)

		switch payload.Op {
		case 10: // Hello
			var d struct {
				HeartbeatInterval int `json:"heartbeat_interval"`
			}
			if err := mapToStruct(payload.D, &d); err != nil {
				log.Printf("Failed to map Hello data: %v", err)
				continue
			}
			heartbeatInterval := time.Duration(d.HeartbeatInterval) * time.Millisecond
			heartbeatTicker.Reset(heartbeatInterval)
			log.Printf("Heartbeat interval updated to: %v", heartbeatInterval)
		case 0: // Dispatch
			// tフィールドでイベントタイプを処理
			switch payload.T {
			case "READY":
				log.Println("Bot is online! 🚀")
			case "MESSAGE_CREATE":
				var d MessageData
				if err := mapToStruct(payload.D, &d); err != nil {
					log.Printf("Failed to map MessageCreate data: %v", err)
					continue
				}
				handleMessageCreate(ctx, d)
			case "INTERACTION_CREATE":
				var d InteractionData
				if err := mapToStruct(payload.D, &d); err != nil {
					log.Printf("Failed to map InteractionCreate data: %v", err)
					continue
				}
				handleInteractionCreate(ctx, token, d)
			}
		}

		// Update SeqNum
		if payload.S != nil {
			GatewaySeqNum = payload.S
		}
	}
}

// JSONエンコーディングとWebSocket送信
func sendJSON(ctx context.Context, conn *websocket.Conn, v interface{}) {
	w, err := conn.Writer(ctx, websocket.MessageText)
	if err != nil {
		log.Printf("Failed to get WebSocket writer: %v", err)
		return
	}
	defer w.Close()

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Failed to encode and send JSON: %v", err)
	}
}

// API呼び出しを行う関数
func makeAPIRequest(ctx context.Context, token, method, url string, body io.Reader) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("API request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("API Response: %s", string(respBody))
}

// map[string]interface{}を構造体に変換
func mapToStruct(m interface{}, v interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// MESSAGE_CREATEイベントを処理
func handleMessageCreate(ctx context.Context, d MessageData) error {
	cli := discord.GetDiscord(ctx)
	sl := songlink.GetSonglink(ctx)

	explodeChannelID := os.Getenv("EXPLODE_CHANNEL_ID")
	if explodeChannelID == "" {
		log.Fatal("EXPLODE_CHANNEL_ID not found in .env file")
		return fmt.Errorf("EXPLODE_CHANNEL_ID not found in .env file")
	}

	// 指定されたチャンネルからの投稿かチェック
	if d.ChannelID == explodeChannelID {
		if d.Author.Bot {
			// Botの投稿（自分も含む）は無視する
			return nil
		}

		if _, err := url.ParseRequestURI(d.Content); err != nil {
			// URIでない場合は握りつぶす
			log.Printf("Is not a URI: %v", err)
			return nil
		}

		// タイピング表示を起動する
		if err := cli.StartTyping(ctx, d.ChannelID); err != nil {
			log.Printf("Failed to start typing: %v", err)
			return err
		}

		songLink, err := sl.Link(ctx, d.Content)
		if err != nil {
			log.Printf("Failed to get link: %v", err)
			return nil // 握りつぶす
		}

		if err := cli.SendMessage(ctx, d.ChannelID, &discord.MessageSend{Content: songLink}); err != nil {
			log.Printf("Failed to send message: %v", err)
			return err
		}
	}

	switch d.Content {
	case "help":
		bodyData := &discord.MessageSend{
			Content: "Hello, World!",
			TTS:     false,
			Embeds: []map[string]interface{}{
				{"title": "Hello World", "description": "Embed Message"},
			},
		}
		if err := cli.SendMessage(ctx, d.ChannelID, bodyData); err != nil {
			log.Printf("Failed to send message: %v", err)
			return err
		}
	case "react":
		if err := cli.ReactMessage(ctx, d.ChannelID, d.ID, "👍"); err != nil {
			log.Printf("Failed to react to message: %v", err)
			return err
		}
	}
	return nil
}

// INTERACTION_CREATEイベントを処理
func handleInteractionCreate(ctx context.Context, token string, d InteractionData) {
	url := fmt.Sprintf("https://discord.com/api/v9/interactions/%s/%s/callback", d.ID, d.Token)
	bodyData := &APIRequest{
		Type: 4,
		Data: map[string]interface{}{
			"content": "Hello",
		},
	}
	body, _ := json.Marshal(bodyData)
	makeAPIRequest(ctx, token, "POST", url, bytes.NewReader(body))
}
