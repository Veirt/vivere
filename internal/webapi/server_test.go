package webapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/veirt/vivere/internal/masterdata"
	"github.com/vmihailenco/msgpack/v5"
)

func TestAPIListsCharacters(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/characters?lang=TextResourceEnUsMB", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("content-type=%q want application/json", contentType)
	}

	var characters []masterdata.Character
	if err := json.NewDecoder(response.Body).Decode(&characters); err != nil {
		t.Fatal(err)
	}
	if len(characters) != 2 {
		t.Fatalf("len(characters)=%d want 2", len(characters))
	}
	character := characters[0]
	if character.ID != 101 || character.Name != "Iris" || character.Title != "The Bloom" {
		t.Fatalf("unexpected character: %#v", character)
	}
	if character.Profile.Description != "A gentle singer." || character.Profile.VoiceJP != "Kana Voice" {
		t.Fatalf("unexpected profile: %#v", character.Profile)
	}
	if character.IconPath != "/assets/Assets/AddressableConvertAssets/CharacterIcon/CHR_000101/CHR_000101_00_s.png" {
		t.Fatalf("unexpected icon path: %q", character.IconPath)
	}
}

func TestAPICharacterJSONMatchesFrontendContract(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/characters?lang=TextResourceEnUsMB", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	var payload []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	first := payload[0]
	if first["iconPath"] == nil || first["profile"] == nil || first["stats"] == nil || first["assets"] == nil || first["skills"] == nil || first["uniqueWeapon"] == nil {
		t.Fatalf("missing frontend contract keys: %#v", first)
	}
	profile := first["profile"].(map[string]any)
	stats := first["stats"].(map[string]any)
	assets := first["assets"].(map[string]any)
	stories := first["stories"].([]any)
	story := stories[0].(map[string]any)
	voices := first["voices"].([]any)
	voice := voices[0].(map[string]any)
	skills := first["skills"].(map[string]any)
	uniqueWeapon := first["uniqueWeapon"].(map[string]any)
	if profile["voiceJp"] != "Kana Voice" || stats["normalSkillId"] == nil || assets["iconSmall"] == nil || assets["elementIcon"] == nil || assets["jobIcon"] == nil || story["rewardItems"] == nil || skills["active"] == nil || uniqueWeapon["bonusParameters"] == nil {
		t.Fatalf("unexpected frontend JSON shape: profile=%#v stats=%#v assets=%#v story=%#v skills=%#v uniqueWeapon=%#v", profile, stats, assets, story, skills, uniqueWeapon)
	}
	if got := story["audioJp"].([]any); len(got) != 2 || got[0] != "/assets/Assets/AddressableConvertAssets/Voice/JP/Character/CHR_000101/Memory/MEM_000001/MV_JP_000101_01_001.wav" {
		t.Fatalf("unexpected story audio paths: %#v", got)
	}
	if voice["audioJp"] != "/assets/Assets/AddressableConvertAssets/Voice/JP/Character/CHR_000101/Common/CV_JP_000101_Mypage_001.wav" {
		t.Fatalf("unexpected voice audio path: %#v", voice)
	}
	second := payload[1]
	for _, key := range []string{"stories", "voices"} {
		if _, ok := second[key].([]any); !ok {
			t.Fatalf("%s must be a JSON array, got %#v", key, second[key])
		}
	}
	secondStats := second["stats"].(map[string]any)
	for _, key := range []string{"activeSkillIds", "passiveSkillIds"} {
		if _, ok := secondStats[key].([]any); !ok {
			t.Fatalf("stats.%s must be a JSON array, got %#v", key, secondStats[key])
		}
	}
	secondProfile := second["profile"].(map[string]any)
	if secondProfile["voiceUs"] != "" || secondProfile["vocalUs"] != "" {
		t.Fatalf("incomplete profile should not expose unresolved synthetic keys: %#v", secondProfile)
	}
}

func TestAPICharactersDefaultsToJapanese(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/characters", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	var characters []masterdata.Character
	if err := json.NewDecoder(response.Body).Decode(&characters); err != nil {
		t.Fatal(err)
	}
	if characters[0].Name != "アイリス" {
		t.Fatalf("default language name=%q want Japanese", characters[0].Name)
	}
}

func TestAPIRejectsInvalidCharacterLanguage(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/characters?lang=../TextResourceEnUsMB", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestAPIGetsSingleCharacter(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/characters/101?lang=TextResourceEnUsMB", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	var character masterdata.Character
	if err := json.NewDecoder(response.Body).Decode(&character); err != nil {
		t.Fatal(err)
	}
	if character.ID != 101 || character.Name != "Iris" {
		t.Fatalf("unexpected character: %#v", character)
	}
	if character.Profile.LamentUS != "US Lament" || len(character.Stories) != 1 || len(character.Voices) != 1 {
		t.Fatalf("missing rich character data: %#v", character)
	}
}

func TestAPIRejectsInvalidCharacterID(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/characters/nope", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestAPIReturnsNotFoundForMissingCharacter(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/characters/999?lang=TextResourceEnUsMB", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusNotFound, response.Body.String())
	}
}

func TestAPIListsMasterBooks(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/master-books", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	var books []masterdata.BookInfo
	if err := json.NewDecoder(response.Body).Decode(&books); err != nil {
		t.Fatal(err)
	}
	if len(books) != 6 {
		t.Fatalf("len(books)=%d want 6: %#v", len(books), books)
	}
	if books[0].Name != "CharacterDetailVoiceMB" || books[1].Name != "CharacterMB" || books[2].Name != "CharacterProfileMB" || books[3].Name != "CharacterStoryMB" || books[4].Name != "TextResourceEnUsMB" || books[5].Name != "TextResourceJaJpMB" {
		t.Fatalf("books not sorted by name: %#v", books)
	}
}

