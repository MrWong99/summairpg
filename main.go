package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/MrWong99/summairpg/pkg/config"
	"github.com/MrWong99/summairpg/pkg/summarize"
	"github.com/MrWong99/summairpg/pkg/transcribe"
)

func main() {
	cfg := initConfig()

	lines := evaluateTranscript(cfg)

	if !cfg.Ollama.Enabled && !cfg.OpenAI.Enabled {
		slog.Info("no summary requested")
		return
	}

	evaluateSummary(cfg, lines)
}

func initConfig() *config.App {
	cfg, err := config.Init()
	if err != nil {
		slog.Error("could not initialize config", "error", err)
		os.Exit(1)
	}
	slog.Info("configuration read")
	return cfg
}

func evaluateTranscript(cfg *config.App) []transcribe.Line {
	if cfg.Audio.TranscriptFile != "" {
		slog.Info("transcript will be read via input file", "file", cfg.Audio.TranscriptFile)
		lines, err := transcribe.LinesFromFile(cfg.Audio.TranscriptFile)
		if err != nil {
			slog.Error("transcription file invalid", "error", err)
			os.Exit(1)
		}
		slog.Info("transcription read", "lines", len(lines))
		return lines
	}

	slog.Info("starting transcription now", "audio-dir", cfg.Audio.Dir, "file-types", cfg.Audio.FileTypes, "language", cfg.Audio.Language, "model", cfg.Audio.Model)
	words, err := transcribe.AsWords(cfg.Audio.Dir, cfg.Audio.Language, cfg.Audio.Model, cfg.Audio.FileTypes)
	if err != nil {
		slog.Error("error during transcription", "error", err)
		os.Exit(1)
	}
	lines := transcribe.ToLines(words)
	slog.Info("transcription finished", "words", len(words), "lines", len(lines))

	if cfg.Audio.DisplayTranscript {
		fmt.Println("")
		for _, line := range lines {
			fmt.Println(line.String())
		}
		fmt.Println("")
	}
	return lines
}

func evaluateSummary(cfg *config.App, lines []transcribe.Line) {
	var summary string
	var err error
	switch {
	case cfg.Ollama.Enabled:
		slog.Info("starting summary now", "model", cfg.Ollama.Model, "address", cfg.Ollama.Address)
		oc := summarize.NewOllamaClient(cfg.Ollama.Address, cfg.Ollama.Model)
		if cfg.Ollama.UpdateModel {
			slog.Info("updating Ollama model", "model", cfg.Ollama.Model)
			if err := oc.UpdateModel(); err != nil {
				slog.Error("update failed", "error", err)
				os.Exit(1)
			}
		}
		summary, err = oc.Summarize(lines)
	case cfg.OpenAI.Enabled:
		slog.Info("starting summary now", "model", cfg.OpenAI.Model, "url", cfg.OpenAI.Url)
		oc := summarize.NewOpenAIClient(cfg.OpenAI.Url, cfg.OpenAI.Model, cfg.OpenAI.OrgId, cfg.OpenAI.ApiType, cfg.OpenAI.ApiVersion)
		summary, err = oc.Summarize(lines)
	}

	if err != nil {
		slog.Error("error during summarization", "error", err)
		os.Exit(1)
	}
	slog.Info("summary finished")
	fmt.Println("")
	fmt.Println(summary)
}
