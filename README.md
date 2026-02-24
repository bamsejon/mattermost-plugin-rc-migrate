# Mattermost → Rocket.Chat Migration Plugin

A Mattermost server plugin that helps migrate channels from Mattermost to Rocket.Chat. It blocks new messages in migrated channels and redirects users to the corresponding Rocket.Chat channel.

## Features

- **`/migrate <url>`** — Mark a channel as migrated and point users to the Rocket.Chat URL
- **`/unmigrate`** — Remove the migration flag and restore normal channel operation
- **Message blocking** — Posts in migrated channels are rejected with an ephemeral redirect message
- **Channel header update** — The channel header is automatically set to show the redirect info
- **Configurable** — Base URL and redirect message can be set in System Console

## Installation

### Prerequisites

- Mattermost Server 7.0+
- Go 1.21+ (for building)

### Build

```bash
cd mattermost-plugin-rc-migrate
go mod vendor
make build
```

This produces `dist/se.bylund.mattermost-plugin-rc-migrate-1.0.0.tar.gz`.

### Deploy

1. Go to **System Console → Plugins → Plugin Management**
2. Upload the `.tar.gz` file
3. Enable the plugin

## Configuration

In **System Console → Plugins → RC Migrate**:

| Setting | Description | Default |
|---------|-------------|---------|
| Rocket.Chat Base URL | Your Rocket.Chat instance URL (informational) | *(empty)* |
| Redirect Message | Message shown to users in migrated channels | `Denna kanal har flyttat till Rocket.Chat` |

## Usage

### Migrate a channel

In the channel you want to migrate, run:

```
/migrate https://rc.example.com/channel/general
```

This will:
1. Store the migration info in the plugin's KV store
2. Update the channel header with a redirect notice
3. Block all new user messages with an ephemeral redirect

### Unmigrate a channel

```
/unmigrate
```

This removes the migration flag and clears the channel header.

### Permissions

Only **channel admins** and **system admins** can run `/migrate` and `/unmigrate`.

## How it works

- Migration state is stored in the Mattermost KV store with key `migrated:<channelId>`
- The `MessageWillBePosted` hook intercepts messages and rejects them if the channel is migrated
- System messages and bot posts are not blocked
- An ephemeral message is sent to the user with the Rocket.Chat link

## License

MIT
