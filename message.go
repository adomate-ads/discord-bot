package main

import (
	// "encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
	"strings"
	"time"
)
/*
	TODO Parse JSON file for departments
*/
type Department struct {
	Name        string `json:"name"`
	Message     string `json:"message"`
	Description string `json:"description"`
	Emote       string `json:"emote"`
}

type Config struct {
	Departments []Department `json:"departments"`
	Default     Department   `json:"default"`
}


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
					Name:        "department",
					Description: "What department are you in?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "icebreaker",
					Description: "Share an interesting fact about yourself!",
					Required:    false,
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
		departmentMsg := ""
		departmentDesc := ""
		roleEmote := ""

		switch department {
		case "engineering":
			departmentMsg = "Engineering"
			departmentDesc = "Engineering Department has access to all the engineering channels in the server. Please read the rules and regulations of the server before posting anything.\n**Important Guidelines** \n 1. Be respectful \n 2. Respect Privacy \n 3. Have Fun!\nIf you have any questions, please contact the @moderators.\nWe wish you a great time at Adomate!"
			roleEmote = "🛠📋💻"
		case "design":
			departmentMsg = "Design"
			departmentDesc = "Design Department has access to all the design channels in the server. Please read the rules and regulations of the server before posting anything. If you have any questions, please contact the moderators."
			roleEmote = "📋🎨🖌"
		case "marketing":
			departmentMsg = "Marketing"
			departmentDesc = "Marketing Department has access to all the marketing channels in the server. Please read the rules and regulations of the server before posting anything. If you have any questions, please contact the moderators."
			roleEmote = "📋📣📈"
		case "support":
			departmentMsg = "Support"
			departmentDesc = "Support Department has access to all the support channels in the server. Please read the rules and regulations of the server before posting anything. If you have any questions, please contact the moderators."
			roleEmote = "📋📞📧"
		case "hr":
			departmentMsg = "HR"
			departmentDesc = "HR Department has access to all the HR channels in the server. Please read the rules and regulations of the server before posting anything. If you have any questions, please contact the moderators."
			roleEmote = "📋📝📧"
		default:
			departmentMsg = "Adomate"
			departmentDesc = "Welcome to Adomate, to be placed in a department, please contact HR."
			roleEmote = "<Adomate Emoji ID>"
		}
		icebreaker := ""
		if len(data.Options) > 1 {
			icebreaker = data.Options[1].StringValue()
		}
		if icebreaker == "" {
			message := "Howdy @everyone! Welcome <@" + i.Member.User.ID + "> to Adomate!\nThey have joined the " + departmentMsg + " department!"
			_, err := s.ChannelMessageSend(os.Getenv("INTRO_CHANNEL_ID"), message)
			if err != nil {
				fmt.Println("Error sending introduction channel message:", err)
			}
		} else {
			message := "Howdy @everyone! Welcome <@" + i.Member.User.ID + "> to Adomate!\nFun Fact about " + i.Member.User.Username + ": " + icebreaker + "\nThey have joined the " + departmentMsg + " department!"
			_, err := s.ChannelMessageSend(os.Getenv("INTRO_CHANNEL_ID"), message)
			if err != nil {
				fmt.Println("Error sending introduction channel message:", err)
			}
		}
		
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Howdy " + i.Member.User.Username + "!\nWelcome to the ***" + departmentMsg + "*** Department!" + roleEmote + "\n\nMessage from the " + departmentMsg + " Department: \n" + departmentDesc,
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
			content = "Frontend is having issues.\nError Code: " + status
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
