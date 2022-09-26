package tts

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/bwmarrin/discordgo"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

func GetVoice(v *discordgo.VoiceConnection, text string) {
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Printf("Failed to create new TTS client: %v", err)
		return
	}
	defer client.Close()

	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Printf("Error communicating with Google TTS API: %v", err)
		return
	}

	fname := "out.mp3"
	err = ioutil.WriteFile(fname, resp.AudioContent, 0644)
	if err != nil {
		log.Printf("Error outputting audio to file: %v", err)
	}
	PlayAudioFile(v, "./"+fname, make(chan bool))
	os.Remove(fname)
}
