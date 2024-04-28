package transcribe

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Word is a singular transcribed word.
type Word struct {
	// Nickname of the speaker.
	Nickname string
	// Text that was spoken including puctuation.
	Text string
	// StartTime relative to the beginning of the recording in second floating-point precision.
	StartTime float64
}

func (w *Word) String() string {
	return w.Text
}

// Line is a line of spoken text by a singular speaker.
type Line struct {
	Nickname string
	Words    []Word
}

func (l *Line) String() string {
	return fmt.Sprintf("%s: %s", l.Nickname, l.WordsString())
}

// WordsString returns all words joined by a space.
func (l *Line) WordsString() string {
	wordStrings := make([]string, len(l.Words))
	for i, word := range l.Words {
		wordStrings[i] = word.Text
	}
	return strings.Join(wordStrings, " ")
}

type audioFile struct {
	Nickname string
	Filename string
}

// AsWords will transcribe all audio files in the given directory that match the fileExtensions.
//
// The language and model will just be passed as-is to WhisperX.
func AsWords(dir, language, model string, fileExtensions []string) ([]Word, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not open directory %q: %w", dir, err)
	}
	requests := make([]audioFile, 0)
	for _, file := range files {
		// Ignore unwanted files by extension
		ext := filepath.Ext(file.Name())
		if !slices.ContainsFunc(fileExtensions, func(desiredExt string) bool {
			return ext == "."+desiredExt
		}) {
			continue
		}
		requests = append(requests, audioFile{
			Nickname: strings.TrimSuffix(filepath.Base(file.Name()), ext),
			Filename: filepath.Join(dir, file.Name()),
		})
	}

	allWords := make([]Word, 0)
	tmpDir, err := os.MkdirTemp("", "summairpg-*")
	if err != nil {
		return nil, fmt.Errorf("could not create temporary directory: %w", err)
	}
	for _, audioFile := range requests {
		words, err := transcribeWhisperx(transcribeRequest{
			file:     audioFile,
			outDir:   tmpDir,
			language: language,
			model:    model,
		})
		if err != nil {
			return nil, fmt.Errorf("could not transcribe file %q: %w", audioFile.Filename, err)
		}
		allWords = append(allWords, words...)
	}
	slices.SortFunc(allWords, func(a, b Word) int {
		if a.StartTime < b.StartTime {
			return -1
		}
		if a.StartTime > b.StartTime {
			return 1
		}
		return 0
	})
	return allWords, nil
}

type transcribeRequest struct {
	file     audioFile
	outDir   string
	language string
	model    string
}

func transcribeWhisperx(req transcribeRequest) ([]Word, error) {
	abs, err := filepath.Abs(req.file.Filename)
	if err != nil {
		return nil, err
	}
	nickname := strings.TrimSuffix(filepath.Base(req.file.Filename), filepath.Ext(req.file.Filename))
	cmd := exec.Command("whisperx",
		"--model", req.model, "--align_model", "WAV2VEC2_ASR_LARGE_LV60K_960H",
		"--batch_size", "4", "--task", "transcribe", "--output_dir", req.outDir, "--output_format", "json", "--language", req.language, abs)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("could not run %s, output:\n%s", cmd, out))
	}
	outFile := filepath.Join(req.outDir, nickname+".json")
	f, err := os.Open(outFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var res WhisperxResult
	if err := json.NewDecoder(f).Decode(&res); err != nil {
		return nil, err
	}
	words := make([]Word, len(res.WordSegments))
	lastStart := float64(0)
	for i, word := range res.WordSegments {
		w := Word{
			Nickname: nickname,
			Text:     word.Word,
		}
		if word.Start != 0 {
			w.StartTime = word.Start
			lastStart = word.Start
		} else {
			w.StartTime = lastStart
		}
		words[i] = w
	}
	return words, nil
}

type WhisperxResult struct {
	Segments []struct {
		Start float64 `json:"start"`
		End   float64 `json:"end"`
		Text  string  `json:"text"`
		Words []struct {
			Word  string  `json:"word"`
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Score float64 `json:"score"`
		} `json:"words"`
	} `json:"segments"`
	WordSegments []struct {
		Word  string  `json:"word"`
		Start float64 `json:"start"`
		End   float64 `json:"end"`
		Score float64 `json:"score"`
	} `json:"word_segments"`
}

// ToLines converts all of the given words to lines of text.
// A line of text will always be spoken by one speaker, so if speakers switch
// there will be a new line. There will also be a new line if no speaker is speaking for more than 5 seconds.
func ToLines(words []Word) []Line {
	if len(words) == 0 {
		return nil
	}
	lines := make([]Line, 0)
	continousWords := make([]Word, 1)
	lastWord := words[0]
	continousWords[0] = lastWord
	for i, currentWord := range words {
		if i == 0 {
			continue
		}
		if lastWord.Nickname == currentWord.Nickname && (currentWord.StartTime-lastWord.StartTime) < 5 {
			// Word is in streak
			continousWords = append(continousWords, currentWord)
			lastWord = currentWord
			continue
		}
		// current word is not in streak. Form new line
		lines = append(lines, asLine(continousWords))
		continousWords = make([]Word, 1)
		continousWords[0] = currentWord
		lastWord = currentWord
	}
	lines = append(lines, asLine(continousWords))
	return lines
}

func asLine(words []Word) Line {
	return Line{
		Nickname: words[0].Nickname,
		Words:    words,
	}
}
