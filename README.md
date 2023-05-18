# discord-bot

## Introduction
The following is a guide for utilizing a Discord ChatBot that connects to the Adomate API. The guide provides instructions for installation, integration, and command usage.
## Installation

### Generate a DiscordBot Token

Create a bot using the Discord Developers Portal

`https://discord.com/developers/applications/`

Allow the bot to send and manage messages in the channel.

### Clone this repository

```git clone https://github.com/adomate-ads/discord-bot.git```


### Install rabbitMQ

```docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management```

Run the RabbitMQ container and sign-in with your credentials.

Run the go files from the cloned repository using `go run .`.

## Integration
To integrate the ChatBot with the Adomate API, update the .env file with the following parameters:


- DISCORD_BOT_TOKEN - Produced by the Discord Developers Portal
- CHANNEL_ID - Can be obtained from the server channel
- GUILD_ID - Server ID
- BETTERSTACK_TOKEN - BetterStack API to get status 
- RABBIT_HOST - localhost or url for the RabbitMQ server
- RABBIT_PORT - 15672 by default
- RABBIT_USER - RabbitMQ Username
- RABBIT_PASS - RabbitMQ Password
- RABBIT_DISCORD_QUEUE - RabbitMQ queue name
- BETTERSTACK_TOKEN - BetterStack API key to get status
- FRONTEND_ROLE_ID - Role ID for frontend team
- BACKEND_ROLE_ID - Role ID for backend team
- DISCORD_ROLE_ID - Role ID for discord team
- ERROR_CHANNEL_ID - Channel ID for error messages
- LOG_CHANNEL_ID - Channel ID for log messages
- GITHUB_PAT - Github Personal Access Token

## Commands
To Activate commands for the Adomate Bot, Send an `Activate` message on the channel.

 `/welcome` - Howdy! Provide your department name to get access to your assigned channels and features.
  
 `/status` - provides status for the bot 
 
 `/frontend` - provides status for adomate.ai
 
 `/api` - provides status for the Adomate API
 
> More Commands Coming Soon... 

## Message Format

### Message Types

Messages are categorized into four types:

`Error` - The API has encountered an error. Error event messages are marked with red embed.

`Warning` - The API has encountered a potential error. Proceed with caution! Warning event messages are marked with yellow embed.

`Log` - The API logged an event. Log event messages are marked with green embed.

### Title

The title provides a brief description of the event.
### Message

Provides the message content summary of the event.

### Origin

Provides the name of the event originator.

### Time

Provides the time when the event occured in the API.

## Message Template

RabbitMQ queue takes in the following message format to send messages to the discord bot.

```
{
"type":"Error",
"title":"API server down",
"message":"contact Adomate Support support@adomate.ai",
"origin":"API",
"time":"2023-03-10T11:45:26.371Z"
}
```

Where type can be `Error`, `Warning`, or `Log`. Title is a brief description of the event. Message is the content summary of the event. Origin is the name of the event originator. Time is the time when the event occured in the API.
