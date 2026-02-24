package main

import "fmt"

type configuration struct {
	RocketChatBaseURL string
	RedirectMessage   string
}

func (c *configuration) isValid() error {
	if c.RedirectMessage == "" {
		return fmt.Errorf("RedirectMessage must not be empty")
	}
	return nil
}

func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{
			RedirectMessage: "Denna kanal har flyttat till Rocket.Chat",
		}
	}
	return p.configuration
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
