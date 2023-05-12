package main

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	github "github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
	"net/http"
	"os"
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
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "github_username",
					Description: "Enter your GitHub username",
					Required:    true,
				},
			},
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
		err :=updateRole(s, i, data.Options[0].StringValue())
		if err != nil {
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
					Value:  "Sends a welcome message and assigns roles.\nChecks the status of the API.\nChecks the status of the Frontend.\nStatus dashboard for all services.\nShows this help message.",
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
			fmt.Println("Error:", err)
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

type Team struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

func getTeam(orgName string) ([]Team, error) {
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_PAT")},
	)

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)

	client := github.NewClient(oauthClient)

	teams, _, err := client.Teams.ListTeams(context.Background(), orgName, nil)
	if err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}

	var teamData []Team

	for _, team := range teams {
		teamMembers, _, err := client.Teams.ListTeamMembersBySlug(context.Background(), orgName, *team.Slug, nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving members for team %s: %v", *team.Name, err)
		}

		var members []string
		for _, member := range teamMembers {
			members = append(members, *member.Login)
		}

		teamInfo := Team{
			Name:    *team.Name,
			Members: members,
		}

		teamData = append(teamData, teamInfo)
	}

	return teamData, nil
}

func updateRole(s *discordgo.Session, i *discordgo.InteractionCreate, githubName string) error {
	teamData, err := getTeam(os.Getenv("GITHUB_ORG"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	var teamNames []string
	found := false

	for _, team := range teamData {
		for _, member := range team.Members {
			if member == githubName {
				teamNames = append(teamNames, team.Name)
				found = true
				break
			}
		}
	}

	if !found {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  1 << 6,
				Content: "Invalid GitHub username. Please try again with a valid GitHub username.",
			},
		})
		if err != nil {
			fmt.Println("Error:", err)
		}
		return nil
	} else {
		var roleIDs []string
		for _, teamName := range teamNames {
			switch teamName {
			case "frontend":
				roleIDs = append(roleIDs, os.Getenv("FRONTEND_ROLE_ID"))
			case "backend":
				roleIDs = append(roleIDs, os.Getenv("BACKEND_ROLE_ID"))
			case "discord":
				roleIDs = append(roleIDs, os.Getenv("DISCORD_ROLE_ID"))
			default:
			}
		}

		for _, roleID := range roleIDs {
			err := s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, roleID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return err
			}
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  1 << 6,
				Content: "Your roles have been updated successfully! Thank you for verifying your GitHub account.",
			},
		})
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
	}
return nil
}
