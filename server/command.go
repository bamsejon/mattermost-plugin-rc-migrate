package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	commandMigrate   = "migrate"
	commandUnmigrate = "unmigrate"
)

func (p *Plugin) registerCommands() error {
	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          commandMigrate,
		AutoComplete:     true,
		AutoCompleteHint: "<rocket.chat-url>",
		AutoCompleteDesc: "Mark this channel as migrated to Rocket.Chat",
	}); err != nil {
		return err
	}

	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          commandUnmigrate,
		AutoComplete:     true,
		AutoCompleteHint: "",
		AutoCompleteDesc: "Remove migration flag from this channel",
	}); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	trigger := strings.TrimPrefix(strings.Fields(args.Command)[0], "/")

	switch trigger {
	case commandMigrate:
		return p.executeMigrate(args)
	case commandUnmigrate:
		return p.executeUnmigrate(args)
	default:
		return &model.CommandResponse{
			Text: fmt.Sprintf("Unknown command: %s", trigger),
		}, nil
	}
}

func (p *Plugin) executeMigrate(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if !p.hasAdminPermission(args) {
		return ephemeralResponse("You need to be a channel admin or system admin to use this command."), nil
	}

	// Parse the URL from the command args
	parts := strings.Fields(args.Command)
	if len(parts) < 2 {
		return ephemeralResponse("Usage: `/migrate <rocket.chat-url>`"), nil
	}
	rcURL := parts[1]

	if !strings.HasPrefix(rcURL, "http://") && !strings.HasPrefix(rcURL, "https://") {
		return ephemeralResponse("URL must start with http:// or https://"), nil
	}

	// Store migration info
	info := MigrationInfo{
		URL:        rcURL,
		MigratedBy: args.UserId,
		MigratedAt: time.Now().Unix(),
	}
	data, err := json.Marshal(info)
	if err != nil {
		return ephemeralResponse("Failed to marshal migration info."), nil
	}

	if appErr := p.API.KVSet(kvPrefix+args.ChannelId, data); appErr != nil {
		return ephemeralResponse("Failed to save migration info: " + appErr.Error()), nil
	}

	// Update channel header
	cfg := p.getConfiguration()
	headerText := fmt.Sprintf(":no_entry: %s: %s", cfg.RedirectMessage, rcURL)

	channel, appErr := p.API.GetChannel(args.ChannelId)
	if appErr != nil {
		return ephemeralResponse("Failed to get channel: " + appErr.Error()), nil
	}

	channel.Header = headerText
	if _, appErr := p.API.UpdateChannel(channel); appErr != nil {
		return ephemeralResponse("Migration saved but failed to update channel header: " + appErr.Error()), nil
	}

	return &model.CommandResponse{
		Text: fmt.Sprintf("Channel migrated to Rocket.Chat: %s\nNew messages will be blocked with a redirect.", rcURL),
	}, nil
}

func (p *Plugin) executeUnmigrate(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if !p.hasAdminPermission(args) {
		return ephemeralResponse("You need to be a channel admin or system admin to use this command."), nil
	}

	// Check if channel is actually migrated
	data, appErr := p.API.KVGet(kvPrefix + args.ChannelId)
	if appErr != nil || data == nil {
		return ephemeralResponse("This channel is not marked as migrated."), nil
	}

	// Remove migration info
	if appErr := p.API.KVDelete(kvPrefix + args.ChannelId); appErr != nil {
		return ephemeralResponse("Failed to remove migration info: " + appErr.Error()), nil
	}

	// Clear channel header
	channel, appErr := p.API.GetChannel(args.ChannelId)
	if appErr != nil {
		return ephemeralResponse("Migration removed but failed to get channel: " + appErr.Error()), nil
	}

	channel.Header = ""
	if _, appErr := p.API.UpdateChannel(channel); appErr != nil {
		return ephemeralResponse("Migration removed but failed to clear channel header: " + appErr.Error()), nil
	}

	return &model.CommandResponse{
		Text: "Migration flag removed. This channel is now active again.",
	}, nil
}

func (p *Plugin) hasAdminPermission(args *model.CommandArgs) bool {
	// Check system admin
	user, appErr := p.API.GetUser(args.UserId)
	if appErr != nil {
		return false
	}
	if user.IsSystemAdmin() {
		return true
	}

	// Check channel admin
	member, appErr := p.API.GetChannelMember(args.ChannelId, args.UserId)
	if appErr != nil {
		return false
	}
	return member.SchemeAdmin
}

func ephemeralResponse(msg string) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         msg,
	}
}
