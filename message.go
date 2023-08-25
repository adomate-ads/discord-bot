package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) == 0 {
		return
	}
}

type Message struct {
	Type    string    `json:"type" example:"error/warning/log"`
	Title   string    `json:"title"`
	Message string    `json:"message,omitempty"`
	Time    time.Time `json:"time,omitempty" example:"2018-12-12T11:45:26.371Z"`
	Origin  string    `json:"origin" example:"api/gac"`
}

/*
	message example
	{
	"type":"Error/Warning/Success/Log",
	"title":"Add title",
	"message":"Add message",
	"origin":"API",
	"time":"2023-04-24T08:45:26.371Z"
	}
*/

func sendDiscordMessage(s *discordgo.Session, msg Message) error {
	unixTime := msg.Time.Unix()
	timestampStr := time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")

	channelID := ""
	statusIcon := ""

	switch msg.Type {
	case "Error":
		channelID = "1108183014468497469"
		statusIcon = ":red_square:"
	case "Warning":
		channelID = "1108183014468497469"
		statusIcon = ":yellow_square:"
	case "Log":
		channelID = "1108183035502936165"
		statusIcon = ":white_large_square:"
	default:
		channelID = "1108183035502936165"
		statusIcon = ":white_circle:"
	}

	_, err := s.ChannelMessageSend(channelID, fmt.Sprintf("%s[%s][%s]: %s - %s", statusIcon, timestampStr, msg.Origin, msg.Title, msg.Message))
	if err != nil {
		return err
	}

	return nil
}
