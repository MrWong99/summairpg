package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/itzg/go-flagsfiller"
)

// ConfigFile to load or safe for App configs.
const ConfigFile = "summairpg-config.json"

// App is the configuration needed for the application.
type App struct {
	// Config contains the settings for the configuration itself.
	Config Config `json:"-"`
	// Audio are just the settings for the input audio files.
	Audio Audio `json:"audio"`
	// Ollama settings for summarizing the transcriptions.
	Ollama Ollama `json:"ollama"`
}

type Config struct {
	// Store is true if the loaded configuration should be created/updated in the ConfigFile.
	Store bool `json:"-" default:"true" usage:"Store the provided command-line arguments in the summairpg-config.json"`
}

// Audio are just the settings for the input audio files.
type Audio struct {
	// Dir is the directory that contains all of the audio files that should be transcribed.
	Dir string `json:"dir" default:"input" usage:"The directory that contains all of the audio files that should be transcribed"`
	// Language that the spoken chat is in.
	Language string `json:"language" default:"en" usage:"The spoken language in the audio files"`
	// FileTypes are the file extensions that should be considered when looking up audio tracks. They should be a comma-separated list.
	FileTypes string `json:"file-types" default:"flac,wav" usage:"the file extensions that should be considered when looking up audio tracks. They should be a comma-separated list"`
	// Model to use. See https://ollama.com/library
	Model string `json:"model" default:"large-v3" usage:"WhisperX model to use. See https://huggingface.co/models?sort=trending&search=whisper"`
	// DisplayTranscript can be true to print the entire transcription to console.
	DisplayTranscript bool `json:"display-transcript" default:"false" usage:"can be set to true to print the entire transcription to console"`
}

// Ollama settings for summarizing the transcriptions.
type Ollama struct {
	// Address is the host:port of the Ollama HTTP API.
	Address string `json:"address" default:"127.0.0.1:11434" usage:"The host:port of the Ollama HTTP API."`
	// Model to use. See https://ollama.com/library
	Model string `json:"model" default:"llama3:70b" usage:"Ollama model to use. See https://ollama.com/library"`
	// UpdateModel if the model should be updated or pulled before use.
	UpdateModel bool `json:"update-model" default:"true" usage:"set to false to disable pulling the latest version of the model"`
}

// Init returns the App config that uses both flags and the config file as input. Flags will override configurations provided in the summairpg-config.json.
// Also the summairpg-config.json will automatically be created/updated if App.Config.Store is set to true.
func Init() (*App, error) {
	var config App
	flagFiller := flagsfiller.New()
	if err := flagFiller.Fill(flag.CommandLine, &config); err != nil {
		return nil, fmt.Errorf("could not prepare command-line flags: %w", err)
	}
	if err := overrideDefaultsFromConfig(); err != nil {
		return nil, fmt.Errorf("could not read config file %q: %w", ConfigFile, err)
	}
	flag.Parse()
	if !config.Config.Store {
		return &config, nil
	}
	if err := UpdateStored(&config); err != nil {
		slog.Warn("could not create/update config file", "file", ConfigFile, "error", err)
	}
	return &config, nil
}

// UpdateStored App config in the summairpg-config.json file with 0644 permissions.
// The file will be truncated if it already exists.
func UpdateStored(config *App) error {
	cfgFile, err := os.OpenFile(ConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer cfgFile.Close()
	enc := json.NewEncoder(cfgFile)
	enc.SetIndent("", "  ")
	return enc.Encode(config)
}

func overrideDefaultsFromConfig() error {
	cfgFile, err := os.Open(ConfigFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer cfgFile.Close()

	var config App
	if err = json.NewDecoder(cfgFile).Decode(&config); err != nil {
		return err
	}
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		switch f.Name {
		case "audio-dir":
			f.Value.Set(config.Audio.Dir)
		case "audio-language":
			f.Value.Set(config.Audio.Language)
		case "audio-file-types":
			f.Value.Set(config.Audio.FileTypes)
		case "audio-model":
			f.Value.Set(config.Audio.Model)
		case "audio-display-transcript":
			f.Value.Set(strconv.FormatBool(config.Audio.DisplayTranscript))
		case "ollama-address":
			f.Value.Set(config.Ollama.Address)
		case "ollama-model":
			f.Value.Set(config.Ollama.Model)
		case "ollama-update-model":
			f.Value.Set(strconv.FormatBool(config.Ollama.UpdateModel))
		}
	})
	return nil
}
