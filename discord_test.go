package main

import (
	"testing"
	"time"
)

func TestDiscordActivity(t *testing.T) {
	started := time.Unix(1700000000, 0)
	tests := []struct {
		name, channel, id, title, icon string
	}{
		{"main channel", "Nightride FM", "nightride", "Artist - Song", "nrfm"},
		{"other channel", "Darksynth", "darksynth", "Other Artist - Other Song", "darksynth"},
		{"song with separator", "Chillsynth", "chillsynth", "Artist - Song - Remix", "chillsynth"},
		{"metadata pending", "Nightride FM", "nightride", "", "nrfm"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := station{name: tt.channel, url: "https://stream.nightride.fm/" + tt.id + ".mp3", title: tt.title}
			got := discordActivity(st, started)
			if got.Type != 2 || got.Details != "Listening to "+tt.channel || got.State != tt.title {
				t.Fatalf("unexpected presence: %#v", got)
			}
			if got.LargeImage != tt.icon || got.SmallImage != "nrfm" || got.LargeText != tt.title {
				t.Fatalf("unexpected assets: %#v", got)
			}
			if got.Timestamps == nil || got.Timestamps.Start == nil || !got.Timestamps.Start.Equal(started) {
				t.Fatalf("unexpected timestamp: %#v", got.Timestamps)
			}
			if len(got.Buttons) != 2 || got.Buttons[0].Label != "Listen to "+tt.channel || got.Buttons[0].Url != "https://nightride.fm/?station="+tt.id {
				t.Fatalf("unexpected buttons: %#v", got.Buttons)
			}
		})
	}
}
