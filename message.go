package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
	"time"
	"strings"
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
			Description: "HOWDY!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "department",
					Description: "What department are you in?",
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
			Description: "Check Bot status",
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
		department := data.Options[0].StringValue()
		department = strings.ToLower(department)
		departmentmsg := "Adomate"  // default
		departmentdesc := "Adomate" // default
		switch department {
		case "engineering":
			departmentmsg = "Engineering"
			departmentdesc = "Head of Department: Engineering Department has access to all the engineering channels in the server. Please read the rules and regulations of the server before posting anything. For starters, please introduce yourself in the `#introductions` channel.\n**Important Guidelines** \n 1. Be respectful \n 2. Respect Privacy \n 3. Have Fun!\nIf you have any questions, please contact the moderators.\nWe wish you a great time at Adomate!"
		case "design":
			departmentmsg = "Design"
			departmentdesc = "Design Department has access to all the design channels in the server. Please read the rules and regulations of the server before posting anything. If you have any questions, please contact the moderators."
		case "marketing":
			departmentmsg = "Marketing"
			departmentdesc = "Marketing Department has access to all the marketing channels in the server. Please read the rules and regulations of the server before posting anything. If you have any questions, please contact the moderators."
		case "support":
			departmentmsg = "Support"
			departmentdesc = "Support Department has access to all the support channels in the server. Please read the rules and regulations of the server before posting anything. If you have any questions, please contact the moderators."
		case "hr":
			departmentmsg = "HR"
			departmentdesc = "HR Department has access to all the HR channels in the server. Please read the rules and regulations of the server before posting anything. If you have any questions, please contact the moderators."
		default:
			departmentmsg = "Adomate"
			departmentdesc = "Welcome to Adomate, to be placed in a department, please contact HR."
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Howdy " + i.Member.User.Username + "!\nWelcome to the ***" + departmentmsg + "*** Department! 🥳\n\nMessage from the " + departmentmsg + " Department: \n" + departmentdesc,
			},
		})
		if err != nil {
			fmt.Println("Error:", err)
		}
	case "api":
		status, err := getStatus("https://api.adomate.ai/v1/")
		if err != nil {
			fmt.Println("Error:", err)
		}
		var content string
		if status == "200 OK" {
			content = "API is operational.\nCode: " + status
		} else {
			content = "API is having issues.\nError Code: " + status
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
		status, err := getStatus("https://www.adomate.ai/")
		if err != nil {
			fmt.Println("Error:", err)
		}
		var content string
		if status == "200 OK" {
			content = "Frontend is operational.\nCode: " + status
		} else {
			content = "Frointend is having issues.\nError Code: " + status
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
		Color:       0x800000, // Maroon - should change later based on message
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

func updateRole(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	guildID := m.GuildID
	memberID := m.User.ID

	// Get the role ID you want to assign to the user
	roleID := "your_role_id"

	// Get the guild roles
	guildRoles, err := s.GuildRoles(guildID)
	if err != nil {
		fmt.Println("Error retrieving guild roles:", err)
		return
	}

	// Find the role by ID
	var role *discordgo.Role
	for _, r := range guildRoles {
		if r.ID == roleID {
			role = r
			break
		}
	}

	// Check if the role exists
	if role == nil {
		fmt.Println("Role not found")
		return
	}

	// Check if the member has the role already
	hasRole := false
	for _, r := range m.Roles {
		if r == roleID {
			hasRole = true
			break
		}
	}

	// If the member doesn't have the role, add it
	if !hasRole {
		err = s.GuildMemberRoleAdd(guildID, memberID, roleID)
		if err != nil {
			fmt.Println("Error adding role to member:", err)
			return
		}

		fmt.Println("Role added to member")
	} else {
		fmt.Println("Member already has the role")
	}
}
