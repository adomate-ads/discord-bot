package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
	"time"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) == 0 {
		return
	}

	_, err := s.ApplicationCommandBulkOverwrite(os.Getenv("APP_ID"), m.GuildID, []*discordgo.ApplicationCommand{
		{
			Name:        "Welcome",
			Description: "HOWDY",
		},
		{
			Name:        "api",
			Description: "Check API status",
		},
		{
			Name:        "frontend",
			Description: "Check Frontend status",
		},
		{
			Name:        "status",
			Description: "Check Bot status",
		},
	})
	if err != nil {
		// Handle the error
		fmt.Println("Error:", err)
	}
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.ApplicationCommandData()
		switch data.Name {
		case "Welcome":
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Howdy!",
				},
			})
			if err != nil {
				fmt.Println("Error:", err)
			}
		case "api":
			status, err := getStatus("https://api.adomate.com/api/v1/")
			if err != nil {
				fmt.Println("Error:", err)
			}
			var content string
			if status == "200 OK" {
				content = "API is operational. " + status
			} else {
				content = "API is having issues. " + status
			}

			err2 := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
			if err2 != nil {
				fmt.Println("Error:", err)
			}
		case "frontend":
			status, err := getStatus("https://api.adomate.com/api/v1/")
			if err != nil {
				fmt.Println("Error:", err)
			}
			var content string
			if status == "200 OK" {
				content = "API is operational. " + status
			} else {
				content = "API is having issues. " + status
			}

			err2 := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
			if err2 != nil {
				fmt.Println("Error:", err)
			}
		case "status":
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Adomate Bot is operational.",
				},
			})
			if err != nil {
				fmt.Println("Error:", err)
			}
		default:
			fmt.Println("Unknown command")
		}
	})
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

// func checkAndAddReaction(s *discordgo.Session, m *discordgo.Message, reaction string) error {
// 	for _, r := range m.Reactions {
// 		if r.Emoji.Name == reaction && r.Count > 0 {
// 			return fmt.Errorf("message already processed")
// 		}
// 	}

// 	err := s.MessageReactionAdd(m.ChannelID, m.ID, reaction)
// 	return err
// }
