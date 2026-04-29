package boi

import (
	"reflect"
	"testing"
)

func TestParseAPKVersion(t *testing.T) {
	got, err := ParseAPKVersion("window.__APK='mementomori_3.20.0.apk';")
	if err != nil {
		t.Fatalf("ParseAPKVersion() error = %v", err)
	}
	if got != "3.20.0" {
		t.Fatalf("ParseAPKVersion() = %q, want %q", got, "3.20.0")
	}
}

func TestParseAPKVersionUTF16StyleText(t *testing.T) {
	got, err := ParseAPKVersion("\xff\xfe/\x00a\x00p\x00p\x00s\x00/\x00m\x00e\x00m\x00e\x00n\x00t\x00o\x00m\x00o\x00r\x00i\x00_\x004\x00.\x001\x003\x00.\x000\x00.\x00a\x00p\x00k\x00")
	if err != nil {
		t.Fatalf("ParseAPKVersion() error = %v", err)
	}
	if got != "4.13.0" {
		t.Fatalf("ParseAPKVersion() = %q, want %q", got, "4.13.0")
	}
}

func TestFormatURI(t *testing.T) {
	got, err := FormatURI("https://cdn.example/{0}/{1}", "123", "master-catalog")
	if err != nil {
		t.Fatalf("FormatURI() error = %v", err)
	}
	if got != "https://cdn.example/123/master-catalog" {
		t.Fatalf("FormatURI() = %q", got)
	}
}

func TestVersionCandidates(t *testing.T) {
	got, err := VersionCandidates("3.20.0")
	if err != nil {
		t.Fatalf("VersionCandidates() error = %v", err)
	}
	wantPrefix := []string{"3.20.0", "3.20.1", "3.20.2", "3.20.3", "3.20.4", "3.20.5", "3.21.0", "3.22.0"}
	if !reflect.DeepEqual(got[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("VersionCandidates() prefix = %v, want %v", got[:len(wantPrefix)], wantPrefix)
	}
	if got[len(got)-1] != "8.0.0" {
		t.Fatalf("VersionCandidates() last = %q, want %q", got[len(got)-1], "8.0.0")
	}
}