func TestAPILimitsMasterBookRows(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/master-books/CharacterMB?limit=1", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	var rows []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("len(rows)=%d want 1", len(rows))
	}
}

func TestAPIMasterBookRowsDefaultAndZeroLimits(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/master-books/CharacterMB", nil)
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	var rows []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("default len(rows)=%d want all fixture rows", len(rows))
	}

	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/master-books/CharacterMB?limit=0", nil)
	server.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("limit=0 len(rows)=%d want 0", len(rows))
	}
}

func TestAPIRejectsInvalidMasterBookName(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/master-books/%2e%2e", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestAPIRejectsInvalidLimit(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/master-books/CharacterMB?limit=many", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestAPIRejectsLimitOverMaximumBeforeLoadingRows(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/master-books/MissingBookMB?limit=1001", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestServerServesAssets(t *testing.T) {
	server := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/assets/example.txt", nil)
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if strings.TrimSpace(response.Body.String()) != "asset body" {
		t.Fatalf("asset body=%q", response.Body.String())
	}
}

func TestServerRejectsUnsafeOrMissingAssets(t *testing.T) {
	server := newTestServer(t)

	for _, target := range []string{"/assets/missing.txt", "/assets/../go.mod", "/assets/%2e%2e/go.mod", "/assets/nested/..%2f..%2fgo.mod"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, target, nil)
		server.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s status=%d want %d body=%s", target, response.Code, http.StatusNotFound, response.Body.String())
		}
	}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	root := t.TempDir()
	masterRoot := filepath.Join(root, "master")
	exportsRoot := filepath.Join(root, "exports")
	if err := os.MkdirAll(exportsRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(exportsRoot, "example.txt"), []byte("asset body\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeExportFile(t, exportsRoot, "Assets/AddressableConvertAssets/Voice/JP/Character/CHR_000101/Memory/MEM_000001/MV_JP_000101_01_001.wav")
	writeExportFile(t, exportsRoot, "Assets/AddressableConvertAssets/Voice/JP/Character/CHR_000101/Memory/MEM_000001/MV_JP_000101_01_002.wav")
	writeExportFile(t, exportsRoot, "Assets/AddressableConvertAssets/Voice/US/Character/CHR_000101/Memory/MEM_000001/MV_US_000101_01_001.wav")
	writeExportFile(t, exportsRoot, "Assets/AddressableConvertAssets/Voice/JP/Character/CHR_000101/Common/CV_JP_000101_Mypage_001.wav")
	writeExportFile(t, exportsRoot, "Assets/AddressableConvertAssets/Voice/US/Character/CHR_000101/Common/CV_US_000101_Mypage_001.wav")
	writeBook(t, masterRoot, "CharacterMB", []map[string]any{
		{
			"Id":              int64(101),
			"Memo":            "test character",
			"NameKey":         "character.name.101",
			"Name2Key":        "character.title.101",
			"RarityFlags":     int64(4),
			"ElementType":     int64(2),
			"JobFlags":        int64(8),
			"CharacterType":   int64(1),
			"StartTimeFixJST": "2024-01-01",
			"EndTimeFixJST":   "",
		},
		{"Id": int64(202), "NameKey": "unused"},
	})
	writeBook(t, masterRoot, "CharacterProfileMB", []map[string]any{
		{
			"Id":                  int64(101),
			"Birthday":            int64(307),
			"Height":              int64(162),
			"Weight":              int64(45),
			"BloodType":           int64(1),
			"DescriptionKey":      "character.description.101",
			"CharacterVoiceJPKey": "character.voice.101",
			"LamentUSKey":         "character.lament.us.101",
		},
	})
	writeBook(t, masterRoot, "CharacterStoryMB", []map[string]any{
		{"Id": int64(1), "CharacterId": int64(101), "EpisodeId": int64(1), "TitleKey": "story.title.101", "TextKey": "story.text.101"},
	})
	writeBook(t, masterRoot, "CharacterDetailVoiceMB", []map[string]any{
		{"Id": int64(2), "CharacterId": int64(101), "SortOrder": int64(1), "SubtitleKey": "voice.subtitle.101", "UnlockedVoiceButtonTextKey": "voice.button.101"},
	})
	writeBook(t, masterRoot, "TextResourceEnUsMB", []map[string]any{
		{"StringKey": "character.name.101", "Text": "Iris"},
		{"StringKey": "character.title.101", "Text": "The Bloom"},
		{"StringKey": "character.description.101", "Text": "A gentle singer."},
		{"StringKey": "character.voice.101", "Text": "Kana Voice"},
		{"StringKey": "character.lament.us.101", "Text": "US Lament"},
		{"StringKey": "story.title.101", "Text": "First memory"},
		{"StringKey": "story.text.101", "Text": "The story text."},
		{"StringKey": "voice.subtitle.101", "Text": "Hello."},
		{"StringKey": "voice.button.101", "Text": "Greeting"},
	})
	writeBook(t, masterRoot, "TextResourceJaJpMB", []map[string]any{
		{"StringKey": "character.name.101", "Text": "アイリス"},
		{"StringKey": "character.title.101", "Text": "花"},
	})

	server, err := New(Config{MasterRoot: masterRoot, ExportsRoot: exportsRoot})
	if err != nil {
		t.Fatal(err)
	}
	return server
}

func writeExportFile(t *testing.T, root string, name string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("audio"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeBook(t *testing.T, root, name string, rows []map[string]any) {
	t.Helper()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := msgpack.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, name), data, 0o644); err != nil {
		t.Fatal(err)
	}
}
