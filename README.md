# summairpg - Summarize your role-playing sessions

I was tired of manually writing a summary of the last session of our role-play. So what better to do than use an AI to do it for me?
Do you feel the same? Well look no further because this is the **open source** and **free** way of doing so!

**summairpg** is just a mere tool to connect two powerful AI frameworks together, namely **[WhisperX](https://github.com/m-bain/whisperX)** and **[Ollama](https://ollama.com/)**.

summairpg follows a three step process:

1. **Record audio** of the role-playing session. Preferrably provide one audio file per player or game master. If you are using Discord check out [Craig](https://craig.chat) to help you with that!
2. Create a full **transcription** of the role playing session. This project assumes you have setup [WhisperX](https://github.com/m-bain/whisperX) in order to do this.
3. **Summarize** the transcription using a generative AI of your choice. We will use [Ollama](https://ollama.com/) as it provides an easy-to-use app.

## Record the voice of all participants

I personally have mostly online role playing sessions using Discord so it's really easy to just add **[Craig](https://craig.chat)**. Afterwards I download the *multitrack FLAC* bundle which includes one .flac file per user. You can choose another format if you want to save up diskspace but the best results will be with a lossless compression.

Once you have the recordings you need to **rename the audio files** so they match the name of the ingame character, e.g. `1-myuser-0.flac` -> `Darell.flac`.
Put all of these files into **one folder** with no other audio files.

## Creating transcriptions

I choose **[WhisperX](https://github.com/m-bain/whisperX)** as the speech-to-text tool as it is very fast (even for audio-tracks of several hours) while being very accurate and also providing timestamps on a per-word basis. It fitted all the needs I had.

I personally setup a conda environment for WhisperX that I can just spin up once needed.

## Summary via AI

To summarize a generative AI of your choice is used. Since this is a very easy task, you don't need the most fancy models and most certainly you don't need to spend money for a cloud AI platform.

For my own summarization I use the `llama3:instruct` model as it responds almost instantly with good results. (I do however have a `NVIDIA GeForce RTX 3080` at my disposal...)

## Prerequirements

- **[WhisperX](https://github.com/m-bain/whisperX)** installed and in PATH (run `whisperx --help` to check if it works)
- **[Ollama](https://ollama.com/)** installed and serving the HTTP API (`ollama serve`)

## Usage

Provide your audio files in a directory and name them after the role playing characters or *GameMaster* for the GM.
Then you can just run the tool:

`./summairpg --audio-dir <dir>`

With the default settings this will create a `summairpg-config.json` file containing your desired configs so if you use the tool again you don't need to provide all parameters again and can just run `./summairpg`.

These are all of the available parameters:

```
$ ./summairpg --help
Usage of ./summairpg:
  -audio-dir string
        The directory that contains all of the audio files that should be transcribed (default "input")
  -audio-display-transcript
        can be set to true to print the entire transcription to console
  -audio-file-types string
        the file extensions that should be considered when looking up audio tracks. They should be a comma-separated list (default "flac,wav")
  -audio-language string
        The spoken language in the audio files (default "en")
  -audio-model string
        WhisperX model to use. See https://huggingface.co/models?sort=trending&search=whisper (default "large-v3")
  -config-store
        Store the provided command-line arguments in the summairpg-config.json (default true)
  -ollama-address string
        The host:port of the Ollama HTTP API. (default "127.0.0.1:11434")
  -ollama-model string
        Ollama model to use. See https://ollama.com/library (default "llama3:70b")
  -ollama-update-model
        set to false to disable pulling the latest version of the model (default true)
```
