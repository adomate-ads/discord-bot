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
	
	if m.Content == "!status" {
		_, err := s.ChannelMessageSend(m.ChannelID, "I'm alive!")
		fmt.Println(err)
	} else if m.Content == "!isdown" {
		_, err := s.ChannelMessageSend(m.ChannelID, "All services are operational")
		fmt.Println(err)
	} else if m.Content[0] == '!'{
		_, err := s.ChannelMessageSend(m.ChannelID, "invalid command")
		fmt.Println(err)
	}
}

type Message struct {
	Type       string    `json:"type" example:"error/warning/log"`
	Message    string    `json:"message"`
	Suggestion string    `json:"suggestion,omitempty"`
	Time       time.Time `json:"time,omitempty" example:"2018-12-12T11:45:26.371Z"`
	Origin     string    `json:"origin" example:"api/gac"`
}
/* message example
{
"type":"Error",
"message":"test",
"suggestion":"lol",
"origin":"api",
"time":"2018-12-12T11:45:26.371Z"
}
*/

func sendDiscordMessage(s *discordgo.Session, channelID string, msg Message) error {
	embedFull := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{Name: "AdomateHelpDesk"},
		Color:  0x800000, // Maroon - should change later based on message
		Description: fmt.Sprintf("%s from %s", msg.Type, msg.Origin),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Type: ",
				Value:  msg.Type,
				Inline: true,
			},
			{
				Name:   "Origin: ",
				Value:  msg.Origin,
				Inline: true,
			},
			{
				Name: "Message: ",
				Value: func() string {
					if msg.Message == "" {
						return "-"
					}
					return "```" + msg.Message + "```"
				}(),
				Inline: false,
			},
			{
				Name: "Suggestion: ",
				Value: func() string {
					if msg.Suggestion == "" {
						return "-"
					}
					return "```" + msg.Suggestion + "```"
				}(),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Time: %s", msg.Time.Format("02 Jan 06 15:04:01 CDT")),
		},
	}

	// switch embed color depending on msg type
	switch msg.Type {
	case "Error":
		embedFull.Color = 0xFF0000
	case "Warning":
		embedFull.Color = 0xFFFF00
	case "Success":
		embedFull.Color = 0x00FF00
	case "Log":
		embedFull.Color = 0x0000FF
	default:
		embedFull.Color = 0x000000
	}

	message := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
				embedFull,
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "Delete Message",
						Style: discordgo.DangerButton,
						Disabled: false,
						CustomID: "response_delete",
					},
				},
			},
		},
	}

	sentMsg, err := s.ChannelMessageSendComplex(channelID, message)

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if i.Type == discordgo.InteractionMessageComponent && i.MessageComponentData().CustomID == "response_delete" {
            err := s.ChannelMessageDelete(channelID, sentMsg.ID)
			if err != nil {
			fmt.Println("Error deleting message")
			}
		}
    })
	return err
}