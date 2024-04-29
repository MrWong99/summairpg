# summairpg - Summarize your role-playing sessions

I was tired of manually writing a summary of the last session of our role-play. So what better to do than use an AI to do it for me?
Do you feel the same? Well look no further because this is the **open source** and **free** way of doing so!
*However since GPT-4 is still the best AI on the market you can also use the OpenAI API...*

**summairpg** is just a mere tool to connect two powerful AI frameworks together, namely **[WhisperX](https://github.com/m-bain/whisperX)** and **[Ollama](https://ollama.com/)**
(or **[OpenAI](https://platform.openai.com/docs/overview)** if you prefer).

summairpg follows a three step process:

1. **Record audio** of the role-playing session. Preferrably provide one audio file per player or game master. If you are using Discord check out [Craig](https://craig.chat) to help you with that!
2. Create a full **transcription** of the role playing session. This project assumes you have setup [WhisperX](https://github.com/m-bain/whisperX) in order to do this.
3. **Summarize** the transcription using a generative AI of your choice.

## Record the voice of all participants

I personally have mostly online role playing sessions using Discord so it's really easy to just add **[Craig](https://craig.chat)**. Afterwards I download the *multitrack FLAC* bundle which includes one .flac file per user. You can choose another format if you want to save up diskspace but the best results will be with a lossless compression.

Once you have the recordings you need to **rename the audio files** so they match the name of the ingame character, e.g. `1-myuser-0.flac` -> `Darell.flac`.
Put all of these files into **one folder** with no other audio files.

## Creating transcriptions

I choose **[WhisperX](https://github.com/m-bain/whisperX)** as the speech-to-text tool as it is very fast (even for audio-tracks of several hours) while being very accurate and also providing timestamps on a per-word basis. It fitted all the needs I had.

I personally setup a conda environment for WhisperX that I can just spin up once needed.

## Summary via AI

For my own summarization I use the `llama3:instruct` model as it responds almost instantly with okay results. (I do however have a `NVIDIA GeForce RTX 3080` at my disposal...)
Still I achieved by far the best results with OpenAIs `gpt-4-turbo`.

## Prerequirements

- **[WhisperX](https://github.com/m-bain/whisperX)** installed and in PATH (run `whisperx --help` to check if it works)
- one of these:
  - **[Ollama](https://ollama.com/)** installed and serving the HTTP API (`ollama serve`)
  - **[OpenAI API Key](https://platform.openai.com/docs/quickstart)** ready to use

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
  -audio-file-types value
        the file extensions that should be considered when looking up audio tracks. They should be a comma-separated list (default flac,wav)
  -audio-language string
        The spoken language in the audio files (default "en")
  -audio-model string
        WhisperX model to use. See https://huggingface.co/models?sort=trending&search=whisper (default "large-v3")
  -audio-transcript-file string
        when set the entire transcription will be skipped and this files content will be used as summarization input
  -config-store
        Store the provided command-line arguments in the summairpg-config.json (default true)
  -ollama-address string
        The host:port of the Ollama HTTP API. (default "127.0.0.1:11434")
  -ollama-enabled
        set to false to disable the Ollama endpoint for summarization (default true)
  -ollama-model string
        Ollama model to use. See https://ollama.com/library (default "llama3:70b")
  -ollama-update-model
        set to false to disable pulling the latest version of the model (default true)
  -open-ai-api-type value
        the type of OpenAI endpoint. Must be one of OPEN_AI, AZURE or AZURE_AD
  -open-ai-api-version string
        the version of the Azure API to use. Not required when openai-api-type is OPEN_AI
  -open-ai-enabled
        set to true to enable the OpenAI endpoint for summarization
  -open-ai-model string
        the OpenAI model to use. See https://platform.openai.com/docs/models/model-endpoint-compatibility (default "gpt-4-turbo")
  -open-ai-org-id string
        will set the OrgID as HTTP header
  -open-ai-url string
        the base url of the OpenAI API endpoint to use (default "https://api.openai.com/v1")
  -openai-api-type value
        the type of OpenAI endpoint. Must be one of OPEN_AI, AZURE or AZURE_AD
  -openai-api-version string
        the version of the Azure API to use. Not required when openai-api-type is OPEN_AI
  -openai-enabled
        set to true to enable the OpenAI endpoint for summarization
  -openai-model string
        the OpenAI model to use. See https://platform.openai.com/docs/models/model-endpoint-compatibility (default "gpt-4-turbo")
  -openai-org-id string
        will set the OrgID as HTTP header
  -openai-url string
        the base url of the OpenAI API endpoint to use (default "https://api.openai.com/v1")
```
