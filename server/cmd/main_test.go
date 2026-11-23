package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const sampleModel3JSON = `{
  "Version": 3,
  "Name": "model",
  "FileReferences": {
    "Moc": "model.moc3",
    "Textures": ["textures/texture_00.png"],
    "Physics": "model.physics3.json",
    "Motions": {
      "Idle.anim": [{"File": "motions/Idle.anim.motion3.json"}],
      "Face_Smile.anim": [{"File": "motions/Face_Smile.anim.motion3.json"}]
    }
  },
  "Groups": [
    {
      "Target": "Parameter",
      "Name": "EyeBlink",
      "Ids": ["ParamEyeLOpen", "ParamEyeROpen"]
    }
  ]
}`

func seedAssetsDir(t *testing.T) string {
	t.Helper()

	assetsDir := t.TempDir()

	validModelDir := filepath.Join(assetsDir, "Models", "CHR_000085", "model")
	if err := os.MkdirAll(validModelDir, 0o755); err != nil {
		t.Fatalf("create valid model dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(validModelDir, "model.model3.json"), []byte(sampleModel3JSON), 0o644); err != nil {
		t.Fatalf("write valid model3.json: %v", err)
	}

	brokenModelDir := filepath.Join(assetsDir, "Models", "BROKEN", "model")
	if err := os.MkdirAll(brokenModelDir, 0o755); err != nil {
		t.Fatalf("create broken model dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(brokenModelDir, "model.model3.json"), []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write broken model3.json: %v", err)
	}

	audioDir := filepath.Join(assetsDir, "AudioClip")
	if err := os.MkdirAll(audioDir, 0o755); err != nil {
		t.Fatalf("create audio dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(audioDir, "CHR_000085_SONG_SHORT_JP.wav"), []byte("RIFF"), 0o644); err != nil {
		t.Fatalf("write sample audio clip: %v", err)
	}

	return assetsDir
}

func TestCharactersEndpoints(t *testing.T) {
	assetsDir := seedAssetsDir(t)
	app, err := NewApp(assetsDir)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		verify         func(t *testing.T, body []byte)
	}{
		{
			name:           "lists discovered characters",
			path:           "/api/characters",
			expectedStatus: http.StatusOK,
			verify: func(t *testing.T, body []byte) {
				t.Helper()
				var payload CharacterJSON
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatalf("unmarshal characters response: %v", err)
				}
				if len(payload.Characters) != 2 {
					t.Fatalf("expected 2 characters, got %d", len(payload.Characters))
				}
			},
		},
		{
			name:           "returns character detail with motions, groups, and resolved music",
			path:           "/api/characters/CHR_000085",
			expectedStatus: http.StatusOK,
			verify: func(t *testing.T, body []byte) {
				t.Helper()
				var detail CharacterDetail
				if err := json.Unmarshal(body, &detail); err != nil {
					t.Fatalf("unmarshal detail response: %v", err)
				}
				if detail.Id != "CHR_000085" {
					t.Fatalf("expected id CHR_000085, got %q", detail.Id)
				}
				if detail.Music.ShortJPMusicPath != "/api/assets/AudioClip/CHR_000085_SONG_SHORT_JP.wav" {
					t.Fatalf("unexpected ShortJPMusicPath: %q", detail.Music.ShortJPMusicPath)
				}
				if detail.Music.FullJPMusicPath != "" {
					t.Fatalf("expected empty FullJPMusicPath, got %q", detail.Music.FullJPMusicPath)
				}
				if len(detail.ModelDetail.FileReferences.Motions["Idle.anim"]) != 1 {
					t.Fatalf("expected Idle.anim motion entry")
				}
				if len(detail.ModelDetail.Groups) != 1 || len(detail.ModelDetail.Groups[0].Ids) != 2 {
					t.Fatalf("unexpected model groups: %+v", detail.ModelDetail.Groups)
				}
			},
		},
		{
			name:           "returns 404 for unknown character id",
			path:           "/api/characters/UNKNOWN",
			expectedStatus: http.StatusNotFound,
			verify: func(t *testing.T, body []byte) {
				t.Helper()
				var resp Response
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("unmarshal 404 response: %v", err)
				}
				if resp.Success || resp.Message != "Invalid id given" {
					t.Fatalf("unexpected 404 response: %+v", resp)
				}
			},
		},
		{
			name:           "returns 500 when model json is malformed",
			path:           "/api/characters/BROKEN",
			expectedStatus: http.StatusInternalServerError,
			verify: func(t *testing.T, body []byte) {
				t.Helper()
				var resp Response
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("unmarshal 500 response: %v", err)
				}
				if resp.Success || resp.Message != "Error when reading model" {
					t.Fatalf("unexpected 500 response: %+v", resp)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			res, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test(%s): %v", tc.path, err)
			}
			defer func() {
				if closeErr := res.Body.Close(); closeErr != nil {
					t.Fatalf("close body: %v", closeErr)
				}
			}()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}

			if res.StatusCode != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d (%s)", tc.expectedStatus, res.StatusCode, string(body))
			}

			tc.verify(t, body)
		})
	}
}

func TestLoadCharactersMissingAssetsDir(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "does-not-exist")
	characters, err := LoadCharacters(missingDir)
	if err != nil {
		t.Fatalf("expected nil error for missing Models dir, got %v", err)
	}
	if len(characters.Characters) != 0 {
		t.Fatalf("expected 0 characters, got %d", len(characters.Characters))
	}
}
