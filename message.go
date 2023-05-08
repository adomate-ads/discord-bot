package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
	"strings"
	"time"
)

type Department struct {
	Name        string `json:"name"`
	Message     string `json:"message"`
	Description string `json:"description"`
	Guidelines  string `json:"guidelines"`
	Emote       string `json:"emote"`
	RoleID      string `json:"role_id"`
}

type DepartmentsData struct {
	Departments []Department `json:"departments"`
	Default     Department   `json:"default"`
}

func parseDepartmentsJSON(filename string) (DepartmentsData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return DepartmentsData{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	var departmentsData DepartmentsData

	err = decoder.Decode(&departmentsData)
	if err != nil {
		return DepartmentsData{}, err
	}

	return departmentsData, nil
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
		department := strings.ToLower(data.Options[0].StringValue())
		departmentMsg := ""
		departmentDesc := ""
		roleEmote := ""
		guidelines := ""
		roleID := ""
		departmentsData, err := parseDepartmentsJSON("departments.json")
		if err != nil {
			fmt.Println("Error:", err)
		}
		for _, departmentData := range departmentsData.Departments {
			if departmentData.Name == department {
				departmentMsg = departmentData.Message
				departmentDesc = departmentData.Description
				guidelines = departmentData.Guidelines
				roleEmote = departmentData.Emote
				roleID = departmentData.RoleID
				break
			}
		}
		if departmentMsg == "" {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: "Howdy " + i.Member.User.Username + ", Please enter a valid department! or contact " + "<@&1104594618701590548>" + " for more information!",
				},
			})
			if err != nil {
				fmt.Println("Error:", err)
			}
		} else {
		err = updateRole(s, os.Getenv("GUILD_ID"), i.Member.User.ID, roleID)
		if err != nil {
			fmt.Println("Error:", err)
		}
		icebreaker := ""
		if len(data.Options) > 1 {
			icebreaker = data.Options[1].StringValue()
		}
		if icebreaker == "" {
			message := "Howdy @everyone! Welcome <@" + i.Member.User.ID + "> to Adomate!\nThey have joined the " + departmentMsg + " department!"
			_, err := s.ChannelMessageSend(os.Getenv("INTRO_CHANNEL_ID"), message)
			if err != nil {
				fmt.Println("Error:", err)
			}
		} else {
			message := "Howdy @everyone! Welcome <@" + i.Member.User.ID + "> to Adomate!\nFun Fact about " + i.Member.User.Username + ": " + icebreaker + "\nThey have joined the " + departmentMsg + " department!"
			_, err := s.ChannelMessageSend(os.Getenv("INTRO_CHANNEL_ID"), message)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}

		err2 := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6,
				Content: "Howdy " + i.Member.User.Username + "!\nWelcome to the ***" + departmentMsg + "*** Department!" + roleEmote + "\n\nMessage from the " + departmentMsg + " Department: \n" + departmentDesc + "\n\n" + guidelines + "\n\nHave a great time at Adomate!",
			},
		})
		if err2 != nil {
			fmt.Println("Error:", err)
		}
	}
	case "api":
		if !hasRequiredRole(s, os.Getenv("GUILD_ID"), i.Member.User.ID, "Developer", "Support") {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You do not have permission to use this command.",
				},
			}); err != nil {
				fmt.Println("Error responding to interaction:", err)
			}
		} else {
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
		}
	case "frontend":
		if !hasRequiredRole(s, os.Getenv("GUILD_ID"), i.Member.User.ID, "Developer", "Support") {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You do not have permission to use this command.",
				},
			}); err != nil {
				fmt.Println("Error responding to interaction:", err)
			}
			return
		} else {
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
			Description: "Adomate status information <:AdomateLogo:1104587690139205713>",
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

func updateRole(session *discordgo.Session, guildID string, userID string, roleID string) error {
	err := session.GuildMemberRoleAdd(guildID, userID, roleID)
	if err != nil {
		return err
	}
	return nil
}

func hasRequiredRole(s *discordgo.Session, guildID, userID string, roles ...string) bool {
	guildRoles, err := s.GuildRoles(guildID)
	if err != nil {
		fmt.Println("Error retrieving guild roles:", err)
		return false
	}

	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		fmt.Println("Error retrieving guild member information:", err)
		return false
	}
	for _, roleID := range member.Roles {
		for _, role := range guildRoles {
			if role.ID == roleID {
				for _, requiredRole := range roles {
					if role.Name == requiredRole {
						return true
					}
				}
			}
		}
	}
	return false
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	_, err := s.ChannelMessageSend(os.Getenv("CHANNEL_ID"), "Hello, @everyone! The bot is back online. Thank you for your patience! <:AdomateLogo:1104587690139205713>")
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func sendClosingMessage(s *discordgo.Session) {
	_, err := s.ChannelMessageSend(os.Getenv("CHANNEL_ID"), "Hello, @everyone! The bot is going down for maintenance. Please be patient while we work on it. Thank you! <:AdomateLogo:1104587690139205713>")
	if err != nil {
		fmt.Println("Error	:", err)
	}
}
