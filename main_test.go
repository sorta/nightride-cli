package main

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindStationIndex(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "exact name", query: "Darksynth", want: 1},
		{name: "case insensitive", query: "dArKsYnTh", want: 1},
		{name: "station id", query: "nightride", want: 0},
		{name: "trim whitespace", query: "  datawave  ", want: 3},
		{name: "unique prefix", query: "chill", want: 2},
		{name: "exact match wins over longer prefix", query: "rekt", want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findStationIndex(tt.query)
			if err != nil {
				t.Fatalf("findStationIndex(%q) returned error: %v", tt.query, err)
			}
			if got != tt.want {
				t.Fatalf("findStationIndex(%q) = %d, want %d", tt.query, got, tt.want)
			}
		})
	}
}

func TestFindStationIndexErrors(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantError string
	}{
		{name: "empty", query: "  ", wantError: "cannot be empty"},
		{name: "ambiguous prefix", query: "re", wantError: "ambiguous"},
		{name: "unknown", query: "missing", wantError: "not found"},
		{name: "regex characters are literal", query: "[", wantError: "not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := findStationIndex(tt.query); err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("findStationIndex(%q) error = %v, want error containing %q", tt.query, err, tt.wantError)
			}
		})
	}
}

func TestVolumeExponent(t *testing.T) {
	tests := []struct {
		volume float64
		want   float64
	}{
		{volume: 1, want: 0},
		{volume: 0.5, want: -1},
		{volume: 0.25, want: -2},
		{volume: 0, want: 0},
	}

	for _, tt := range tests {
		if got := volumeExponent(tt.volume); math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("volumeExponent(%v) = %v, want %v", tt.volume, got, tt.want)
		}
	}
}

func TestAdjustVolumeClampsAndUnmutes(t *testing.T) {
	m := model{volume: 0.98, muted: true}
	m.adjustVolume(volumeStep)
	if m.volume != 1 || m.muted {
		t.Fatalf("volume increase produced volume=%v muted=%v, want volume=1 muted=false", m.volume, m.muted)
	}

	m.adjustVolume(-2)
	if m.volume != 0 || m.muted {
		t.Fatalf("volume decrease produced volume=%v muted=%v, want volume=0 muted=false", m.volume, m.muted)
	}
}

func TestAppSettingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", settingsFilename)
	want := appSettings{Volume: 0.65, Muted: true, Station: "darksynth"}
	if err := saveAppSettings(path, want); err != nil {
		t.Fatalf("saveAppSettings() error: %v", err)
	}

	got, err := loadAppSettings(path)
	if err != nil {
		t.Fatalf("loadAppSettings() error: %v", err)
	}
	if got != want {
		t.Fatalf("loadAppSettings() = %#v, want %#v", got, want)
	}

	want = appSettings{Volume: 0.25, Station: "datawave"}
	if err := saveAppSettings(path, want); err != nil {
		t.Fatalf("saveAppSettings() overwrite error: %v", err)
	}
	got, err = loadAppSettings(path)
	if err != nil {
		t.Fatalf("loadAppSettings() after overwrite error: %v", err)
	}
	if got != want {
		t.Fatalf("loadAppSettings() after overwrite = %#v, want %#v", got, want)
	}
}

func TestLoadAppSettingsDefaultsAndValidation(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.json")
	got, err := loadAppSettings(missing)
	if err != nil {
		t.Fatalf("loadAppSettings() missing file error: %v", err)
	}
	if got != defaultAppSettings() {
		t.Fatalf("loadAppSettings() missing file = %#v, want defaults %#v", got, defaultAppSettings())
	}

	path := filepath.Join(t.TempDir(), settingsFilename)
	if err := os.WriteFile(path, []byte(`{"volume":2,"station":"rekt"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err = loadAppSettings(path)
	if err != nil {
		t.Fatalf("loadAppSettings() validation error: %v", err)
	}
	if got.Volume != 1 || got.Station != "rekt" {
		t.Fatalf("loadAppSettings() validated settings = %#v", got)
	}

	if err := os.WriteFile(path, []byte(`{"volume":`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadAppSettings(path); err == nil {
		t.Fatal("loadAppSettings() accepted malformed JSON")
	}
}
