package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
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

type Message struct {
	Type       string    `json:"type"`
	Message    string    `json:"message"`
	Suggestion string    `json:"suggestion,omitempty"`
	Time       time.Time `json:"time,omitempty"`
}

func sendDiscordMessage(s *discordgo.Session, channelID string, msg Message) error {
	_, err := s.ChannelMessageSend(channelID, fmt.Sprintf("**TYPE**\n%s\nMessage:%s\nSuggestion:%s", msg.Type, msg.Message, msg.Suggestion))
	return err
}
