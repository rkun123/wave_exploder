package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/rkun123/wave_exploder/handler"
	"github.com/rkun123/wave_exploder/songlink"
	"go.uber.org/dig"
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
		if !errors.Is(err, fs.ErrNotExist) {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN not found in Environment Variables")
	}

	c := dig.New()

	if err := c.Provide(handler.New); err != nil {
		log.Fatal(err)
	}

	if err := c.Provide(songlink.New); err != nil {
		log.Fatal(err)
	}

	if err := c.Invoke(func(h handler.Handler, sl songlink.SongLink) error {

		session, err := discordgo.New(fmt.Sprintf("Bot %s", token))
		if err != nil {
			return err
		}

		session.AddHandler(handler.HandleMessageCreate(h.MessageCreate))
		session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
			log.Println("Bot is online! 🚀")
		})

		if err := session.Open(); err != nil {
			return err
		}

		// 終了シグナルを待機
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		<-sigCh

		if err := session.Close(); err != nil {
			return err
		}

		fmt.Println("👋Goodbye!!")
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
