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

const (
	InitialBackOff = 5 * time.Second
	MaxBackOff     = 1 * time.Minute
)

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Queue    string
}

func connectRabbitMQ(RMQConfig RabbitMQConfig) (*amqp.Connection, *amqp.Channel, error) {
	config := amqp.Config{
		Heartbeat: 60 * time.Second,
	}

	conn, err := amqp.DialConfig(fmt.Sprintf("amqp://%s:%s@%s:%s/", RMQConfig.User, RMQConfig.Password, RMQConfig.Host, RMQConfig.Port), config)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

func handleReconnection(RMQConfig RabbitMQConfig) (*amqp.Connection, *amqp.Channel) {
	backoff := InitialBackOff

	for {
		conn, ch, err := connectRabbitMQ(RMQConfig)
		if err == nil {
			log.Println("Successfully connected to RabbitMQ!")
			return conn, ch
		}

		log.Printf("Failed to connect to RabbitMQ. Retrying in %v... Error: %v", backoff, err)
		time.Sleep(backoff)

		if backoff < MaxBackOff {
			backoff *= 2
		}
	}
}

func setupConsumer(ch *amqp.Channel, RMQConfig RabbitMQConfig, discord *discordgo.Session) (<-chan amqp.Delivery, error) {
	q, err := ch.QueueDeclare(
		RMQConfig.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return nil, err
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		for d := range msgs {
			var msg Message
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("Failed to parse messages: %v", err)
				continue
			}
			err = sendDiscordMessage(discord, msg)
			if err != nil {
				log.Printf("Failed to send message to Discord: %v", err)
				continue
			}
			err = d.Ack(false)
			if err != nil {
				log.Printf("Failed to acknowledge message: %v", err)
			}
		}
	}()

	return msgs, nil
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

	RMQConfig := RabbitMQConfig{
		Host:     os.Getenv("RABBIT_HOST"),
		Port:     os.Getenv("RABBIT_PORT"),
		User:     os.Getenv("RABBIT_USER"),
		Password: os.Getenv("RABBIT_PASS"),
		Queue:    os.Getenv("RABBIT_DISCORD_QUEUE"),
	}

	// Initial connection to RabbitMQ
	conn, ch := handleReconnection(RMQConfig)

	go func() {
		closeErrChan := conn.NotifyClose(make(chan *amqp.Error))
		for {
			<-closeErrChan
			log.Println("RabbitMQ connection closed. Attempting to reconnect...")
			conn, ch = handleReconnection(RMQConfig)
			closeErrChan = conn.NotifyClose(make(chan *amqp.Error))
			_, err := setupConsumer(ch, RMQConfig, discord)
			if err != nil {
				log.Printf("Failed to setup consumer after reconnection: %v", err)
			}
		}
	}()

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
				continue
			}
			err = sendDiscordMessage(discord, msg)
			if err != nil {
				log.Printf("Failed to send message to Discord: %v", err)
				continue
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
	if err := sendDiscordMessage(discord, Message{
		Type:    "Log",
		Title:   "Bot is now running",
		Message: "",
		Origin:  "Discord",
		Time:    time.Now(),
	}); err != nil {
		log.Fatal(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	// Cleanly close down the Discord session.
	discord.Close()
	conn.Close()
}
