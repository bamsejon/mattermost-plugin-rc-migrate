package main

import "fmt"

type configuration struct {
	RocketChatBaseURL string
	RedirectMessage   string
	APISecret         string
}

const defaultRedirectMessage = "This channel has moved to Rocket.Chat"

func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{
			RedirectMessage: defaultRedirectMessage,
		}
	}

	cfg := *p.configuration
	if cfg.RedirectMessage == "" {
		cfg.RedirectMessage = defaultRedirectMessage
	}
	return &cfg
}

func (p *Plugin) OnConfigurationChange() error {
	var cfg configuration
	if err := p.API.LoadPluginConfiguration(&cfg); err != nil {
		return fmt.Errorf("failed to load plugin configuration: %w", err)
	}

	p.configurationLock.Lock()
	p.configuration = &cfg
	p.configurationLock.Unlock()

	return nil
}
