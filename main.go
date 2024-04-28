package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/MrWong99/summairpg/pkg/config"
	"github.com/MrWong99/summairpg/pkg/summarize"
	"github.com/MrWong99/summairpg/pkg/transcribe"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		slog.Error("could not initialize config", "error", err)
		os.Exit(1)
	}
	slog.Info("configuration read")
	slog.Info("starting transcription now", "audio-dir", cfg.Audio.Dir, "file-types", cfg.Audio.FileTypes, "language", cfg.Audio.Language, "model", cfg.Audio.Model)
	words, err := transcribe.AsWords(cfg.Audio.Dir, cfg.Audio.Language, cfg.Audio.Model, strings.Split(cfg.Audio.FileTypes, ","))
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
	slog.Info("starting summary now", "model", cfg.Ollama.Model, "address", cfg.Ollama.Address)
	oc := summarize.NewClient(cfg.Ollama.Address, cfg.Ollama.Model)
	if cfg.Ollama.UpdateModel {
		slog.Info("updating Ollama model", "model", cfg.Ollama.Model)
		if err := oc.UpdateModel(); err != nil {
			slog.Error("update failed", "error", err)
			os.Exit(1)
		}
	}
	summary, err := oc.Summarize(lines)
	if err != nil {
		slog.Error("error during summarization", "error", err)
		os.Exit(1)
	}
	slog.Info("summary finished")
	fmt.Println("")
	fmt.Println(summary)
}
