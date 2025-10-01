package handler

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/rkun123/wave_exploder/songlink"
)

type Handler interface {
	MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) error
}

type HandlerImpl struct {
	sl songlink.SongLink
}

func New(sl songlink.SongLink) Handler {
	return &HandlerImpl{
		sl,
	}
}

func (h HandlerImpl) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) error {
	ctx := context.Background()

	explodeChannelID := os.Getenv("EXPLODE_CHANNEL_ID")
	if explodeChannelID == "" {
		log.Fatal("EXPLODE_CHANNEL_ID not found in Environment Variables")
		return fmt.Errorf("EXPLODE_CHANNEL_ID not found in Environment Variables")
	}

	// 指定されたチャンネルからの投稿かチェック
	if m.ChannelID == explodeChannelID {
		if m.Author.Bot {
			// Botの投稿（自分も含む）は無視する
			return nil
		}

		if m.Content == "ping" {
			if _, err := s.ChannelMessageSend(m.ChannelID, "pong"); err != nil {
				return err
			}
			return nil
		}

		if _, err := url.ParseRequestURI(m.Content); err != nil {
			// URIでない場合は握りつぶす
			log.Printf("Is not a URI: %v", err)
			return nil
		}

		// タイピング表示を起動する
		if err := s.ChannelTyping(m.ChannelID); err != nil {
			log.Printf("Failed to start typing: %v", err)
			return err
		}

		songLink, err := h.sl.Info(ctx, m.Content)
		if err != nil {
			log.Printf("Failed to get link: %v", err)
			return nil // 握りつぶす
		}

		message, err := formatSongLinkResponse(songLink)
		if err != nil {
			log.Printf("Failed to format message: %v", err)
			return err
		}

		// as Reply
		message.Reference = &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
		}

		if _, err := s.ChannelMessageSendComplex(m.ChannelID, message); err != nil {
			log.Println(m.ChannelID)
			log.Printf("Failed to send message: %v", err)
			return err
		}
	}
	return nil
}

func formatSongLinkResponse(r *songlink.LinkResponse) (*discordgo.MessageSend, error) {
	// ref: https://linktree.notion.site/API-d0ebe08a5e304a55928405eb682f6741

	linkButtons := make([]discordgo.MessageComponent, 0, len(r.LinksByPlatform))

	for platform, link := range r.LinksByPlatform {
		switch platform {
		case "spotify":
			linkButtons = append(linkButtons, discordgo.Button{
				Label: "Spotify",
				Style: discordgo.LinkButton,
				URL:   link.URL,
			})
		case "youtube":
			linkButtons = append(linkButtons, discordgo.Button{
				Label: "YouTube",
				Style: discordgo.LinkButton,
				URL:   link.URL,
			})
		case "appleMusic":
			linkButtons = append(linkButtons, discordgo.Button{
				Label: "Apple Music",
				Style: discordgo.LinkButton,
				URL:   link.URL,
			})
		case "youtubeMusic":
			linkButtons = append(linkButtons, discordgo.Button{
				Label: "YouTube Music",
				Style: discordgo.LinkButton,
				URL:   link.URL,
			})
		case "amazonMusic":
			linkButtons = append(linkButtons, discordgo.Button{
				Label: "Amazon Music",
				Style: discordgo.LinkButton,
				URL:   link.URL,
			})
		case "soundcloud":
			linkButtons = append(linkButtons, discordgo.Button{
				Label: "SoundCloud",
				Style: discordgo.LinkButton,
				URL:   link.URL,
			})
		case "bandcamp":
			linkButtons = append(linkButtons, discordgo.Button{
				Label: "Bandcamp",
				Style: discordgo.LinkButton,
				URL:   link.URL,
			})
		}
	}

	if len(linkButtons) == 0 {
		return nil, fmt.Errorf("no link found")
	}

	actionsRows := make([]discordgo.ActionsRow, (len(linkButtons)-1)/5+1)
	for i, button := range linkButtons {
		exists := actionsRows[i/5]
		actionsRows[i/5] = discordgo.ActionsRow{
			Components: append(exists.Components, button),
		}
	}

	components := make([]discordgo.MessageComponent, len(actionsRows))
	for i, row := range actionsRows {
		components[i] = row
	}

	return &discordgo.MessageSend{
		Components: components,
	}, nil

}
