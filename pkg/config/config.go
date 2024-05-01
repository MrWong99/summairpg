package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/itzg/go-flagsfiller"
	"github.com/sashabaranov/go-openai"
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
	// OpenAI settings for summarizing the transcriptions.
	OpenAI OpenAI `json:"openai" env:"openai"`
}

type Config struct {
	// Store is true if the loaded configuration should be created/updated in the ConfigFile.
	Store bool `json:"-" default:"true" usage:"Store the provided command-line arguments in the summairpg-config.json"`
}

// Audio are just the settings for the input audio files.
type Audio struct {
	// TranscriptFile will be used as the audio transcript if set. This will skip the execution of WhisperX entirely.
	TranscriptFile string `json:"transcript-file" default:"" usage:"when set the entire transcription will be skipped and this files content will be used as summarization input"`
	// Dir is the directory that contains all of the audio files that should be transcribed.
	Dir string `json:"dir" default:"input" usage:"The directory that contains all of the audio files that should be transcribed"`
	// Language that the spoken chat is in.
	Language string `json:"language" default:"en" usage:"The spoken language in the audio files"`
	// FileTypes are the file extensions that should be considered when looking up audio tracks. They should be a comma-separated list.
	FileTypes []string `json:"file-types" default:"flac,wav" usage:"the file extensions that should be considered when looking up audio tracks. They should be a comma-separated list"`
	// Model to use. See https://ollama.com/library
	Model string `json:"model" default:"large-v3" usage:"WhisperX model to use. See https://huggingface.co/models?sort=trending&search=whisper"`
	// DisplayTranscript can be true to print the entire transcription to console.
	DisplayTranscript bool `json:"display-transcript" default:"false" usage:"can be set to true to print the entire transcription to console"`
}

// Ollama settings for summarizing the transcriptions.
type Ollama struct {
	// Enabled if the Ollama endpoint should be used for summarization.
	Enabled bool `json:"enabled" default:"true" usage:"set to false to disable the Ollama endpoint for summarization"`
	// Address is the host:port of the Ollama HTTP API.
	Address string `json:"address" default:"127.0.0.1:11434" usage:"The host:port of the Ollama HTTP API."`
	// Model to use. See https://ollama.com/library
	Model string `json:"model" default:"llama3:70b" usage:"Ollama model to use. See https://ollama.com/library"`
	// ContextLengthOverride will override the context length (num_ctx) that would else be determined by the model.
	ContextLengthOverride int `json:"content-length-override" default:"0" usage:"override the context length (num_ctx) that would else be determined by the model"`
	// UpdateModel if the model should be updated or pulled before use.
	UpdateModel bool `json:"update-model" default:"true" usage:"set to false to disable pulling the latest version of the model"`
}

// OpenAI settings for summarizing the transcriptions.
type OpenAI struct {
	// Enabled if the OpenAI chat endpoint should be used for summarization.
	Enabled bool `json:"enabled" aliases:"openai-enabled" default:"false" usage:"set to true to enable the OpenAI endpoint for summarization"`
	// Url is the base url of the OpenAI API endpoint to use. Usually in the format https://host[:port]/v1.
	Url string `json:"url" aliases:"openai-url" default:"https://api.openai.com/v1" usage:"the base url of the OpenAI API endpoint to use"`
	// Model of the OpenAI API to use.
	Model string `json:"model" aliases:"openai-model" default:"gpt-4-turbo" usage:"the OpenAI model to use. See https://platform.openai.com/docs/models/model-endpoint-compatibility"`
	// OrgId optional HTTP header to set.
	OrgId string `json:"org-id" aliases:"openai-org-id" default:"" usage:"will set the OrgID as HTTP header"`
	// ApiType to use. See openai.APIType
	ApiType openai.APIType `json:"api-type" aliases:"openai-api-type" default:"OPEN_AI" usage:"the type of OpenAI endpoint. Must be one of OPEN_AI, AZURE or AZURE_AD"`
	// ApiVersion to use. Only required if ApiType is AZURE or AZURE_AD.
	ApiVersion string `json:"api-version" aliases:"openai-api-version" default:"" usage:"the version of the Azure API to use. Not required when openai-api-type is OPEN_AI"`
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
	if config.Ollama.Enabled && config.OpenAI.Enabled {
		return &config, errors.New("you must not enable both Ollama and OpenAI")
	}
	if config.OpenAI.Enabled {
		if _, ok := os.LookupEnv("OPENAI_API_KEY"); !ok {
			return &config, errors.New("when using the OpenAI API you must set the environment variable OPENAI_API_KEY")
		}
	}
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
		case "audio-transcript-file":
			f.Value.Set(config.Audio.TranscriptFile)
		case "audio-dir":
			f.Value.Set(config.Audio.Dir)
		case "audio-language":
			f.Value.Set(config.Audio.Language)
		case "audio-file-types":
			f.Value.Set(strings.Join(config.Audio.FileTypes, ","))
		case "audio-model":
			f.Value.Set(config.Audio.Model)
		case "audio-display-transcript":
			f.Value.Set(strconv.FormatBool(config.Audio.DisplayTranscript))
		case "ollama-enabled":
			f.Value.Set(strconv.FormatBool(config.Ollama.Enabled))
		case "ollama-address":
			f.Value.Set(config.Ollama.Address)
		case "ollama-model":
			f.Value.Set(config.Ollama.Model)
		case "ollama-content-length-override":
			f.Value.Set(strconv.Itoa(config.Ollama.ContextLengthOverride))
		case "ollama-update-model":
			f.Value.Set(strconv.FormatBool(config.Ollama.UpdateModel))
		case "openai-enabled":
			f.Value.Set(strconv.FormatBool(config.OpenAI.Enabled))
		case "openai-url":
			f.Value.Set(config.OpenAI.Url)
		case "openai-model":
			f.Value.Set(config.OpenAI.Model)
		case "openai-org-id":
			f.Value.Set(config.OpenAI.OrgId)
		case "openai-api-type":
			f.Value.Set(string(config.OpenAI.ApiType))
		case "openai-api-version":
			f.Value.Set(config.OpenAI.ApiVersion)
		}
	})
	return nil
}
