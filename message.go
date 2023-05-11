package main

import (
	"context"
	"fmt"
	github "github.com/google/go-github/v52/github"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
	"golang.org/x/oauth2"
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

func registerCommands(s *discordgo.Session, guildID string) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "welcome",
			Description: "Howdy! Welcome to Adomate! Let's get you started!",
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
			Description: "Check status of all services",
		},
		{
			Name:        "help",
			Description: "Get help with the bot",
		},
	}

	_, err := s.ApplicationCommandBulkOverwrite(os.Getenv("APP_ID"), guildID, commands)
	if err != nil {
		return err
	}

	return nil
}

func handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionMessageComponent {
		return
	}
	data := i.ApplicationCommandData()
	switch data.Name {
	case "welcome":
		/*
			call the updateRole function here
		*/
		err := getTeam(os.Getenv("GITHUB_ORG"))
		if err != nil {
			fmt.Println("Error:", err)
		}
		err2 := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Welcome to Adomate! Let's get you started!",
			},
		})
		if err2 != nil {
			fmt.Println("Error:", err)
		}
	case "api":

			status, err := getStatus(os.Getenv("API_URL"))
			if err != nil {
				fmt.Println("Error:", err)
			}
			var content string
			if status == "200 OK" {
				content = "API is operational." + "```" + "\nCode: " + status + "```" + "\n" + os.Getenv("API_URL")
			} else {
				content = "API is having issues." + "```" + "\nCode: " + status + "```" + "\n" + os.Getenv("API_URL")
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

			status, err := getStatus(os.Getenv("FRONTEND_URL"))
			if err != nil {
				fmt.Println("Error:", err)
			}
			var content string
			if status == "200 OK" {
				content = "Frontend is operational." + "```" + "\nCode: " + status + "```" + "\n" + os.Getenv("FRONTEND_URL")
			} else {
				content = "Frontend is having issues." + "```" + "\nCode: " + status + "```" + "\n" + os.Getenv("FRONTEND_URL")
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
		frontendStatus, err := getStatus(os.Getenv("FRONTEND_URL"))
		if err != nil {
			fmt.Println("Error:", err)
		}
		apiStatus, err := getStatus(os.Getenv("API_URL"))
		if err != nil {
			fmt.Println("Error:", err)
		}
		embed := &discordgo.MessageEmbed{
			Title:       "Adomate Status Dashboard - " + time.Now().Format("01-02-2006 15:04:05") + " CST",
			Description: "Adomate status information" + os.Getenv("ADOMATE_EMOJI_ID"),
			Color:       0x637EFE,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Service",
					Value:  "Frontend\nAPI\nBot",
					Inline: true,
				},
				{
					Name:   "Status",
					Value:  frontendStatus + "\n" + apiStatus + "\n200 OK",
					Inline: true,
				},
				{
					Name:   "URL",
					Value:  os.Getenv("FRONTEND_URL") + "\n" + os.Getenv("API_URL") + "\n" + os.Getenv("BOT_URL"),
					Inline: true,
				},
			},
		}
		embed.URL = os.Getenv("STATUS_URL")

		err2 := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
		if err2 != nil {
			fmt.Println("Error:", err)
		}
	case "help":
		embed := &discordgo.MessageEmbed{
			Title:       "Adomate Bot Help",
			Description: "Adomate Bot is a bot that helps with Adomate's Discord Server.",
			Color:       0x637EFE,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Commands",
					Value:  "1. /welcome\n2. /api\n3. /frontend\n4. /status\n5. /help",
					Inline: true,
				},
				{
					Name:   "Description",
					Value:  "Sends a welcome message.\nChecks the status of the API.\nChecks the status of the Frontend.\nStatus dashboard for all services.\nShows this help message.",
					Inline: true,
				},
				{
					Name:   "Usage",
					Value:  "#lobby channel\nDevelopers and Support\nDevelopers and Support\nAnyone\nAnyone",
					Inline: true,
				},
			},
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  1 << 6,
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
		if err != nil {
			fmt.Println("Error:", err)
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
	"type":"Error/Warning/Success/Log",
	"message":"Add Message",
	"suggestion":"Add Suggestion",
	"origin":"API",
	"time":"2023-04-24T08:45:26.371Z"
	}
*/

func sendDiscordMessage(s *discordgo.Session, channelID string, msg Message) error {
	unixTime := msg.Time.Unix()
	timestampStr := time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")
	embedFull := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{Name: "Adomate Discord Bot"},
		Color:       0x800000, // Maroon
		Description: fmt.Sprintf("%s message from %s", msg.Type, msg.Origin),
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

	switch msg.Type {
	case "Error":
		embedFull.Color = 0xFF0000 // Red
	case "Warning":
		embedFull.Color = 0xFFFF00 // Yellow
	case "Success":
		embedFull.Color = 0x00FF00 // Green
	default:
		embedFull.Color = 0xFFFFFF // White
	}
	if msg.Type == "Log" {
		_, err := s.ChannelMessageSend(channelID, " ["+timestampStr+"] "+msg.Message)
		return err
	} else {
		button := discordgo.Button{
			Label:    "Delete",
			Style:    discordgo.DangerButton,
			CustomID: "delete_message",
		}

		actionRow := discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{&button},
		}

		messageSendData := &discordgo.MessageSend{
			Embed: embedFull,
			Components: []discordgo.MessageComponent{
				&actionRow,
			},
		}

		_, err := s.ChannelMessageSendComplex(channelID, messageSendData)
		if err != nil {
			fmt.Println("Error sending message:", err)
		}
	}
	return nil
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

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionMessageComponent && i.MessageComponentData().CustomID == "delete_message" {
		err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
		if err != nil {
			fmt.Println("Error deleting message:", err)
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		if err != nil {
			fmt.Println("Error sending interaction response:", err)
		}
	}
}
// TODO: update role from github

func getTeam(orgName string) error {
	// Create an OAuth2 token source
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_PAT")},
	)

	// Create an OAuth2 HTTP client
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)

	// Create a GitHub client using the OAuth2 client
	client := github.NewClient(oauthClient)

	// Get the organization teams
	teams, _, err := client.Teams.ListTeams(context.Background(), orgName, nil)
	if err != nil {
		return fmt.Errorf("error retrieving teams: %v", err)
	}

	// Iterate over the teams and get their members
	for _, team := range teams {
		// Get the team members
		members, _, err := client.Teams.ListTeamMembersBySlug(context.Background(), orgName, *team.Slug, nil)
		if err != nil {
			return fmt.Errorf("error retrieving members for team %s: %v", *team.Name, err)
		}

		// Print the team name and its members
		fmt.Printf("Team: %s\n", *team.Name)
		for _, member := range members {
			fmt.Printf("Member: %s\n", *member.Login)
		}
		fmt.Println()
	}
	return nil
}