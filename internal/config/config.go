package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type AIConfig struct {
	Enabled      bool   `toml:"enabled"`
	Model        string `toml:"model"`
	VoiceProfile string `toml:"voice_profile"`
}

type BackupConfig struct {
	GitHubEnabled bool   `toml:"github_enabled"`
	GitHubRepo    string `toml:"github_repo"`
}

type PrivacyConfig struct {
	EncryptionEnabled bool `toml:"encryption_enabled"`
}

type NotificationsConfig struct {
	DailyNudge bool   `toml:"daily_nudge"`
	NudgeTime  string `toml:"nudge_time"`
}

type Config struct {
	JournalPath   string              `toml:"journal_path"`
	Editor        string              `toml:"editor"`
	AI            AIConfig            `toml:"ai"`
	Backup        BackupConfig        `toml:"backup"`
	Privacy       PrivacyConfig       `toml:"privacy"`
	Notifications NotificationsConfig `toml:"notifications"`

	// Runtime-only (not persisted)
	IsFirstRun bool   `toml:"-"`
	ConfigPath string `toml:"-"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		JournalPath: filepath.Join(homeDir, ".journal"),
		Editor:      "$EDITOR",
		AI: AIConfig{
			Enabled: true,
			Model:   "llama3",
			VoiceProfile: "Writes conversationally. Short sentences. Dry humor occasionally.\n" +
				"Doesn't over-explain emotions — states facts and lets them land.",
		},
		Backup: BackupConfig{
			GitHubEnabled: false,
			GitHubRepo:    "",
		},
		Privacy: PrivacyConfig{
			EncryptionEnabled: false,
		},
		Notifications: NotificationsConfig{
			DailyNudge: true,
			NudgeTime:  "21:00",
		},
	}
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".journal")
	configPath := filepath.Join(configDir, "config.toml")

	cfg := DefaultConfig()
	cfg.ConfigPath = configPath

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg.IsFirstRun = true
		return cfg, nil
	}

	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return nil, err
	}

	cfg.JournalPath = ExpandPath(cfg.JournalPath)
	cfg.IsFirstRun = false

	return cfg, nil
}

func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.ConfigPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(c.ConfigPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(c)
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[2:])
	}
	return path
}
