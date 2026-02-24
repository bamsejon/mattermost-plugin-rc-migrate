# Mattermost → Rocket.Chat Migration Plugin

A Mattermost server plugin that helps migrate channels from Mattermost to Rocket.Chat. It blocks new messages in migrated channels and redirects users to the corresponding Rocket.Chat channel.

## Features

- **`/migrate <url>`** — Mark a channel as migrated and point users to the Rocket.Chat URL
- **`/unmigrate`** — Remove the migration flag and restore normal channel operation
- **Message blocking** — Posts in migrated channels are rejected with an ephemeral redirect message
- **Channel header update** — The channel header is automatically set to show the redirect info
- **Configurable** — Base URL and redirect message can be set in System Console

## Installation

### Download and install

1. Download the plugin bundle (right-click → **Save Link As** to avoid browser auto-extraction):

   [**Download latest release (.tar.gz)**](https://github.com/bamsejon/mattermost-plugin-rc-migrate/releases/latest/download/se.bylund.mattermost-plugin-rc-migrate-1.1.0.tar.gz)

2. In Mattermost, go to **System Console → Plugins → Plugin Management**
3. Click **Upload Plugin** and select the `.tar.gz` file
4. Click **Enable**

> **Tip:** If the file picker doesn't show the file, make sure your browser didn't auto-extract it. The file must end in `.tar.gz`. You can also download it from the terminal:
> ```bash
> curl -LO https://github.com/bamsejon/mattermost-plugin-rc-migrate/releases/latest/download/se.bylund.mattermost-plugin-rc-migrate-1.1.0.tar.gz
> ```

### Build from source

Requires Go 1.21+:

```bash
cd mattermost-plugin-rc-migrate
go mod tidy
go mod vendor
make build
```

This produces `dist/se.bylund.mattermost-plugin-rc-migrate-1.1.0.tar.gz`.

## Configuration

In **System Console → Plugins → RC Migrate**:

| Setting | Description | Default |
|---------|-------------|---------|
| Rocket.Chat Base URL | Your Rocket.Chat instance URL (informational) | *(empty)* |
| Redirect Message | Message shown to users in migrated channels | `This channel has moved to Rocket.Chat` |
| API Secret | Shared secret for the REST API (required for external integrations) | *(empty — API disabled)* |

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

## REST API

The plugin exposes a REST API for external tools to trigger migrations. Requires the **API Secret** to be configured.

### POST `/plugins/se.bylund.mattermost-plugin-rc-migrate/api/v1/migrate`

```bash
curl -X POST https://mattermost.example.com/plugins/se.bylund.mattermost-plugin-rc-migrate/api/v1/migrate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-secret" \
  -d '{"channel_id": "abc123", "url": "https://rc.example.com/channel/general"}'
```

### POST `/plugins/se.bylund.mattermost-plugin-rc-migrate/api/v1/unmigrate`

```bash
curl -X POST https://mattermost.example.com/plugins/se.bylund.mattermost-plugin-rc-migrate/api/v1/unmigrate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-secret" \
  -d '{"channel_id": "abc123"}'
```

### Integration with Rocket.Chat Import App

The [Mattermost Import](https://github.com/bamsejon/rocketchat-mattermost-import) Rocket.Chat app (v2.7.0+) can automatically call this API after importing a channel. To set this up:

1. Configure an **API Secret** in the Mattermost plugin settings
2. In Rocket.Chat, go to the Mattermost Import app settings
3. Enable **Mattermost Redirect** and enter the same API secret
4. When you run `/importmattermost`, the source channel in Mattermost will automatically be redirected

## How it works

- Migration state is stored in the Mattermost KV store with key `migrated:<channelId>`
- The `MessageWillBePosted` hook intercepts messages and rejects them if the channel is migrated
- System messages and bot posts are not blocked
- An ephemeral message is sent to the user with the Rocket.Chat link

## License

MIT
