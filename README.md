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

Run the go files from the cloned repository using `go run all`.

## Integration
To integrate the ChatBot with the Adomate API, update the .env file with the following parameters:


- DISCORD_BOT_TOKEN - Produced my the Discord Developers Portal
- CHANNEL_ID - Can be obtained from the server channel
- RABBIT_HOST - localhost or url for the RabbitMQ server
- RABBIT_PORT - 15672 by default
- RABBIT_USER - RabbitMQ Username
- RABBIT_PASS - RabbitMQ Password
- RABBIT_DISCORD_QUEUE - RabbitMQ queue name


## Commands
 `!status` - provides status for the bot 

 `!isdown` - provides status for the Adomate API
 
> More Commands Coming Soon...

## Message Format

### Message Types

Messages are categorized into four types:

`Error` - The API has encountered an error. Error event messages are marked with red embed.

`Warning` - The API has encountered a potential error. Proceed with caution! Warning event messages are marked with yellow embed.

`Log` - The API logged an event. Log event messages are marked with green embed.

`General` - General event messages. General event messages are marked with white embed.

### Suggestions

The Suggestion message is sent by the API with a suggestion to fix errors or warnings. 

### Origin

Provides the name of the event originator.

### Time

Provides the time when the event occured in the API.

### Delete Message button

The message contains a delete message button that allows the client to  delete the message sent from the API after the conflict/errors are resolved.

## Message Template

RabbitMQ queue takes in the following message format to send messages to the discord bot.

```
{
"type":"Error",
"message":"API server down",
"suggestion":"contact AdomateHelpDesk",
"origin":"API",
"time":"2023-03-10T11:45:26.371Z"
}
```

Where type is the message type, message provides the message content, suggestion provides suggestions to fix errors or warnings, origin provides the name of the event originator, and time indicates the time when the event occurred in the API.

### Changelog
> Coming Soon...
