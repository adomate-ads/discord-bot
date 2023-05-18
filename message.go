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
			Name:        "status",
			Description: "Check status of all services",
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
		err := updateRole(s, i, data.Options[0].StringValue())
		if err != nil {
			fmt.Println("Error:", err)
		}

	case "status":
		frontendStatus, err := getStatus("https://www.adomate.ai/")
		if err != nil {
			fmt.Println("Error:", err)
		}
		apiStatus, err := getStatus("https://api.adomate.ai/v1/")
		if err != nil {
			fmt.Println("Error:", err)
		}
		embed := &discordgo.MessageEmbed{
			Title:       "Adomate Status Dashboard - " + time.Now().Local().Format("01/02/2006 03:04:05 PM (MST)"),
			Description: "Adomate status information <:logo:1106617488533372998>",
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
					Value:  "https://www.adomate.ai/" + "\n" + "https://api.adomate.ai/v1/" + "\n" + os.Getenv("BOT_URL"),
					Inline: true,
				},
			},
		}
		embed.URL = "https://status.adomate.ai/"

		err2 := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
		if err2 != nil {
			fmt.Println("Error:", err)
		}
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

func sendDiscordMessage(s *discordgo.Session, channelID string, msg Message) error {
	unixTime := msg.Time.Unix()
	timestampStr := time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")
	embedFull := &discordgo.MessageEmbed{
		Color:       0x800000, // Maroon
		Title:       "Adomate " + msg.Type + " Message from " + msg.Origin,
		Description: fmt.Sprintf("> " + msg.Title + "\n> " + msg.Message),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "⏰ " + timestampStr,
		},
	}
	if msg.Message == "" {
		embedFull.Description = fmt.Sprintf("> " + msg.Title)
	}
	switch msg.Type {
	case "Error":
		embedFull.Color = 0xFF0000 // Red
	case "Warning":
		embedFull.Color = 0xFFFF00 // Yellow
	case "Success":
		embedFull.Color = 0x00FF00 // Green
	case "Log":
		embedFull.Color = 0x637EFE // Adomate Purple
	default:
		embedFull.Color = 0xFFFFFF // White
	}
	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{Embed: embedFull})
	if err != nil {
		return err
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
	teamData, err := getTeam("adomate-ads")
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
				Flags:   1 << 6,
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
				roleIDs = append(roleIDs, os.Getenv("MEMBER_ROLE_ID"))
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
				Flags:   1 << 6,
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
