package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	kvPrefix = "migrated:"
)

type MigrationInfo struct {
	URL        string `json:"url"`
	MigratedBy string `json:"migratedBy"`
	MigratedAt int64  `json:"migratedAt"`
}

type Plugin struct {
	plugin.MattermostPlugin
	configurationLock sync.RWMutex
	configuration     *configuration
}

func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	if err := p.registerCommands(); err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	return nil
}

func (p *Plugin) MessageWillBePosted(_ *plugin.Context, post *model.Post) (*model.Post, string) {
	// Ignore system messages
	if post.Type != "" {
		return post, ""
	}

	// Ignore bot posts
	if post.GetProp("from_bot") == "true" {
		return post, ""
	}
	user, appErr := p.API.GetUser(post.UserId)
	if appErr != nil {
		return post, ""
	}
	if user.IsBot {
		return post, ""
	}

	// Check KV store for migration info
	data, appErr := p.API.KVGet(kvPrefix + post.ChannelId)
	if appErr != nil || data == nil {
		return post, ""
	}

	var info MigrationInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return post, ""
	}

	// Channel is migrated — send ephemeral message and reject the post
	cfg := p.getConfiguration()
	msg := fmt.Sprintf("%s: %s", cfg.RedirectMessage, info.URL)

	p.API.SendEphemeralPost(post.UserId, &model.Post{
		ChannelId: post.ChannelId,
		Message:   msg,
	})

	return nil, cfg.RedirectMessage
}
