package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Queue    string
}

func main() {
	err := godotenv.Load(".env")
	if err != nil && os.Getenv("PROD") != "true" {
		log.Fatalf("Error loading .env file.")
	}

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		fmt.Println("Failed to connect to Discord")
		log.Fatal(err)
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	discord.AddHandler(messageCreate)
	discord.AddHandler(interactionCreate)
	discord.AddHandler(handleInteraction)
	err = registerCommands(discord, os.Getenv("GUILD_ID"))
	if err != nil {
		fmt.Println("Error registering commands: ", err)
	}

	// In this example, we only care about receiving message events.
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	RMQConfig := RabbitMQConfig{
		Host:     os.Getenv("RABBIT_HOST"),
		Port:     os.Getenv("RABBIT_PORT"),
		User:     os.Getenv("RABBIT_USER"),
		Password: os.Getenv("RABBIT_PASS"),
		Queue:    os.Getenv("RABBIT_DISCORD_QUEUE"),
	}

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", RMQConfig.User, RMQConfig.Password, RMQConfig.Host, RMQConfig.Port))
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ")
		fmt.Println(RMQConfig)
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to create Channel")
		log.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		RMQConfig.Queue, // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		fmt.Println("Failed to declare queue")
		log.Fatal(err)
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		fmt.Println("Failed to set QoS")
		log.Fatal(err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		fmt.Println("Failed to register a consumer")
		log.Fatal(err)
	}

	go func() {
		for d := range msgs {
			var msg Message
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("Failed to parse messages: %v", err)
			}
			err = sendDiscordMessage(discord, os.Getenv("CHANNEL_ID"), msg)
			if err != nil {
				log.Printf("Failed to send message to Discord: %v", err)
			}
			err = d.Ack(false)
			if err != nil {
				log.Printf("Failed to acknowledge message: %v", err)
			}
		}
	}()

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	go checkStatusForever(os.Getenv("FRONTEND_URL"), discord, os.Getenv("CHANNEL_ID"), os.Getenv("FRONTEND_ROLE_ID"))
	go checkStatusForever(os.Getenv("API_URL"), discord, os.Getenv("CHANNEL_ID"), os.Getenv("BACKEND_ROLE_ID"))
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	// Cleanly close down the Discord session.
	discord.Close()
}

func checkStatusForever(url string, discord *discordgo.Session, channelID string, roleID string) {
	isDown := false

	for {
		status, err := getStatus(url)
		if err != nil {
			log.Printf("Error: %v", err)
			isDown = true 
		} else {
			if status != "200 OK" {
				if isDown {
					embed := &discordgo.MessageEmbed{
						Title:       "Adomate Error Status",
						Description: "<@&"+ roleID + ">\nError Resolved! Check the status of the service below:",
						Color:       0x00ff00, // Green
						Timestamp:   time.Now().Format(time.RFC3339),
					}
					_, err := discord.ChannelMessageSendEmbed(channelID, embed)
					if err != nil {
						fmt.Println("Error:", err)
					}

					isDown = false
				}
			} else {
				if !isDown {
					embed := &discordgo.MessageEmbed{
						Title:       "Adomate Error Status",
						Description: "<@&"+ roleID + ">\nError Reported! Check the status of the service below:",
						Color:       0xff0000, // Red
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "URL",
								Value:  url,
								Inline: true,
							},
							{
								Name:   "Status",
								Value:  status,
								Inline: true,
							},
						},
						Timestamp: time.Now().Format(time.RFC3339),
					}
					_, err := discord.ChannelMessageSendEmbed(channelID, embed)
					if err != nil {
						fmt.Println("Error:", err)
					}

					isDown = true
				}
			}
		}

		time.Sleep(time.Second * 30)
	}
}
