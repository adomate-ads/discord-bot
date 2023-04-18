package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
	"strings"
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

	if strings.HasPrefix(m.Content, "!") {
		err := checkAndAddReaction(s, m.Message, "🤖")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		command := strings.TrimPrefix(m.Content, "!")

		switch command {
		case "status":
			_, err := s.ChannelMessageSend(m.ChannelID, "Adomate Bot is operational.")
			if err != nil {
				fmt.Println("Error:", err)
			}
		case "frontend":
			status, err := getStatus("https://betteruptime.com/api/v2/monitors?url=https://www.adomate.ai")
			switch status {
			case "200 OK":
				_, err := s.ChannelMessageSend(m.ChannelID, "Frontend is operational.")
				if err != nil {
					fmt.Println("Error:", err)
				}
			default:
				_, err := s.ChannelMessageSend(m.ChannelID, "Frontend is having issues.")
				if err != nil {
					fmt.Println("Error:", err)
				}
			}
			s.ChannelMessageSend(m.ChannelID, status)
			if err != nil {
				fmt.Println("Error:", err)
			}
		case "api":
			status, err := getStatus("https://betteruptime.com/api/v2/monitors?url=https://api.adomate.ai/v1/")
			switch status {
			case "200 OK":
				_, err := s.ChannelMessageSend(m.ChannelID, "API is operational.")
				if err != nil {
					fmt.Println("Error:", err)
				}
			default:
				_, err := s.ChannelMessageSend(m.ChannelID, "API is having issues.")
				if err != nil {
					fmt.Println("Error:", err)
				}
			}
			s.ChannelMessageSend(m.ChannelID, status)
			if err != nil {
				fmt.Println("Error:", err)
			}
		default:
			_, err := s.ChannelMessageSend(m.ChannelID, "Invalid command Available commands: !status, !frontend, !api")
			if err != nil {
				fmt.Println("Error:", err)
			}
		}
	}
}

type Message struct {
	Type       string    `json:"type" example:"error/warning/log"`
	Message    string    `json:"message"`
	Suggestion string    `json:"suggestion,omitempty"`
	Time       time.Time `json:"time,omitempty" example:"2018-12-12T11:45:26.371Z"`
	Origin     string    `json:"origin" example:"api/gac"`
}

/*
	message example
	{
	"type":"Error",
	"message":"test",
	"suggestion":"lol",
	"origin":"api",
	"time":"2018-12-12T11:45:26.371Z"
	}
*/

func sendDiscordMessage(s *discordgo.Session, channelID string, msg Message) error {

	unixTime := msg.Time.Unix()
	timestampStr := time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")
	embedFull := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{Name: "Adomate Discord Bot"},
		Color:       0x800000, // Maroon - should change later based on message
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
			Text: fmt.Sprintf("Time: %s", timestampStr),
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
	default:
		embedFull.Color = 0xFFFFFF
	}
	if msg.Type == "Log" {
		_, err := s.ChannelMessageSend(channelID, " ["+timestampStr+"] "+msg.Message)
		return err
	} else {
		message := &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				embedFull,
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Delete Message",
							Style:    discordgo.DangerButton,
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
					fmt.Println("Error occurred during deletion:", err)
				}
			}
		})
		return err
	}
}

func getStatus(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("BETTERSTACK_TOKEN"))
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	return resp.Status, nil
}

func checkAndAddReaction(s *discordgo.Session, m *discordgo.Message, reaction string) error {
	for _, r := range m.Reactions {
		if r.Emoji.Name == reaction && r.Count > 0 {
			return fmt.Errorf("message already processed")
		}
	}

	err := s.MessageReactionAdd(m.ChannelID, m.ID, reaction)
	return err
}
