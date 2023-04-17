package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"hash/fnv"
	"log"
	"os"
	"strings"
	"time"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := "!"

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) == 0 {
		return
	}

	if strings.HasPrefix(m.Content, prefix) {
		podName := os.Getenv("POD_NAME")
		emoji := generateInstanceEmoji(podName)

		err := s.MessageReactionAdd(m.ChannelID, m.ID, emoji)
		if err != nil {
			log.Println("Error adding reaction:", err)
			return
		}

		time.Sleep(1 * time.Second) // Wait for other instance's reaction

		reactions, err := s.MessageReactions(m.ChannelID, m.ID, emoji, 2, "", "")
		if err != nil {
			log.Println("Error fetching reactions:", err)
			return
		}

		if len(reactions) > 0 && reactions[0].ID == s.State.User.ID {
			command := strings.TrimPrefix(m.Content, prefix)

			switch command {
			case "status":
				_, err := s.ChannelMessageSend(m.ChannelID, "I'm alive!")
				if err != nil {
					fmt.Println("Error:", err)
				}
			case "isdown":
				_, err := s.ChannelMessageSend(m.ChannelID, "All services are operational.")
				if err != nil {
					fmt.Println("Error:", err)
				}
			default:
				_, err := s.ChannelMessageSend(m.ChannelID, "Invalid command.")
				if err != nil {
					fmt.Println("Error:", err)
				}
			}
		} else {
			// Remove the bot's own reaction if it didn't respond
			_ = s.MessageReactionRemove(m.ChannelID, m.ID, emoji, s.State.User.ID)
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
			Text: fmt.Sprintf("Time: %s", msg.Time.Format("02 Jan 06 15:04:01 CDT")),
		},
	}

	// switch embed color depending on msg type
	switch msg.Type {
	case "Error":
		embedFull.Color = 0xFF0000
	case "Warn":
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

func generateInstanceEmoji(podName string) string {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(podName))

	emojis := []string{
		"🟥",
		"🟦",
		"🟩",
		"🟨",
		"🟧",
		"🟪",
		"🟫",
	}

	return emojis[hash.Sum32()%uint32(len(emojis))]
}
