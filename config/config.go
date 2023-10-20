package config

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config holds all settings to start the server
type Config struct {
	WebListenAddr string `toml:"web_listen_addr"`

	AdminToken string `toml:"admin_token"`

	Secred string `toml:"secred"`
}

// defaultConfig returns a config object with default values.
func defaultConfig() Config {
	return Config{
		WebListenAddr: ":8080",
		AdminToken:    CreatePassword(8),
		Secred:        CreatePassword(32),
	}
}

// LoadConfig loads the config from a toml file.
func LoadConfig(file string) (Config, error) {
	c := defaultConfig()

	f, err := os.Open(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If an error happens, return the error and the default config. The
			// caller can deside, if he wants to use the config even when the
			// default could not be saved.
			err := saveConfig(file, c)
			return c, err
		}
		return Config{}, fmt.Errorf("open config file: %w", err)
	}

	if err := toml.NewDecoder(f).Decode(&c); err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}
	return c, nil
}

func saveConfig(file string, config Config) (err error) {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if err := toml.NewEncoder(f).Encode(config); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// CreatePassword creates a random password string.
func CreatePassword(length int) string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
		"abcdefghijklmnopqrstuvwxyzåäö" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
