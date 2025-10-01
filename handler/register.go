package handler

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleMessageCreate(f func(s *discordgo.Session, m *discordgo.MessageCreate) error) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if err := f(s, m); err != nil {
			s.ChannelMessage(m.ChannelID, "エラーが発生しました")
			log.Println(err.Error())
		}
	}
}
