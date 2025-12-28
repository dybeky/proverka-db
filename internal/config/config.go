package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DiscordBotToken  string `json:"discord_bot_token"`
	DiscordChannelID string `json:"discord_channel_id"`
}

var cfg *Config

func Load() error {
	// Try to load embedded config first
	data, err := decryptConfig()
	if err != nil || len(data) == 0 {
		// Fallback to external config.json for development
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}

		configPath := filepath.Join(filepath.Dir(exePath), "config.json")
		data, err = os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config (neither embedded nor config.json found): %w", err)
		}
	}

	cfg = &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.DiscordBotToken == "" || cfg.DiscordBotToken == "YOUR_BOT_TOKEN_HERE" {
		return fmt.Errorf("discord bot token not configured")
	}

	if cfg.DiscordChannelID == "" || cfg.DiscordChannelID == "YOUR_CHANNEL_ID_HERE" {
		return fmt.Errorf("discord channel ID not configured")
	}

	return nil
}

func Get() *Config {
	return cfg
}
