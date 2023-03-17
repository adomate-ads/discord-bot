package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!status" {
		_, err := s.ChannelMessageSend(m.ChannelID, "I'm alive!")
		if err != nil {
			fmt.Println(err)
		}
	} else if m.Content[0] == '!' {
		_, err := s.ChannelMessageSend(m.ChannelID, "Unknown Command Dumbass")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func sendDiscordMessage(s *discordgo.Session, channelID string, content string) error {
	_, err := s.ChannelMessageSend(channelID, content)
	return err
}
