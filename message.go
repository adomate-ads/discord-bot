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
	Type       string    `json:"type" example:"error/warning/log"`
	Message    string    `json:"message"`
	Suggestion string    `json:"suggestion,omitempty"`
	Time       time.Time `json:"time,omitempty"`
	Origin 	   string	 `json:"origin" example:"from:api/from:gdc"`
}

func sendDiscordMessage(s *discordgo.Session, channelID string, msg Message) error {

	embed_full := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{Name: "Adomate Bot"},
    	Color:       0x00ff00, // Green - should change later based on message 
		Description: "This is a message from the api bot",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Type:",
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
				Value:  msg.Suggestion,
				Inline: true,
			},
			// &discordgo.MessageEmbedField{ --- not sure how to pass time field into rabbitmq
			// 	Name:   "Time: ",
			// 	Value:  msg.Time.String(),
			// 	Inline: true,
			// },

		},
		Footer: &discordgo.MessageEmbedFooter{
				Text: msg.Origin,
		},
	}	
	embed_no_suggestion := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{Name: "Adomate Bot"},
    	Color:       0x00ff00, // Green - should change later based on message 
		Description: "This is a message from the api bot",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Type:",
				Value:  msg.Type,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Message: ",
				Value:  msg.Message,
				Inline: true,
			},
			// &discordgo.MessageEmbedField{
			// 	Name:   "Time: ",
			// 	Value:  msg.Time.String(),
			// 	Inline: true,
			// },
		},
		Footer: &discordgo.MessageEmbedFooter{
				Text: msg.Origin,
		},
	}	

	if msg.Suggestion == "" { //discord fills out embed with all fields even if they are empty in struct and omitempty
		_, err := s.ChannelMessageSendEmbed(channelID, embed_no_suggestion)
		return err
	} else{
		_, err := s.ChannelMessageSendEmbed(channelID, embed_full)
		return err
	}
}
