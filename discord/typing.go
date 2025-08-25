package discord

import (
	"context"
	"fmt"
)

func (d Discord) StartTyping(ctx context.Context, channelID string) error {
	url := fmt.Sprintf("https://discord.com/api/v9/channels/%s/typing", channelID)
	return makeAPIRequest(ctx, d.token, "POST", url, nil)
}
