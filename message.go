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

	if len(m.Content) == 0 {
		return
	}
	// FIXED removed nil check below
	// FIXME doesn't send msg to the channel
	if m.Content == "!status" {
		_, err := s.ChannelMessageSend(m.ChannelID, "I'm alive!")
			fmt.Println(err)
	} else if m.Content == "!isdown" {
		_, err := s.ChannelMessageSend(m.ChannelID, "All services are operational")
			fmt.Println(err)
	} else if m.Content[0] == '!' {
		_, err := s.ChannelMessageSend(m.ChannelID, "Invalid, try again!")
			fmt.Println(err)
	}
}

type Message struct {
	Type       string    `json:"type" example:"error/warning/log"`
	Message    string    `json:"message"`
	Suggestion string    `json:"suggestion,omitempty"`
	//FIXED use golang time package to get time from the user's location
	// Time       string 	 `json:"time,omitempty"` 
	Origin 	   string	 `json:"origin" example:"from:api/from:gdc"`
}

func sendDiscordMessage(s *discordgo.Session, channelID string, msg Message) error {

	embed_full := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{Name: "AdomateHelpDesk"},
    	Color:       0x800000, // Maroon - should change later based on message 
		// FIXED using switch statement for msg.Type
		// Red for error
		// Yellow for slow service
		// Green for fixes
		// Orange for server down
		// Blue for general

		Description: "This is a message from Team Adomate",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Type: ",
				Value:  msg.Type,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Message: ",
				Value:  msg.Message,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Suggestion: ",
				// FIXED  suggestion omit logic here
				Value:  func() string {
					if msg.Suggestion == "" {
						return "-"
					}
					return msg.Suggestion
				}(),
			},
			&discordgo.MessageEmbedField{ //FIXED Pass time with golang
				Name:   "Time: ",
				Value:  time.Now().Format("02 Jan 06 15:04 CDT"),
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
				Text: msg.Origin,
		},
	}
	switch msg.Type {
		case "Code Red":
			embed_full.Color = 0xFF0000
		case "Code Yellow":
			embed_full.Color = 0xFFFF00
		case "Code Green":
			embed_full.Color = 0x00FF00
		case "Code Orange":
			embed_full.Color = 0xFFA500
		case "Code Blue":
			embed_full.Color = 0x0000FF
		default:
			embed_full.Color = 0x000000
	}
	
		_, err := s.ChannelMessageSendEmbed(channelID, embed_full)
		return err
	
}
