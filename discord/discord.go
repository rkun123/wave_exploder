package discord

import "context"

type contextKey struct{}

var discordKey = contextKey{}

type Discord struct {
	token string
}

// APIリクエストのボディ
type apiRequest struct {
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

func InitDiscord(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, discordKey, newDiscord(token))
}

func newDiscord(token string) *Discord {
	return &Discord{
		token: token,
	}
}

func GetDiscord(ctx context.Context) *Discord {
	return ctx.Value(discordKey).(*Discord)
}
